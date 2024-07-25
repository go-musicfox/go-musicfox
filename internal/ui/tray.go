package ui

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/getlantern/systray"
	"github.com/go-musicfox/go-musicfox/internal/icon"
	//"github.com/skratchdot/open-golang/open"
)

func RunTray() {
	onExit := func() {
		now := time.Now()
		ioutil.WriteFile(fmt.Sprintf(`on_exit_%d.txt`, now.UnixNano()), []byte(now.String()), 0644)
	}

	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetTemplateIcon(icon.IconData, icon.IconData)
	systray.SetTitle("musicfox")

	mContinueOrStop := systray.AddMenuItem("播放/暂停", "continue or stop")
	mPreviousSong := systray.AddMenuItem("上一曲", "previous song")
	mNextSong := systray.AddMenuItem("️下一曲", "next song")
	mLikePlayingSong := systray.AddMenuItem("Love ❤️", "love the playing song")
	mPlayModeMenuItem := systray.AddMenuItem("播放模式", "play mode")
	mQuitMenuItem := systray.AddMenuItem("Quit", "quit the musicfox")

	// children item
	singleRepeatMode := mPlayModeMenuItem.AddSubMenuItem("单曲循环", "single repeat")
	listRepeatMode := mPlayModeMenuItem.AddSubMenuItem("列表循环", "list repeat")
	randomPlayMode := mPlayModeMenuItem.AddSubMenuItem("随即播放", "random play")

	go func() {
		for {
			select {
			case <-mQuitMenuItem.ClickedCh:
				os.Exit(0)
			case <-singleRepeatMode.ClickedCh:
				//写不来，先放着
			case <-listRepeatMode.ClickedCh:
				//同上
			case <-randomPlayMode.ClickedCh:
				//同上
			case <-mLikePlayingSong.ClickedCh:
				//同上
			case <-mPreviousSong.ClickedCh:
				//同上
			case <-mNextSong.ClickedCh:
				//同上
			case <-mContinueOrStop.ClickedCh:
				//同上
			}
		}
	}()

}
