package commands

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gookit/gcli/v2"

	"github.com/go-musicfox/go-musicfox/internal/headless"
	"github.com/go-musicfox/go-musicfox/utils/slogx"
)

// ctrlCommands lists the control commands accepted by the ctrl subcommand and
// the --once mode. It is used for usage/error messages.
var ctrlCommands = []string{
	"status", "play <query>", "pause", "resume", "toggle", "stop",
	"next", "prev", "seek <seconds>", "volume [value]", "repeat <off|one|all>",
	"shuffle <on|off>", "like", "dislike", "quit",
}

// NewCtrlCommand builds the `musicfox ctrl <cmd> [args...]` subcommand that
// drives a running headless daemon over its local control channel.
func NewCtrlCommand() *gcli.Command {
	return &gcli.Command{
		Name:   "ctrl",
		UseFor: "control a headless musicfox instance",
		Config: func(c *gcli.Command) {
			// ctrl takes an arbitrary variable number of positional args
			// (<cmd> [args...]); disable the strict argument-count validation.
			c.Arguments.SetValidateNum(false)
		},
		Examples: "{$binName} {$cmd} status\n" +
			"  {$binName} {$cmd} play 周杰伦\n" +
			"  {$binName} {$cmd} volume 60\n" +
			"  {$binName} {$cmd} seek 30\n" +
			"  {$binName} {$cmd} repeat one\n" +
			"  {$binName} {$cmd} shuffle on\n" +
			"  {$binName} {$cmd} quit",
		Func: runCtrl,
	}
}

// runCtrl parses `<cmd> [args...]`, dials the headless daemon, calls the
// command and prints a human-readable result. Failures print to the real
// stderr (slogx.StderrWriter bypasses the global stderr→log redirection) and
// exit with code 1 so scripting can rely on the exit status.
func runCtrl(_ *gcli.Command, args []string) error {
	if len(args) == 0 {
		fmt.Fprintf(slogx.StderrWriter(), "musicfox ctrl: missing command (available: %s)\n", strings.Join(ctrlCommands, " | "))
		os.Exit(1)
	}

	cmd := args[0]
	argMap := buildCtrlArgs(cmd, args[1:])

	client, err := headless.Dial()
	if err != nil {
		fmt.Fprintln(slogx.StderrWriter(), "无头模式未在运行")
		os.Exit(1)
	}
	defer func() { _ = client.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.Call(ctx, cmd, argMap)
	if err != nil {
		fmt.Fprintf(slogx.StderrWriter(), "musicfox ctrl: %v\n", err)
		os.Exit(1)
	}
	if !resp.Ok {
		fmt.Fprintf(slogx.StderrWriter(), "musicfox ctrl: %s\n", resp.Error)
		os.Exit(1)
	}

	printCtrlResult(cmd, resp.Data)
	return nil
}

// buildCtrlArgs maps the raw CLI arguments to the dispatcher's args map. Only
// commands that take arguments get a non-nil map; everything else sends nil.
func buildCtrlArgs(cmd string, rest []string) map[string]any {
	switch cmd {
	case "play":
		if query := strings.TrimSpace(strings.Join(rest, " ")); query != "" {
			return map[string]any{"query": query}
		}
	case "volume":
		if len(rest) == 1 {
			if n, err := strconv.Atoi(rest[0]); err == nil {
				return map[string]any{"value": n}
			}
		}
	case "seek":
		if len(rest) == 1 {
			if f, err := strconv.ParseFloat(rest[0], 64); err == nil {
				return map[string]any{"seconds": f}
			}
		}
	case "repeat":
		if len(rest) == 1 {
			return map[string]any{"mode": rest[0]}
		}
	case "shuffle":
		if len(rest) == 1 {
			on := false
			switch strings.ToLower(rest[0]) {
			case "on", "true", "1":
				on = true
			}
			return map[string]any{"on": on}
		}
	}
	return nil
}

// printCtrlResult renders a command result for humans. The output is kept
// stable and simple so it is also safe for scripting.
func printCtrlResult(cmd string, data any) {
	switch cmd {
	case "status":
		m, ok := data.(map[string]any)
		if !ok {
			fmt.Println("status: (no data)")
			return
		}
		song, _ := m["song"].(map[string]any)
		songName, _ := song["name"].(string)
		artist, _ := song["artist"].(string)
		album, _ := song["album"].(string)
		state, _ := m["state"].(string)
		pos, _ := m["positionSeconds"].(float64)
		dur, _ := m["durationSeconds"].(float64)
		vol, _ := m["volume"].(float64)
		mode, _ := m["mode"].(string)
		playlistLen, _ := m["playlistLen"].(float64)
		user, _ := m["user"].(string)

		fmt.Printf("状态: %s\n", state)
		fmt.Printf("歌曲: %s - %s (%s)\n", songName, artist, album)
		fmt.Printf("进度: %.1fs / %.1fs\n", pos, dur)
		fmt.Printf("音量: %d\n", int(vol))
		fmt.Printf("播放模式: %s\n", mode)
		fmt.Printf("播放列表: %d 首\n", int(playlistLen))
		fmt.Printf("用户: %s\n", user)
	case "play":
		if m, ok := data.(map[string]any); ok {
			id, _ := m["id"].(float64)
			name, _ := m["name"].(string)
			fmt.Printf("正在播放: %s (id=%d)\n", name, int64(id))
		}
	case "volume":
		if m, ok := data.(map[string]any); ok {
			if v, ok := m["volume"].(float64); ok {
				fmt.Printf("音量: %d\n", int(v))
			}
		}
	default:
		if data != nil {
			fmt.Printf("%v\n", data)
		} else {
			fmt.Println("ok")
		}
	}
}
