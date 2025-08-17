package composer

import (
	"fmt"

	"github.com/go-musicfox/go-musicfox/internal/structs"
)

const (
	tplShareSong     = "song"
	tplShareAlbum    = "album"
	tplShareArtist   = "artist"
	tplSharePlaylist = "playlist"
	tplShareUser     = "user"
	tplShareDjRadio  = "djRadio"
	tplShareEpisode  = "episode"
)

var defaultShareTemplates = map[string]string{
	tplShareSong:     "分享{{.SongArtists}}的单曲《{{.SongName}}》: {{.SongUrl}} (来自@网易云音乐)",
	tplShareAlbum:    "分享{{.AlbumArtists}}的专辑《{{.AlbumName}}》: {{.AlbumUrl}} (来自@网易云音乐)",
	tplShareArtist:   "分享歌手{{.ArtistName}} {{.ArtistUrl}} (来自@网易云音乐)",
	tplSharePlaylist: "分享{{.UserName}}创建的歌单「{{.PlaylistName}}」: {{.PlaylistUrl}} (来自@网易云音乐)",
	tplShareUser:     "【推荐】来自网易云音乐的{{.UserName}} {{.UserUrl}}",
	tplShareDjRadio:  "分享了#by: {{.UserName}}#的节目《播客：{{.DjRadioName}}》: {{.DjRadioUrl}} (来自@网易云音乐)",
	tplShareEpisode:  "分享了#{{.DjRadioName}}#的节目《{{.EpisodeName}}: {{.EpisodeUrl}}  (来自@网易云音乐)",
}

// ShareService 负责根据不同的数据类型生成标准化的分享文本。
type ShareService struct {
	tplManager *TemplateManager
}

// Sharer 定义了提供可分享项目的行为。
// 任何实现了此接口的类型（通常是 Menu），都表明其内容支持分享功能。
type Sharer interface {
	// ItemToShare 根据指定的索引返回对应的可分享项目。
	// 若索引越界或不存在可分享的项目，则应返回 nil。
	ItemToShare(index int) any
}

// NewShareService 创建并初始化一个 ShareService 实例。
// 此函数会加载所有默认的分享模板。
func NewShareService() *ShareService {
	tplManager := NewTemplateManager()
	if err := tplManager.LoadFromMap(defaultShareTemplates); err != nil {
		panic(fmt.Sprintf("加载默认模板失败: %v", err))
	}
	return &ShareService{
		tplManager: tplManager,
	}
}

// RegisterTemplates 用于动态注册或覆盖分享模板。
func (s *ShareService) RegisterTemplates(templates map[string]string) error {
	return s.tplManager.LoadFromMap(templates)
}

// Share 根据传入的数据项（如 structs.Song）生成对应的分享文本。
// 它会根据数据类型匹配并执行相应的模板。
func (s *ShareService) Share(data any) (string, error) {
	switch v := data.(type) {
	case structs.DjRadio:
		if v.Privacy {
			return "", fmt.Errorf("无法分享私密播客")
		}
		props := NewPropsBuilder().WithDjRadio(v).Build()
		return s.tplManager.Execute(tplShareDjRadio, props)
	case structs.Song:
		if v.UnMatched {
			return "", fmt.Errorf("无法分享未匹配的云盘歌曲")
		}
		if v.DjRadio.Privacy {
			return "", fmt.Errorf("无法分享私密播客中的节目")
		}
		props := NewPropsBuilder().WithSong(v).Build()
		if v.DjRadioEpisodeId == 0 {
			return s.tplManager.Execute(tplShareSong, props) // 普通音乐
		}
		return s.tplManager.Execute(tplShareEpisode, props) // 播客节目
	case structs.Album:
		props := NewPropsBuilder().WithAlbum(v).Build()
		return s.tplManager.Execute(tplShareAlbum, props)
	case structs.Artist:
		props := NewPropsBuilder().WithArtist(v).Build()
		return s.tplManager.Execute(tplShareArtist, props)
	case structs.User:
		props := NewPropsBuilder().WithUser(v).Build()
		return s.tplManager.Execute(tplShareUser, props)
	case structs.Playlist:
		if v.Privacy {
			return "", fmt.Errorf("无法分享私密歌单")
		}
		props := NewPropsBuilder().WithPlaylist(v).Build()
		return s.tplManager.Execute(tplSharePlaylist, props)
	default:
		return "", fmt.Errorf("不支持分享的数据类型: %T", data)
	}
}
