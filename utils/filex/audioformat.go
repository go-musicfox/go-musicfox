package filex

import (
	"bytes"
	"io"
)

const sniffHeaderSize = 4096

// SniffAudioFormat 通过文件内容魔数嗅探真实音频格式，返回小写扩展名
// （"mp3"/"flac"/"ogg"/"wav"），用于纠正音源声明与实际内容不一致的情况
// （如云盘歌曲 API 声明 mp3 但内容实为 FLAC）。
// 仅识别客户端支持的格式：FLAC（fLaC）、OGG（OggS）、WAV（RIFF....WAVE）、
// MP3（ID3 标签或 0xFFE0 掩码帧同步启发式）。其余返回 ok=false，
// 由调用方维持原声明，让解码器按原路径报错兜底。
// 无论结果如何，读取位置都会被重置到文件开头。
func SniffAudioFormat(r io.ReadSeeker) (ext string, ok bool) {
	if _, err := r.Seek(0, io.SeekStart); err != nil {
		return "", false
	}
	defer func() { _, _ = r.Seek(0, io.SeekStart) }()

	data := make([]byte, sniffHeaderSize)
	n, _ := io.ReadFull(r, data)
	data = data[:n]

	switch {
	case bytes.HasPrefix(data, []byte("fLaC")):
		return "flac", true
	case bytes.HasPrefix(data, []byte("OggS")):
		return "ogg", true
	case n >= 12 && bytes.Equal(data[:4], []byte("RIFF")) && bytes.Equal(data[8:12], []byte("WAVE")):
		return "wav", true
	case sniffMP3(data):
		return "mp3", true
	default:
		return "", false
	}
}

// sniffMP3 启发式判断 MP3：以 ID3 标签开头，或前 1KB 内出现合法 MPEG 帧同步头。
func sniffMP3(data []byte) bool {
	if bytes.HasPrefix(data, []byte("ID3")) {
		return true
	}
	limit := len(data)
	if limit > 1024 {
		limit = 1024
	}
	for i := 0; i+4 <= limit; i++ {
		if mp3FrameHeader(data[i:]) {
			return true
		}
	}
	return false
}

// mp3FrameHeader 校验 MPEG 音频帧头：11 位帧同步（0xFFE0 掩码）+
// 合法版本位/层位/比特率索引。
func mp3FrameHeader(b []byte) bool {
	if len(b) < 4 || b[0] != 0xFF || b[1]&0xE0 != 0xE0 {
		return false
	}
	if b[1]&0x18 == 0x08 { // 保留版本
		return false
	}
	if b[1]&0x06 == 0 { // 保留层
		return false
	}
	if b[2]&0xF0 == 0xF0 { // 无效比特率索引
		return false
	}
	return true
}
