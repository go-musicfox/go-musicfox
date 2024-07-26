package tray

import (
	"fmt"
	"os"
	"time"

	"github.com/getlantern/systray"
	"github.com/go-musicfox/go-musicfox/internal/icon"
	"github.com/godbus/dbus/v5"
	//"github.com/skratchdot/open-golang/open"
)

func RunTray() {
	onExit := func() {
		now := time.Now()
		os.WriteFile(fmt.Sprintf(`on_exit_%d.txt`, now.UnixNano()), []byte(now.String()), 0644)
	}

	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetTemplateIcon(icon.IconData, icon.IconData)
	systray.SetTitle("musicfox")

	mPlay := systray.AddMenuItem("播放", "play")
	mStop := systray.AddMenuItem("暂停", "stop")
	mPreviousSong := systray.AddMenuItem("上一曲", "previous song")
	mNextSong := systray.AddMenuItem("下一曲", "next song")
	mLikePlayingSong := systray.AddMenuItem("Love ❤️", "love the playing song")
	mPlayModeMenuItem := systray.AddMenuItem("播放模式", "play mode")
	mQuitMenuItem := systray.AddMenuItem("Quit", "quit the musicfox")

	// children item
	singleRepeatMode := mPlayModeMenuItem.AddSubMenuItem("单曲循环", "single repeat")
	listRepeatMode := mPlayModeMenuItem.AddSubMenuItem("列表循环", "list repeat")
	randomPlayMode := mPlayModeMenuItem.AddSubMenuItem("随即播放", "random play")

	conn, err := dbus.SessionBus()
	if err != nil {
		return
	}

	var bus_name string = fmt.Sprintf("org.mpris.MediaPlayer2.musicfox.instance%d", os.Getpid())
	player := conn.Object(bus_name, dbus.ObjectPath("/org/mpris/MediaPlayer2"))

	callDbus := func(dbus_method string) {
		call := player.Call("org.mpris.MediaPlayer2.Player."+dbus_method, 0)
		if call.Err != nil {
			return
		}
	}

	go func() {
		for {
			select {
			case <-mQuitMenuItem.ClickedCh:
				os.Exit(0)
			case <-mPreviousSong.ClickedCh:
				callDbus("Previous")
			case <-mNextSong.ClickedCh:
				callDbus("Next")
			case <-mPlay.ClickedCh:
				callDbus("Play")
			case <-mStop.ClickedCh:
				callDbus("Stop")

			// 写不来😭
			case <-singleRepeatMode.ClickedCh:
			case <-listRepeatMode.ClickedCh:
			case <-randomPlayMode.ClickedCh:
			case <-mLikePlayingSong.ClickedCh:
			}

		}
	}()

}
