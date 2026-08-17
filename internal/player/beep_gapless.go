package player

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/gopxl/beep"

	"github.com/go-musicfox/go-musicfox/utils/app"
	"github.com/go-musicfox/go-musicfox/utils/iox"
)

type preparedGapless struct {
	fromID int64
	music  URLMusic
	raw    beep.StreamSeekCloser
	stream beep.Streamer
	format beep.Format
	file   *os.File
}

func (p *preparedGapless) close() {
	if p.raw != nil {
		_ = p.raw.Close()
	}
	if p.file != nil {
		name := p.file.Name()
		_ = p.file.Close()
		_ = os.Remove(name)
	}
}

type gaplessState struct {
	mu          sync.Mutex
	generation  atomic.Int64
	preloaded   *preparedGapless
	preloading  int64
	cancelLoad  context.CancelFunc
	transitions chan GaplessTransition
}

func newGaplessState() *gaplessState {
	return &gaplessState{
		transitions: make(chan GaplessTransition, 1),
	}
}

func streamAcrossBoundary(samples [][2]float64, current beep.Streamer, next func() beep.Streamer) (n int, ok, switched bool) {
	n, ok = current.Stream(samples)
	if ok || next == nil {
		return n, ok, false
	}
	nextStreamer := next()
	if nextStreamer == nil {
		return n, ok, false
	}
	// If the current stream ends exactly at the buffer boundary, switch now
	// but leave the next stream untouched for the following callback. Returning
	// ok=true prevents beep.Seq from firing its completion callback between the
	// two tracks.
	if n == len(samples) {
		return n, true, true
	}
	more, nextOK := nextStreamer.Stream(samples[n:])
	deClickBoundary(samples, n, more)
	return n + more, nextOK, true
}

const gaplessDeClickSamples = 64

// MP3 encoders can leave a small PCM discontinuity at an otherwise gapless
// boundary. Fade only the DC step into the next track; this is short enough
// to avoid audible crossfade while removing the sharp click.
func deClickBoundary(samples [][2]float64, boundary, nextCount int) {
	if boundary == 0 || nextCount == 0 {
		return
	}
	width := gaplessDeClickSamples
	if width > nextCount {
		width = nextCount
	}
	last := samples[boundary-1]
	first := samples[boundary]
	correction := [2]float64{last[0] - first[0], last[1] - first[1]}
	for i := 0; i < width; i++ {
		factor := 1.0
		if width > 1 {
			factor = 1 - float64(i)/float64(width-1)
		}
		samples[boundary+i][0] += correction[0] * factor
		samples[boundary+i][1] += correction[1] * factor
	}
}

func (g *gaplessState) cancel() {
	g.generation.Add(1)
	g.mu.Lock()
	if g.cancelLoad != nil {
		g.cancelLoad()
		g.cancelLoad = nil
	}
	if g.preloaded != nil {
		g.preloaded.close()
		g.preloaded = nil
	}
	g.mu.Unlock()
}

func (g *gaplessState) preload(fromID int64, music URLMusic, outputRate beep.SampleRate, client *http.Client, closed <-chan struct{}) {
	id := g.generation.Add(1)
	ctx, cancel := context.WithCancel(context.Background())
	g.mu.Lock()
	if g.cancelLoad != nil {
		g.cancelLoad()
	}
	if g.preloaded != nil {
		g.preloaded.close()
		g.preloaded = nil
	}
	g.preloading = music.Id
	g.cancelLoad = cancel
	g.mu.Unlock()

	go func() {
		prepared := prepareGapless(ctx, fromID, music, outputRate, client, closed)
		g.mu.Lock()
		defer g.mu.Unlock()
		if id != g.generation.Load() || g.preloading != music.Id {
			if prepared != nil {
				prepared.close()
			}
			return
		}
		g.preloading = 0
		g.cancelLoad = nil
		g.preloaded = prepared
	}()
}

func (g *gaplessState) takeIfReady(currentID int64) *preparedGapless {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.preloaded == nil || g.preloaded.fromID != currentID {
		return nil
	}
	prepared := g.preloaded
	g.preloaded = nil
	return prepared
}

func prepareGapless(ctx context.Context, fromID int64, music URLMusic, outputRate beep.SampleRate, client *http.Client, closed <-chan struct{}) *preparedGapless {
	select {
	case <-closed:
		return nil
	default:
	}
	var reader io.ReadCloser
	var err error
	var response *http.Response
	if strings.HasPrefix(music.URL, "file://") {
		reader, err = os.Open(strings.TrimPrefix(music.URL, "file://"))
	} else {
		request, requestErr := http.NewRequestWithContext(ctx, http.MethodGet, music.URL, nil)
		if requestErr != nil {
			return nil
		}
		response, err = client.Do(request)
		if response != nil {
			reader = response.Body
			if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
				slog.Warn("gapless preload http error", "song_id", music.Id, "status", response.StatusCode)
				_ = reader.Close()
				return nil
			}
		}
	}
	if err != nil || reader == nil {
		slog.Warn("gapless preload open error", "song_id", music.Id, "error", err)
		return nil
	}
	defer reader.Close()

	file, err := os.CreateTemp(app.RuntimeDir(), "beep_gapless_*")
	if err != nil {
		return nil
	}
	bytes, copyErr := iox.CopyClose(ctx, file, reader)
	if copyErr != nil {
		slog.Warn("gapless preload download error", "song_id", music.Id, "bytes", bytes, "error", copyErr)
		_ = file.Close()
		_ = os.Remove(file.Name())
		return nil
	}
	if _, err = file.Seek(0, io.SeekStart); err != nil {
		_ = file.Close()
		_ = os.Remove(file.Name())
		return nil
	}
	raw, format, err := decodeSong(music.Type, file, music.Duration, true)
	if err != nil {
		slog.Warn("gapless preload decode error", "song_id", music.Id, "bytes", bytes, "error", err)
		_ = file.Close()
		_ = os.Remove(file.Name())
		return nil
	}
	var stream beep.Streamer = raw
	if format.SampleRate != outputRate {
		stream = beep.Resample(resampleQuiality, format.SampleRate, outputRate, raw)
	}
	slog.Info("gapless preload ready", "song_id", music.Id, "bytes", bytes, "sample_rate", format.SampleRate, "samples", raw.Len())
	return &preparedGapless{fromID: fromID, music: music, raw: raw, stream: stream, format: format, file: file}
}
