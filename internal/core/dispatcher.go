package core

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-musicfox/netease-music/service"

	"github.com/go-musicfox/go-musicfox/internal/types"
	_struct "github.com/go-musicfox/go-musicfox/utils/struct"
)

// Dispatcher executes control commands against the core player. The embedded
// mutex serializes concurrent dispatch (the transport shares one dispatcher
// across all connections, and playback control must stay single-threaded).
type Dispatcher struct {
	engine *Engine
	mu     sync.Mutex
}

// NewDispatcher builds a dispatcher bound to the given engine.
func NewDispatcher(engine *Engine) *Dispatcher {
	return &Dispatcher{engine: engine}
}

// Dispatch executes cmd with args and returns the command result (or nil). The
// player is always reached through core.Player: Ctrl* channel-backed methods
// for control signals (they exist precisely to serialize and avoid GC-panic
// issues) and direct methods for queries/status.
func (d *Dispatcher) Dispatch(ctx context.Context, cmd string, args map[string]any) (any, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	player := d.engine.Player()
	switch cmd {
	case "status":
		return d.cmdStatus(player)
	case "play":
		return d.cmdPlay(ctx, args)
	case "pause":
		player.CtrlPause()
	case "resume":
		player.CtrlResume()
	case "toggle":
		player.CtrlToggle()
	case "stop":
		player.CtrlStop()
	case "next":
		player.CtrlNext()
	case "prev":
		player.CtrlPrevious()
	case "seek":
		return d.cmdSeek(args)
	case "volume":
		return d.cmdVolume(args)
	case "repeat":
		return d.cmdRepeat(args)
	case "shuffle":
		return d.cmdShuffle(args)
	case "like":
		player.CtrlLikeNowPlaying()
	case "dislike":
		player.CtrlDislikeNowPlaying()
	default:
		return nil, fmt.Errorf("未知命令: %s", cmd)
	}
	return nil, nil
}

// cmdStatus builds the status snapshot for the "status" command.
func (d *Dispatcher) cmdStatus(player *Player) (any, error) {
	info := player.PlayingInfo()

	var userName string
	if u := player.User(); u != nil {
		userName = u.Nickname
	}

	return map[string]any{
		"playing": info.State == types.Playing,
		"state":   stateName(info.State),
		"song": map[string]any{
			"id":     info.TrackID,
			"name":   info.Name,
			"artist": info.Artist,
			"album":  info.Album,
		},
		"positionSeconds": info.PassedDuration.Seconds(),
		"durationSeconds": info.TotalDuration.Seconds(),
		"volume":          info.Volume,
		"mode":            player.Mode().Name(),
		"playlistLen":     len(player.Playlist()),
		"user":            userName,
	}, nil
}

// cmdPlay searches for the first song matching the query and starts playing it
// with a fresh playlist. Search is a public API and does not require login.
func (d *Dispatcher) cmdPlay(ctx context.Context, args map[string]any) (any, error) {
	query, _ := args["query"].(string)
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, errors.New("play 缺少搜索关键词 (usage: play <query>)")
	}

	searchService := service.SearchService{
		S:     query,
		Type:  "1", // single-song search
		Limit: strconv.Itoa(types.SearchPageSize),
	}
	code, resp := searchService.Search()
	if codeType := _struct.CheckCode(code); codeType != _struct.Success {
		return nil, fmt.Errorf("搜索失败: code=%v", code)
	}

	songs := _struct.GetSongsOfSearchResult(resp)
	if len(songs) == 0 {
		return nil, errors.New("未找到歌曲")
	}

	player := d.engine.Player()
	player.ReinitializePlaylist(0, songs)
	player.StartPlay()

	first := songs[0]
	return map[string]any{"id": first.Id, "name": first.Name}, nil
}

// cmdSeek seeks to the given position (args.seconds, float64).
func (d *Dispatcher) cmdSeek(args map[string]any) (any, error) {
	v, ok := args["seconds"]
	if !ok {
		return nil, errors.New("seek 需要 seconds 参数 (usage: seek <seconds>)")
	}
	secs, err := ctrlFloat(v)
	if err != nil {
		return nil, errors.New("seek seconds 必须是数字")
	}
	d.engine.Player().CtrlSeek(time.Duration(secs * float64(time.Second)))
	return nil, nil
}

// cmdVolume gets the current volume when args.value is absent, otherwise sets
// it via CtrlSetVolume.
func (d *Dispatcher) cmdVolume(args map[string]any) (any, error) {
	player := d.engine.Player()
	if v, ok := args["value"]; ok {
		n, err := ctrlInt(v)
		if err != nil {
			return nil, errors.New("volume value 必须是整数 (usage: volume [value])")
		}
		player.CtrlSetVolume(n)
		return map[string]any{"volume": n}, nil
	}
	return map[string]any{"volume": player.Volume()}, nil
}

// cmdRepeat maps args.mode ("off"|"one"|"all") to CtrlSetRepeat(0|1|2),
// mirroring the core setRepeat semantics (0=off/ordered, 1=one/single,
// 2=all/list loop).
func (d *Dispatcher) cmdRepeat(args map[string]any) (any, error) {
	mode, _ := args["mode"].(string)
	switch mode {
	case "off":
		d.engine.Player().CtrlSetRepeat(0)
	case "one":
		d.engine.Player().CtrlSetRepeat(1)
	case "all":
		d.engine.Player().CtrlSetRepeat(2)
	default:
		return nil, fmt.Errorf("repeat mode 必须是 off/one/all, got %q (usage: repeat <off|one|all>)", mode)
	}
	return nil, nil
}

// cmdShuffle maps args.on (bool) to CtrlSetShuffle(1|0).
func (d *Dispatcher) cmdShuffle(args map[string]any) (any, error) {
	on, ok := args["on"].(bool)
	if !ok {
		return nil, errors.New("shuffle 需要 on 布尔参数 (usage: shuffle <on|off>)")
	}
	if on {
		d.engine.Player().CtrlSetShuffle(1)
	} else {
		d.engine.Player().CtrlSetShuffle(0)
	}
	return nil, nil
}

// stateName maps a types.State to a stable lowercase wire string.
func stateName(s types.State) string {
	switch s {
	case types.Playing:
		return "playing"
	case types.Paused:
		return "paused"
	case types.Stopped:
		return "stopped"
	case types.Interrupted:
		return "interrupted"
	default:
		return "unknown"
	}
}

// ctrlFloat converts a JSON-decoded number (float64) or an int to float64.
func ctrlFloat(v any) (float64, error) {
	switch n := v.(type) {
	case float64:
		return n, nil
	case float32:
		return float64(n), nil
	case int:
		return float64(n), nil
	case int64:
		return float64(n), nil
	}
	return 0, errors.New("not a number")
}

// ctrlInt converts a JSON-decoded number or an int to an int.
func ctrlInt(v any) (int, error) {
	switch n := v.(type) {
	case float64:
		return int(n), nil
	case float32:
		return int(n), nil
	case int:
		return n, nil
	case int64:
		return int(n), nil
	}
	return 0, errors.New("not an integer")
}
