package automator

import (
	"math/rand"

	"github.com/pkg/errors"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/netease"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
)

type AutoPlayerBackend interface {
	Mode() types.Mode
	SetMode(mode types.Mode)
	Playlist() []structs.Song
	SetPlaylist(playlist []structs.Song)
	CurSongIndex() int
	SetCurSongIndex(index int)
	StartPlay()
}

type AutoPlayer struct {
	backend AutoPlayerBackend
	options configs.AutoPlayerOptions
	user    *structs.User
}

func NewAutoPlayer(user *structs.User, backend AutoPlayerBackend, options configs.AutoPlayerOptions) *AutoPlayer {
	return &AutoPlayer{
		user:    user,
		backend: backend,
		options: options,
	}
}

func (p *AutoPlayer) Start() error {
	if p.user == nil || p.user.UserId == 0 {
		return errors.New("账号未登录")
	}
	var (
		getAll bool
		songs  []structs.Song
		index  int
		mode   = p.options.Mode
		err    error
	)
	if p.options.Offset >= 1000 || p.options.Offset < 0 {
		getAll = true
	}

	if mode == types.PmUnknown {
		mode = p.backend.Mode()
	}
	switch p.options.Playlist {
	case configs.AutoPlayerPlaylistDailyReco:
		songs, err = netease.FetchDailySongs()
	case configs.AutoPlayerPlaylistLike:
		songs, err = netease.FetchLikeSongs(p.user.UserId, getAll)
	case configs.AutoPlayerPlaylistNo:
		fallthrough
	default:
		playlistName := p.options.Playlist.SpecialPlaylist()
		if playlistName != "" {
			// name:xxx
			songs, err = netease.FetchUserPlaylistByName(p.user.UserId, playlistName, getAll)
		} else {
			songs = p.backend.Playlist()
		}
	}
	if err != nil {
		return err
	}

	length := len(songs)
	switch {
	case p.options.Playlist == configs.AutoPlayerPlaylistNo:
		// 保持原来状态
		index = p.backend.CurSongIndex()
	case p.options.Mode == types.PmRandom:
		index = rand.Intn(length)
	default:
		if p.options.Offset >= length || -p.options.Offset > length {
			return errors.Errorf("无效的偏移量：%d", p.options.Offset)
		}
		index = (p.options.Offset + length) % length // 无论offset正负都能工作
	}
	p.backend.SetMode(mode)
	p.backend.SetPlaylist(songs)
	p.backend.SetCurSongIndex(index)
	p.backend.StartPlay()
	return nil
}
