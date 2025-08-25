package track

import (
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/utils/netease"
)

// SourceType 定义了可播放源的类型
type SourceType int

const (
	SourceDownloaded SourceType = iota // 来源于最终下载目录
	SourceCached                       // 来源于缓存
	SourceRemote                       // 来源于网络
)

type PlayableSource struct {
	structs.Song
	Type SourceType
	Path string
	Info *netease.PlayableInfo
}
