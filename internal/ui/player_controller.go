package ui

import (
	"log/slog"
	"strconv"
	"time"

	"github.com/buger/jsonparser"
	"github.com/go-musicfox/go-musicfox/internal/macdriver/mediaplayer"
	"github.com/go-musicfox/go-musicfox/internal/remote_control"
	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/utils/likelist"
	"github.com/go-musicfox/go-musicfox/utils/netease"
	"github.com/go-musicfox/go-musicfox/utils/notify"
	"github.com/go-musicfox/netease-music/service"
)

var _ remote_control.Controller = (*Player)(nil)

// Deprecated: Only remote_control.Handler can call this method, others please use Player instead.
func (p *Player) CtrlPause() {
	// NOTICE: 提供给state_handler调用，因为有GC panic问题，这里使用chan传递
	p.ctrl <- CtrlSignal{Type: CtrlPaused}
}

// Deprecated: Only remote_control.Handler can call this method, others please use Player instead.
func (p *Player) CtrlResume() {
	// NOTICE: 提供给state_handler调用，因为有GC panic问题，这里使用chan传递
	p.ctrl <- CtrlSignal{Type: CtrlResume}
}

// Deprecated: Only remote_control.Handler can call this method, others please use Player instead.
func (p *Player) CtrlStop() {
	// NOTICE: 提供给state_handler调用，因为有GC panic问题，这里使用chan传递
	p.ctrl <- CtrlSignal{Type: CtrlStop}
}

// Deprecated: Only remote_control.Handler can call this method, others please use Player instead.
func (p *Player) CtrlToggle() {
	// NOTICE: 提供给state_handler调用，因为有GC panic问题，这里使用chan传递
	p.ctrl <- CtrlSignal{Type: CtrlToggle}
}

// Deprecated: Only remote_control.Handler can call this method, others please use Player instead.
func (p *Player) CtrlNext() {
	// NOTICE: 提供给state_handler调用，因为有GC panic问题，这里使用chan传递
	p.ctrl <- CtrlSignal{Type: CtrlNext}
}

// Deprecated: Only remote_control.Handler can call this method, others please use Player instead.
func (p *Player) CtrlPrevious() {
	// NOTICE: 提供给state_handler调用，因为有GC panic问题，这里使用chan传递
	p.ctrl <- CtrlSignal{Type: CtrlPrevious}
}

// Deprecated: Only remote_control.Handler can call this method, others please use Player instead.
func (p *Player) CtrlSeek(duration time.Duration) {
	// NOTICE: 提供给state_handler调用，因为有GC panic问题，这里使用chan传递
	p.ctrl <- CtrlSignal{Type: CtrlSeek, Duration: duration}
}

// Deprecated: Only remote_control.Handler can call this method, others please use Player instead.
func (p *Player) CtrlSetVolume(volume int) {
	// NOTICE: 提供给state_handler调用，因为有GC panic问题，这里使用chan传递
	p.Player.SetVolume(volume)
}

// Deprecated: Only remote_control.Handler can call this method, others please use Player instead.
func (p *Player) CtrlLikeNowPlaying() {
	p.likeOrDislike(true)
}

// Deprecated: Only remote_control.Handler can call this method, others please use Player instead.
func (p *Player) CtrlDislikeNowPlaying() {
	p.likeOrDislike(false)
}

// likeOrDislike likes or unlikes the current playing song and updates NowPlaying info.
func (p *Player) likeOrDislike(isLike bool) {
	n := p.netease
	user := n.user
	if user == nil || user.UserId == 0 {
		return
	}

	song := p.CurSong()
	if song.Id == 0 {
		return
	}

	go func() {
		// Ensure MyLikePlaylistID is set
		if n.user.MyLikePlaylistID == 0 {
			userPlaylists := service.UserPlaylistService{
				Uid:    strconv.FormatInt(user.UserId, 10),
				Limit:  "1",
				Offset: "0",
			}
			code, response := userPlaylists.UserPlaylist()
			if code == 200 {
				if id, err := jsonparser.GetInt(response, "playlist", "[0]", "id"); err == nil {
					n.user.MyLikePlaylistID = id
				}
			}
		}

		if n.user.MyLikePlaylistID == 0 {
			slog.Error("MyLikePlaylistID is still 0 after fetch")
			return
		}

		op := "add"
		if !isLike {
			op = "del"
		}

		likeService := service.PlaylistTracksService{
			TrackIds: []string{strconv.FormatInt(song.Id, 10)},
			Op:       op,
			Pid:      strconv.FormatInt(n.user.MyLikePlaylistID, 10),
		}

		code, _ := likeService.PlaylistTracks()
		if code != 200 {
			slog.Error("likeSong failed", "code", code, "song", song.Name)
			return
		}

		// Refresh like list and update NowPlaying
		likelist.RefreshLikeList(user.UserId)
		p.stateHandler.SetPlayingInfo(p.PlayingInfo())

		// Send notification
		title := "已添加到我喜欢的歌曲"
		if !isLike {
			title = "已从我喜欢的歌曲移除"
		}
		notify.Notify(notify.NotifyContent{
			Title:   title,
			Text:    song.Name,
			Url:     netease.WebUrlOfPlaylist(n.user.MyLikePlaylistID),
			GroupId: types.GroupID,
		})
	}()
}

// Deprecated: Only remote_control.Handler can call this method, others please use Player instead.
func (p *Player) CtrlShuffle() {
	p.ctrl <- CtrlSignal{Type: CtrlShuffle}
}

// Deprecated: Only remote_control.Handler can call this method, others please use Player instead.
func (p *Player) CtrlRepeat() {
	p.ctrl <- CtrlSignal{Type: CtrlRepeat}
}

// Deprecated: Only remote_control.Handler can call this method, others please use Player instead.
func (p *Player) CtrlSetRepeat(mode mediaplayer.MPRepeatType) {
	p.ctrl <- CtrlSignal{Type: CtrlRepeat, RepeatType: mode}
}

// Deprecated: Only remote_control.Handler can call this method, others please use Player instead.
func (p *Player) CtrlSetShuffle(mode mediaplayer.MPShuffleType) {
	p.ctrl <- CtrlSignal{Type: CtrlShuffle, ShuffleType: mode}
}
