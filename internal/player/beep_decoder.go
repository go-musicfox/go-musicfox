package player

import (
	"io"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/flac"
	"github.com/gopxl/beep/minimp3"
	"github.com/gopxl/beep/mp3"
	"github.com/gopxl/beep/vorbis"
	"github.com/gopxl/beep/wav"
	"github.com/pkg/errors"
	minimp3pkg "github.com/tosone/minimp3"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/types"
)

func DecodeSong(t SongType, r io.ReadSeekCloser) (streamer beep.StreamSeekCloser, format beep.Format, err error) {
	return decodeSong(t, r, 0, false)
}

func decodeSong(t SongType, r io.ReadSeekCloser, duration time.Duration, finalized bool) (streamer beep.StreamSeekCloser, format beep.Format, err error) {
	switch t {
	case Mp3:
		gaplessDelay, gaplessPadding := mp3GaplessSamples(r)
		switch configs.AppConfig.Player.Beep.Mp3Decoder {
		case types.BeepMiniMp3Decoder:
			minimp3pkg.BufferSize = 1024 * 50
			streamer, format, err = minimp3.Decode(r)
		default:
			streamer, format, err = mp3.Decode(r)
			if err == nil && configs.AppConfig.Player.Beep.Gapless {
				if finalized && duration > 0 && format.SampleRate > 0 {
					target := int(math.Round(duration.Seconds() * float64(format.SampleRate)))
					decoded := streamer.Len() - gaplessDelay
					if target > 0 && decoded > target {
						gaplessPadding = decoded - target
					}
				}
			}
			if err == nil && configs.AppConfig.Player.Beep.Gapless && (gaplessDelay > 0 || gaplessPadding > 0) {
				streamer = &trimmedMP3{StreamSeekCloser: streamer, start: gaplessDelay, end: gaplessPadding, trimEnd: finalized}
			}
		}
	case Wav:
		streamer, format, err = wav.Decode(r)
	case Ogg:
		streamer, format, err = vorbis.Decode(r)
	case Flac:
		streamer, format, err = flac.Decode(r)
	default:
		err = errors.Errorf("Unknown song type(%d)", t)
	}
	return
}

type trimmedMP3 struct {
	beep.StreamSeekCloser
	start, end int
	trimEnd    bool
	pos        int
	discarded  int
}

func (d *trimmedMP3) Len() int {
	length := d.StreamSeekCloser.Len() - d.start - d.end
	if !d.trimEnd {
		length = d.StreamSeekCloser.Len() - d.start
	}
	if length < 0 {
		return 0
	}
	return length
}

func (d *trimmedMP3) Position() int { return d.pos }

func (d *trimmedMP3) Stream(samples [][2]float64) (n int, ok bool) {
	if d.trimEnd && d.pos >= d.Len() {
		return 0, false
	}
	if d.discarded < d.start {
		remaining := d.start - d.discarded
		discard := remaining
		buf := make([][2]float64, discard)
		got, rawOK := d.StreamSeekCloser.Stream(buf)
		d.discarded += got
		if got < discard {
			return 0, rawOK
		}
	}
	if d.trimEnd {
		remaining := d.Len() - d.pos
		if remaining < len(samples) {
			samples = samples[:remaining]
		}
	}
	n, rawOK := d.StreamSeekCloser.Stream(samples)
	d.pos += n
	if !d.trimEnd {
		return n, rawOK
	}
	if n < len(samples) {
		return n, rawOK
	}
	return n, d.pos < d.Len()
}

func (d *trimmedMP3) Seek(pos int) error {
	if pos < 0 || pos > d.Len() {
		return io.ErrUnexpectedEOF
	}
	if err := d.StreamSeekCloser.Seek(d.start + pos); err != nil {
		return err
	}
	d.pos, d.discarded = pos, d.start
	return nil
}

func mp3GaplessSamples(r io.ReadSeeker) (delay, padding int) {
	if _, err := r.Seek(0, io.SeekStart); err != nil {
		return
	}
	data := make([]byte, 4096)
	n, _ := r.Read(data)
	_, _ = r.Seek(0, io.SeekStart)
	data = data[:n]
	metadataEnd := len(data)
	if metadataEnd > 512 {
		metadataEnd = 512
	}
	for i := 0; i+4 <= metadataEnd; i++ {
		if string(data[i:i+4]) != "LAME" {
			continue
		}
		if i+24 > len(data) {
			break
		}
		value := uint32(data[i+21])<<16 | uint32(data[i+22])<<8 | uint32(data[i+23])
		return int(value >> 12), int(value & 0xfff)
	}
	// FFmpeg-authored MP3s commonly retain Xing/Info but omit a LAME delay
	// field. Their demuxer-visible start skip is the standard 576 encoder plus
	// 529 decoder samples.
	header := string(data[:metadataEnd])
	if strings.Contains(header, "Xing") || strings.Contains(header, "Info") {
		return 1105, 0
	}
	// Some encoders write the same values in the iTunSMPB ID3 text frame.
	if i := strings.Index(string(data), "iTunSMPB"); i >= 0 {
		fields := strings.Fields(string(data[i:]))
		if len(fields) >= 4 {
			d, _ := strconv.ParseInt(fields[2], 16, 32)
			p, _ := strconv.ParseInt(fields[3], 16, 32)
			return int(d), int(p)
		}
	}
	return
}
