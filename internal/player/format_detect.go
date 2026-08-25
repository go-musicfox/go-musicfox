package player

import (
	"io"

	"github.com/go-musicfox/go-musicfox/utils/filex"
)

// sniffFormat 通过文件内容魔数嗅探真实音频格式，用于纠正音源声明与实际内容
// 不一致的情况（如云盘歌曲 API 声明 mp3 但内容实为 FLAC）。
// 嗅探实现与下载命名共用 utils/filex.SniffAudioFormat，此处仅将返回的
// 扩展名转换为 SongType。无论结果如何，读取位置都会被重置到文件开头。
func sniffFormat(r io.ReadSeeker) (t SongType, ok bool) {
	ext, ok := filex.SniffAudioFormat(r)
	if !ok {
		return 0, false
	}
	return SongTypeMapping[ext], true
}
