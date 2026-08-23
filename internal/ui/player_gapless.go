package ui

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/player"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/utils/app"
	"github.com/go-musicfox/go-musicfox/utils/errorx"
	"github.com/go-musicfox/go-musicfox/utils/netease"
	"github.com/go-musicfox/go-musicfox/utils/notify"
)

func (p *Player) maybePreloadGapless(position time.Duration) {
	if !configs.AppConfig.Player.Beep.Gapless {
		return
	}
	gapless, ok := p.Player.(player.GaplessPlayer)
	preloadSeconds := configs.AppConfig.Player.Beep.GaplessPreloadSeconds
	if preloadSeconds <= 0 {
		preloadSeconds = 15
	}
	if !ok || p.CurMusic().Duration-position > time.Duration(preloadSeconds)*time.Second {
		return
	}
	next, ok := p.peekGaplessSong()
	if !ok {
		return
	}
	p.gaplessMu.Lock()
	if p.gaplessLoading || p.gaplessPending == next.Id || p.gaplessTriedFor == p.CurMusic().Id {
		p.gaplessMu.Unlock()
		return
	}
	p.gaplessLoading = true
	fromID := p.CurMusic().Id
	p.gaplessTriedFor = fromID
	p.gaplessMu.Unlock()

	errorx.Go(func() {
		url, musicType, err := p.getPlayInfo(next)
		p.gaplessMu.Lock()
		defer p.gaplessMu.Unlock()
		p.gaplessLoading = false
		if err != nil || url == "" || p.CurMusic().Id != fromID {
			return
		}
		gapless.Preload(player.URLMusic{URL: url, Song: next, Type: player.SongTypeMapping[musicType]})
		p.gaplessPending = next.Id
	}, true)
}

func (p *Player) peekGaplessSong() (structs.Song, bool) {
	songs, index := p.Playlist(), p.CurSongIndex()
	if len(songs) == 0 || index < 0 || index >= len(songs) {
		return structs.Song{}, false
	}
	switch p.Mode() {
	case types.PmOrdered:
		if index+1 >= len(songs) {
			return structs.Song{}, false
		}
		return songs[index+1], true
	case types.PmListLoop:
		return songs[(index+1)%len(songs)], true
	case types.PmSingleLoop:
		return songs[index], true
	default:
		return structs.Song{}, false
	}
}

func (p *Player) commitGaplessTransition(transition player.GaplessTransition) {
	p.gaplessMu.Lock()
	if p.gaplessPending != transition.Music.Id || p.CurMusic().Id != transition.Music.Id {
		p.gaplessMu.Unlock()
		return
	}
	// The stream boundary can race with the normal Stopped notification. In
	// that case the playlist manager has already advanced to this song and the
	// transition event must only finalize gapless bookkeeping, not advance again.
	if current := p.CurSong(); current.Id == transition.Music.Id {
		p.gaplessPending = 0
		p.gaplessLoading = false
		p.gaplessTriedFor = 0
		p.gaplessMu.Unlock()
		return
	}
	p.gaplessPending = 0
	p.gaplessLoading = false
	p.gaplessTriedFor = 0
	p.gaplessMu.Unlock()

	song, err := p.playlistManager.NextSong(false)
	if err != nil || song.Id != transition.Music.Id {
		slog.Warn("gapless playlist transition mismatch", "error", err, "song", transition.Music.Id)
		return
	}
	p.reporter.ReportEnd(transition.PlayedTime)
	p.reporter.ReportStart(song)
	errorx.Go(func() { _ = p.lyricService.SetSong(context.Background(), song) }, true)
	p.LocatePlayingSong()
	p.stateHandler.SetPlayingInfo(p.PlayingInfo())
	p.updateDesktopLyrics()
	p.svc.Rerender(false)
	go notify.Notify(notify.NotifyContent{
		Title:   "正在播放: " + song.Name,
		Text:    fmt.Sprintf("%s - %s", song.ArtistName(), song.Album.Name),
		Icon:    app.AddResizeParamForPicUrl(song.PicUrl, 60),
		Url:     netease.WebUrlOfSong(song.Id),
		GroupId: types.GroupID,
	})
}

func (p *Player) cancelGaplessPreload() {
	if gapless, ok := p.Player.(player.GaplessPlayer); ok {
		gapless.CancelPreload()
	}
	p.gaplessMu.Lock()
	p.gaplessPending = 0
	p.gaplessLoading = false
	p.gaplessTriedFor = 0
	p.gaplessMu.Unlock()
}
