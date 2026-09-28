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
	next, fromID, generation, ok := p.beginGaplessPreload()
	if !ok {
		return
	}

	errorx.Go(func() {
		url, musicType, err := p.getPlayInfo(next)
		if err != nil || url == "" {
			p.gaplessMu.Lock()
			if p.gaplessGeneration == generation {
				p.gaplessLoading = false
			}
			p.gaplessMu.Unlock()
			return
		}
		p.finishGaplessPreload(gapless, player.URLMusic{URL: url, Song: next, Type: player.SongTypeMapping[musicType]}, fromID, generation)
	}, true)
}

func (p *Player) beginGaplessPreload() (structs.Song, int64, uint64, bool) {
	p.gaplessMu.Lock()
	defer p.gaplessMu.Unlock()
	queued, playing := p.playlistManager.InsertedSongsState()
	if queued || playing {
		return structs.Song{}, 0, 0, false
	}
	next, ok := p.peekGaplessSong()
	fromID := p.CurMusic().Id
	if !ok || p.gaplessLoading || p.gaplessPending == next.Id || p.gaplessTriedFor == fromID {
		return structs.Song{}, 0, 0, false
	}
	p.gaplessLoading = true
	p.gaplessTriedFor = fromID
	return next, fromID, p.gaplessGeneration, true
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
	errorx.Go(func() { p.lyricService.SetSong(context.Background(), song) }, true)
	p.LocatePlayingSong()
	p.stateHandler.SetPlayingInfo(p.PlayingInfo())
	p.updateDesktopLyrics()
	p.netease.Rerender(false)
	go notify.Notify(notify.NotifyContent{
		Title:   "正在播放: " + song.Name,
		Text:    fmt.Sprintf("%s - %s", song.ArtistName(), song.Album.Name),
		Icon:    app.AddResizeParamForPicUrl(song.PicUrl, 60),
		Url:     netease.WebUrlOfSong(song.Id),
		GroupId: types.GroupID,
	})
}

func (p *Player) finishGaplessPreload(gapless player.GaplessPlayer, music player.URLMusic, fromID int64, generation uint64) bool {
	p.gaplessMu.Lock()
	defer p.gaplessMu.Unlock()
	if p.gaplessGeneration != generation {
		return false
	}
	p.gaplessLoading = false
	if p.CurMusic().Id != fromID {
		return false
	}
	gapless.Preload(music)
	p.gaplessPending = music.Song.Id
	return true
}

func (p *Player) addSongsToNext(songs []structs.Song) {
	p.gaplessMu.Lock()
	p.gaplessGeneration++
	p.gaplessPending = 0
	p.gaplessLoading = false
	p.gaplessTriedFor = 0
	if gapless, ok := p.Player.(player.GaplessPlayer); ok {
		gapless.CancelPreload()
	}
	p.playlistManager.AddSongsToNext(songs)
	p.gaplessMu.Unlock()
}

func (p *Player) cancelGaplessPreload() {
	p.gaplessMu.Lock()
	p.gaplessGeneration++
	p.gaplessPending = 0
	p.gaplessLoading = false
	p.gaplessTriedFor = 0
	p.gaplessMu.Unlock()
	if gapless, ok := p.Player.(player.GaplessPlayer); ok {
		gapless.CancelPreload()
	}
}
