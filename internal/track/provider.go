package track

import (
	"context"

	"github.com/go-musicfox/go-musicfox/internal/structs"
)

// PlayableSourceProvider 解析一首歌的可播放源。
// ResolvePlayableSource 背后抽象化的扩展点：外部音源（本地文件、自定义
// provider 等）可实现该接口并经 WithPlayableSourceProvider 注入替换默认的
// 网易云音源解析（downloaded -> cached -> remote 三源逻辑）。
type PlayableSourceProvider interface {
	ResolvePlayableSource(ctx context.Context, song structs.Song) (PlayableSource, error)
}
