package track

import (
	"io"
	"os"
)

// sniffFileFormat 按文件内容魔数嗅探真实音频格式。
// 返回检测到的扩展名（不含点）与是否命中。未命中时调用方应维持原声明兜底，
// 以免把 AAC/M4A 等 beep 不支持的格式误判为其它格式。
func sniffFileFormat(r io.Reader) (ext string, ok bool) {
	head := make([]byte, 4096)
	n, _ := io.ReadFull(r, head)
	head = head[:n]

	switch {
	case len(head) >= 4 && string(head[:4]) == "fLaC":
		return "flac", true
	case len(head) >= 4 && string(head[:4]) == "OggS":
		return "ogg", true
	case len(head) >= 12 && string(head[:4]) == "RIFF" && string(head[8:12]) == "WAVE":
		return "wav", true
	case len(head) >= 3 && string(head[:3]) == "ID3":
		return "mp3", true
	}

	if isMPEGFrameHeader(head) {
		return "mp3", true
	}
	return "", false
}

// isMPEGFrameHeader 判断 head 中是否存在合法的 MPEG 音频帧同步头。
// MP3 没有固定魔数，需要以 0xFFE0 帧同步 + 字段合法性启发式识别；
// AAC(ADTS) 帧头虽以 0xFFF 开头，但 layer 位为保留值 00，会被此校验拒绝。
func isMPEGFrameHeader(head []byte) bool {
	for i := 0; i+4 <= len(head); i++ {
		b := head[i:]
		if b[0] != 0xFF || b[1]&0xE0 != 0xE0 {
			continue
		}
		// version 位（b[1] bit3-4）与 layer 位（b[1] bit1-2）均不允许保留值 00
		if (b[1]>>3)&0x03 == 0 || (b[1]>>1)&0x03 == 0 {
			continue
		}
		// bitrate 索引（b[2] bit7-4）与采样率索引（b[2] bit3-2）均不允许保留值
		if b[2]>>4 == 0x0F || (b[2]>>2)&0x03 == 0x03 {
			continue
		}
		return true
	}
	return false
}

// sniffFileFormatFromPath 打开文件并嗅探其真实音频格式。
func sniffFileFormatFromPath(path string) (string, bool) {
	f, err := os.Open(path)
	if err != nil {
		return "", false
	}
	defer f.Close()
	return sniffFileFormat(f)
}
