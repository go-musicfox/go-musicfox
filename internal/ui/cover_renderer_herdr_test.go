package ui

import (
	"bytes"
	"errors"
	"image"
	"image/png"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/ui/kitty"
)

func TestHerdrViewRetriesUncommittedPlacement(t *testing.T) {
	for _, move := range []bool{false, true} {
		name := "song change rejected by budget"
		if move {
			name = "position change write failure"
		}
		t.Run(name, func(t *testing.T) {
			t.Setenv("HERDR_PANE_ID", "1")
			previousTmux := kitty.UseTmuxPassthrough()
			kitty.SetTmuxPassthroughForTest(false)
			t.Cleanup(func() { kitty.SetTmuxPassthroughForTest(previousTmux) })
			app, netease := newFormPageTestApp(t)
			configs.AppConfig.Main.Lyric.Cover.Show = true
			configs.AppConfig.Main.Lyric.Cover.Spin = true
			netease.lyricRenderer = &LyricRenderer{isVisible: true, lyricStartRow: 10, lyricLines: 3}
			var encoded bytes.Buffer
			if err := png.Encode(&encoded, image.NewRGBA(image.Rect(0, 0, 2, 2))); err != nil {
				t.Fatal(err)
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				_, _ = w.Write(encoded.Bytes())
			}))
			t.Cleanup(server.Close)
			song := structs.Song{Id: 1, Album: structs.Album{PicUrl: server.URL + "/a"}}
			r := &CoverRenderer{netease: netease, state: songInfoTestState{song: song},
				imageCache: kitty.NewImageCache(2), kittySupport: true, cols: 8, rows: 4,
				lastBgTransparent: isAppBackgroundTransparent(app)}
			previousWrite := coverStdoutWrite
			t.Cleanup(func() { coverStdoutWrite = previousWrite })
			writes, fail := 0, false
			coverStdoutWrite = func(s string) (int, error) {
				writes++
				if strings.Contains(s, "a=f") || strings.Contains(s, "a=a") {
					t.Error("Herdr View emitted animation")
				}
				if fail {
					return 0, errors.New("write failed")
				}
				return len(s), nil
			}
			r.View(app, app.MustMain())
			if !r.imageRendered || r.currentSongID != 1 || writes != 1 {
				t.Fatal("initial static cover was not displayed")
			}
			oldSeq, oldRow, oldCol := r.cachedSeq, r.lastStartRow, r.lastStartCol
			limiter := r.imageLimiter()
			if move {
				netease.lyricRenderer.lyricStartRow++
				fail = true
			} else {
				song.Id = 2
				song.PicUrl = server.URL + "/b"
				r.state = songInfoTestState{song: song}
				limiter.tokens = 0
				limiter.lastRefill = time.Now().Add(time.Hour)
			}
			r.View(app, app.MustMain())
			if r.currentSongID != 1 || r.cachedSeq != oldSeq || r.lastStartRow != oldRow || r.lastStartCol != oldCol {
				t.Fatal("unsuccessful placement changed displayed state")
			}
			if r.placeBackoff == 0 {
				t.Fatal("unsuccessful placement did not arm backoff")
			}
			beforeRetry := writes
			r.View(app, app.MustMain())
			if writes != beforeRetry {
				t.Fatal("View wrote during backoff")
			}
			// Advance the stored deadlines instead of sleeping. No worker uses this limiter.
			fail = false
			r.placeFailAt = time.Now().Add(-r.placeBackoff - time.Second)
			limiter.tokens = imageBurstBytes
			limiter.cooldownUntil = time.Time{}
			r.View(app, app.MustMain())
			if writes != beforeRetry+1 || r.currentSongID != song.Id || r.placeBackoff != 0 {
				t.Fatal("View did not retry and commit placement")
			}
			if move && r.lastStartRow != oldRow+1 {
				t.Fatal("retry did not commit new position")
			}
			r.View(app, app.MustMain())
			if writes != beforeRetry+1 {
				t.Fatal("successful placement was retransmitted")
			}
		})
	}
}
