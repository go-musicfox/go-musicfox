package composer

import (
	"fmt"
	"text/template"

	"github.com/go-musicfox/go-musicfox/internal/structs"
)

var (
	tpl *template.Template
)

// ItemProps 是一个扁平化的数据结构，用于聚合所有可能在模板中使用的字段。
type ItemProps struct {
	SongId      int64
	SongName    string
	SongArtists string
	SongUrl     string

	AlbumId      int64
	AlbumName    string
	AlbumUrl     string
	AlbumArtists string

	ArtistId   int64
	ArtistName string
	ArtistUrl  string

	PlaylistId   int64
	PlaylistName string
	PlaylistUrl  string

	UserID   int64
	UserName string
	UserUrl  string

	DjRadioId   int64
	DjRadioName string
	DjRadioUrl  string

	EpisodeId   int64
	EpisodeName string
	EpisodeUrl  string
}

// PropsBuilder 以链式调用的方式安全地构建一个完整的 ItemProps 对象。
type PropsBuilder struct {
	props *ItemProps
}

// NewPropsBuilder 创建并返回一个新的 PropsBuilder 实例。
func NewPropsBuilder() *PropsBuilder {
	return &PropsBuilder{
		props: &ItemProps{},
	}
}

// Build 返回最终生成的 ItemProps 结构体。
func (b *PropsBuilder) Build() ItemProps {
	return *b.props
}

// WithSong 从 structs.Song 对象中提取并填充歌曲相关字段。
// 该方法会同时调用 WithAlbum 来填充专辑信息。
// 若该歌曲为播客节目，则还会进一步填充播客电台及节目的信息。
func (b *PropsBuilder) WithSong(s structs.Song) *PropsBuilder {
	b.props.SongId = s.Id
	b.props.SongName = s.Name
	b.props.SongArtists = s.ArtistName()
	b.WithAlbum(s.Album)
	b.props.SongUrl = fmt.Sprintf("https://music.163.com/m/song?id=%d", s.Id)

	if s.DjRadioEpisodeId > 0 {
		b.WithDjRadio(s.DjRadio)
		b.props.EpisodeId = s.DjRadioEpisodeId
		b.props.EpisodeName = s.Name
		b.props.EpisodeUrl = fmt.Sprintf("https://music.163.com/m/program?id=%d", s.DjRadioEpisodeId)
	}
	return b
}

// WithAlbum 从 structs.Album 对象中提取并填充专辑相关字段。
func (b *PropsBuilder) WithAlbum(a structs.Album) *PropsBuilder {
	b.props.AlbumId = a.Id
	b.props.AlbumName = a.Name
	b.props.AlbumUrl = fmt.Sprintf("http://music.163.com/album/%d", a.Id)
	b.props.AlbumArtists = a.ArtistName()
	return b
}

// WithArtist 从 structs.Artist 对象中提取并填充歌手相关字段。
func (b *PropsBuilder) WithArtist(a structs.Artist) *PropsBuilder {
	b.props.ArtistId = a.Id
	b.props.ArtistName = a.Name
	b.props.ArtistUrl = fmt.Sprintf("http://music.163.com/artist?id=%d", a.Id)
	return b
}

// WithPlaylist 从 structs.Playlist 对象中提取并填充歌单相关字段。
// 该方法会同时调用 WithUser 来填充歌单创建者的用户信息。
func (b *PropsBuilder) WithPlaylist(p structs.Playlist) *PropsBuilder {
	b.props.PlaylistId = p.Id
	b.props.PlaylistName = p.Name
	b.WithUser(p.Creator)
	b.props.PlaylistUrl = fmt.Sprintf("https://music.163.com/m/playlist?id=%d", p.Id)
	return b
}

// WithUser 从 structs.User 对象中提取并填充用户相关字段。
func (b *PropsBuilder) WithUser(p structs.User) *PropsBuilder {
	b.props.UserID = p.UserId
	b.props.UserName = p.Nickname
	b.props.UserUrl = fmt.Sprintf("https://music.163.com/m/user?id=%d", p.UserId)
	return b
}

// WithDjRadio 从 structs.DjRadio 对象中提取并填充播客电台相关字段。
// 该方法会同时调用 WithUser 来填充播客创建者的用户信息。
func (b *PropsBuilder) WithDjRadio(a structs.DjRadio) *PropsBuilder {
	b.props.DjRadioId = a.Id
	b.props.DjRadioName = a.Name
	b.props.DjRadioUrl = fmt.Sprintf("http://music.163.com/radio?id=%d", a.Id)
	b.WithUser(a.Dj)
	return b
}
