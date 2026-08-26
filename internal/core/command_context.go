package core

import (
	"github.com/go-musicfox/go-musicfox/internal/frontend"
	"github.com/go-musicfox/go-musicfox/internal/types"
)

// CommandContext produces a UI-agnostic snapshot of the current player state,
// shared by the TUI CommandMenu and the WebUI command endpoint (replacing
// per-frontend duplicated implementations). It is defensive by design: a nil
// receiver or an uninitialized player yields the zero-value snapshot.
func (p *Player) CommandContext() frontend.CommandContext {
	ctx := frontend.CommandContext{}
	if p == nil || p.Player == nil {
		return ctx
	}

	ctx.Playing = p.State() == types.Playing

	if user := p.User(); user != nil {
		ctx.UserID = user.UserId
		ctx.UserName = user.Nickname
	}

	playlist := p.Playlist()
	if index := p.CurSongIndex(); index >= 0 && index < len(playlist) {
		song := playlist[index]
		ctx.Song = &frontend.SongInfo{
			ID:     song.Id,
			Name:   song.Name,
			Artist: song.ArtistName(),
			Album:  song.Album.Name,
		}
	}

	return ctx
}
