// Copyright 2017 Hajime Hoshi
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package maindata

import (
	"fmt"
	"io"

	"github.com/hajimehoshi/go-mp3/internal/bits"
	"github.com/hajimehoshi/go-mp3/internal/consts"
	"github.com/hajimehoshi/go-mp3/internal/frameheader"
	"github.com/hajimehoshi/go-mp3/internal/sideinfo"
)

type FullReader interface {
	ReadFull([]byte) (int, error)
}

// A MainData is MPEG1 Layer 3 Main Data.
type MainData struct {
	ScalefacL [2][2][22]int      // 0-4 bits
	ScalefacS [2][2][13][3]int   // 0-4 bits
	Is        [2][2][576]float32 // Huffman coded freq. lines
}

var scalefacSizesMpeg1 = [16][2]int{
	{0, 0}, {0, 1}, {0, 2}, {0, 3}, {3, 0}, {1, 1}, {1, 2}, {1, 3},
	{2, 1}, {2, 2}, {2, 3}, {3, 1}, {3, 2}, {3, 3}, {4, 2}, {4, 3},
}

var scalefacSizesMpeg2 = [3][6][4]int{
	{{6, 5, 5, 5}, {6, 5, 7, 3}, {11, 10, 0, 0},
		{7, 7, 7, 0}, {6, 6, 6, 3}, {8, 8, 5, 0}},
	{{9, 9, 9, 9}, {9, 9, 12, 6}, {18, 18, 0, 0},
		{12, 12, 12, 0}, {12, 9, 9, 6}, {15, 12, 9, 0}},
	{{6, 9, 9, 9}, {6, 9, 12, 6}, {15, 18, 0, 0},
		{6, 15, 12, 0}, {6, 12, 9, 6}, {6, 18, 9, 0}}}

var nSlen2 = initSlen() /* MPEG 2.0 slen for 'normal' mode */

func initSlen() (nSlen2 [512]int) {
	for i := 0; i < 4; i++ {
		for j := 0; j < 3; j++ {
			n := j + i*3
			nSlen2[n+500] = i | (j << 3) | (2 << 12) | (1 << 15)
		}
	}

	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			for k := 0; k < 4; k++ {
				for l := 0; l < 4; l++ {
					n := l + k*4 + j*16 + i*80
					nSlen2[n] = i | (j << 3) | (k << 6) | (l << 9) | (0 << 12)
				}
			}
		}
	}
	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			for k := 0; k < 4; k++ {
				n := k + j*4 + i*20
				nSlen2[n+400] = i | (j << 3) | (k << 6) | (1 << 12)
			}
		}
	}
	return
}

func Read(source FullReader, prev *bits.Bits, header frameheader.FrameHeader, sideInfo *sideinfo.SideInfo) (*MainData, *bits.Bits, error) {
	nch := header.NumberOfChannels()
	// Calculate header audio data size
	framesize, err := header.FrameSize()
	if err != nil {
		return nil, nil, err
	}
	if framesize > 2000 {
		return nil, nil, fmt.Errorf("mp3: framesize = %d", framesize)
	}
	sideinfo_size := header.SideInfoSize()

	// Main data size is the rest of the frame,including ancillary data
	main_data_size := framesize - sideinfo_size - 4 // sync+header
	// CRC is 2 bytes
	if header.ProtectionBit() == 0 {
		main_data_size -= 2
	}
	// Assemble main data buffer with data from this frame and the previous
	// two frames. main_data_begin indicates how many bytes from previous
	// frames that should be used. This buffer is later accessed by the
	// Bits function in the same way as the side info is.
	m, err := read(source, prev, main_data_size, sideInfo.MainDataBegin)
	if err != nil {
		// This could be due to not enough data in reservoir
		return nil, nil, err
	}

	if header.LowSamplingFrequency() == 1 {
		return getScaleFactorsMpeg2(m, header, sideInfo)
	}
	return getScaleFactorsMpeg1(nch, m, header, sideInfo)
}

func getScaleFactorsMpeg2(m *bits.Bits, header frameheader.FrameHeader, sideInfo *sideinfo.SideInfo) (*MainData, *bits.Bits, error) {

	nch := header.NumberOfChannels()

	md := &MainData{}

	for ch := 0; ch < nch; ch++ {
		part_2_start := m.BitPos()
		numbits := 0
		slen := nSlen2[sideInfo.ScalefacCompress[0][ch]]
		sideInfo.Preflag[0][ch] = (slen >> 15) & 0x1

		n := 0
		if sideInfo.BlockType[0][ch] == 2 {
			n++
			if sideInfo.MixedBlockFlag[0][ch] != 0 {
				n++
			}
		}

		var scaleFactors []int
		d := (slen >> 12) & 0x7

		for i := 0; i < 4; i++ {
			num := slen & 0x7
			slen >>= 3
			if num > 0 {
				for j := 0; j < scalefacSizesMpeg2[n][d][i]; j++ {
					scaleFactors = append(scaleFactors, m.Bits(num))
				}
				numbits += scalefacSizesMpeg2[n][d][i] * num
			} else {
				for j := 0; j < scalefacSizesMpeg2[n][d][i]; j++ {
					scaleFactors = append(scaleFactors, 0)
				}
			}
		}

		n = (n << 1) + 1
		for i := 0; i < n; i++ {
			scaleFactors = append(scaleFactors, 0)
		}

		if len(scaleFactors) == 22 {
			for i := 0; i < 22; i++ {
				md.ScalefacL[0][ch][i] = scaleFactors[i]
			}
		} else {
			for x := 0; x < 13; x++ {
				for i := 0; i < 3; i++ {
					md.ScalefacS[0][ch][x][i] = scaleFactors[(x*3)+i]
				}
			}
		}

		// Read Huffman coded data. Skip stuffing bits.
		if err := readHuffman(m, header, sideInfo, md, part_2_start, 0, ch); err != nil {
			return nil, nil, err
		}
	}
	// The ancillary data is stored here,but we ignore it.
	return md, m, nil
}

func getScaleFactorsMpeg1(nch int, m *bits.Bits, header frameheader.FrameHeader, sideInfo *sideinfo.SideInfo) (*MainData, *bits.Bits, error) {
	md := &MainData{}
	for gr := 0; gr < 2; gr++ {
		for ch := 0; ch < nch; ch++ {
			part_2_start := m.BitPos()
			// Number of bits in the bitstream for the bands
			slen1 := scalefacSizesMpeg1[sideInfo.ScalefacCompress[gr][ch]][0]
			slen2 := scalefacSizesMpeg1[sideInfo.ScalefacCompress[gr][ch]][1]
			if sideInfo.WinSwitchFlag[gr][ch] == 1 && sideInfo.BlockType[gr][ch] == 2 {
				if sideInfo.MixedBlockFlag[gr][ch] != 0 {
					for sfb := 0; sfb < 8; sfb++ {
						md.ScalefacL[gr][ch][sfb] = m.Bits(slen1)
					}
					for sfb := 3; sfb < 12; sfb++ {
						//slen1 for band 3-5,slen2 for 6-11
						nbits := slen2
						if sfb < 6 {
							nbits = slen1
						}
						for win := 0; win < 3; win++ {
							md.ScalefacS[gr][ch][sfb][win] = m.Bits(nbits)
						}
					}
				} else {
					for sfb := 0; sfb < 12; sfb++ {
						//slen1 for band 3-5,slen2 for 6-11
						nbits := slen2
						if sfb < 6 {
							nbits = slen1
						}
						for win := 0; win < 3; win++ {
							md.ScalefacS[gr][ch][sfb][win] = m.Bits(nbits)
						}
					}
				}
			} else {
				// Scale factor bands 0-5
				if sideInfo.Scfsi[ch][0] == 0 || gr == 0 {
					for sfb := 0; sfb < 6; sfb++ {
						md.ScalefacL[gr][ch][sfb] = m.Bits(slen1)
					}
				} else if sideInfo.Scfsi[ch][0] == 1 && gr == 1 {
					// Copy scalefactors from granule 0 to granule 1
					// TODO: This is not listed on the spec.
					for sfb := 0; sfb < 6; sfb++ {
						md.ScalefacL[1][ch][sfb] = md.ScalefacL[0][ch][sfb]
					}
				}
				// Scale factor bands 6-10
				if sideInfo.Scfsi[ch][1] == 0 || gr == 0 {
					for sfb := 6; sfb < 11; sfb++ {
						md.ScalefacL[gr][ch][sfb] = m.Bits(slen1)
					}
				} else if sideInfo.Scfsi[ch][1] == 1 && gr == 1 {
					// Copy scalefactors from granule 0 to granule 1
					for sfb := 6; sfb < 11; sfb++ {
						md.ScalefacL[1][ch][sfb] = md.ScalefacL[0][ch][sfb]
					}
				}
				// Scale factor bands 11-15
				if sideInfo.Scfsi[ch][2] == 0 || gr == 0 {
					for sfb := 11; sfb < 16; sfb++ {
						md.ScalefacL[gr][ch][sfb] = m.Bits(slen2)
					}
				} else if sideInfo.Scfsi[ch][2] == 1 && gr == 1 {
					// Copy scalefactors from granule 0 to granule 1
					for sfb := 11; sfb < 16; sfb++ {
						md.ScalefacL[1][ch][sfb] = md.ScalefacL[0][ch][sfb]
					}
				}
				// Scale factor bands 16-20
				if sideInfo.Scfsi[ch][3] == 0 || gr == 0 {
					for sfb := 16; sfb < 21; sfb++ {
						md.ScalefacL[gr][ch][sfb] = m.Bits(slen2)
					}
				} else if sideInfo.Scfsi[ch][3] == 1 && gr == 1 {
					// Copy scalefactors from granule 0 to granule 1
					for sfb := 16; sfb < 21; sfb++ {
						md.ScalefacL[1][ch][sfb] = md.ScalefacL[0][ch][sfb]
					}
				}
			}
			// Read Huffman coded data. Skip stuffing bits.
			if err := readHuffman(m, header, sideInfo, md, part_2_start, gr, ch); err != nil {
				return nil, nil, err
			}
		}
	}
	// The ancillary data is stored here,but we ignore it.
	return md, m, nil
}

func read(source FullReader, prev *bits.Bits, size int, offset int) (*bits.Bits, error) {
	if size > 1500 {
		return nil, fmt.Errorf("mp3: size = %d", size)
	}
	// Check that there's data available from previous frames if needed
	if prev != nil && offset > prev.LenInBytes() {
		// No, there is not, so we skip decoding this frame, but we have to
		// read the main_data bits from the bitstream in case they are needed
		// for decoding the next frame.
		buf := make([]byte, size)
		if n, err := source.ReadFull(buf); n < size {
			if err == io.EOF {
				return nil, &consts.UnexpectedEOF{"maindata.Read (1)"}
			}
			return nil, err
		}
		// TODO: Define a special error and enable to continue the next frame.
		return bits.Append(prev, buf), nil
	}
	// Copy data from previous frames
	vec := []byte{}
	if prev != nil {
		vec = prev.Tail(offset)
	}
	// Read the main_data from file
	buf := make([]byte, size)
	if n, err := source.ReadFull(buf); n < size {
		if err == io.EOF {
			return nil, &consts.UnexpectedEOF{"maindata.Read (2)"}
		}
		return nil, err
	}
	return bits.New(append(vec, buf...)), nil
}
