package ui

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/ui/kitty"
)

func TestPlaceBackoffTransitions(t *testing.T) {
	r := &CoverRenderer{}
	now := time.Now()

	if r.placeBackoffActive(now) {
		t.Fatal("backoff should be inactive initially")
	}

	// First failure: 1s window.
	r.recordPlaceFailure(now)
	if !r.placeBackoffActive(now) || !r.placeBackoffActive(now.Add(999*time.Millisecond)) {
		t.Fatal("expected an active 1s backoff window after the first failure")
	}
	if r.placeBackoffActive(now.Add(time.Second)) {
		t.Fatal("backoff should expire after 1s")
	}

	// Consecutive failures double the window: 2s, then 4s.
	secondFailure := now.Add(time.Second)
	r.recordPlaceFailure(secondFailure)
	if !r.placeBackoffActive(secondFailure.Add(2*time.Second - time.Millisecond)) {
		t.Fatal("expected an active 2s backoff window after the second failure")
	}
	if r.placeBackoffActive(secondFailure.Add(2 * time.Second)) {
		t.Fatal("second backoff window should expire after 2s")
	}

	thirdFailure := secondFailure.Add(2 * time.Second)
	r.recordPlaceFailure(thirdFailure)
	if !r.placeBackoffActive(thirdFailure.Add(4*time.Second - time.Millisecond)) {
		t.Fatal("expected an active 4s backoff window after the third failure")
	}
	if r.placeBackoffActive(thirdFailure.Add(4 * time.Second)) {
		t.Fatal("third backoff window should expire after 4s")
	}

	// Repeated failures cap the window at 30s.
	failureTime := thirdFailure.Add(4 * time.Second)
	for range 10 {
		r.recordPlaceFailure(failureTime)
	}
	if r.placeBackoff != placeBackoffMax {
		t.Fatalf("backoff should be capped at %v, got %v", placeBackoffMax, r.placeBackoff)
	}
	if !r.placeBackoffActive(failureTime.Add(placeBackoffMax - time.Millisecond)) {
		t.Fatal("capped backoff window should still be active just before expiry")
	}
	if r.placeBackoffActive(failureTime.Add(placeBackoffMax)) {
		t.Fatal("capped backoff window should expire")
	}

	// Success resets the backoff entirely.
	r.recordPlaceSuccess()
	if r.placeBackoffActive(time.Now()) || r.placeBackoff != 0 || !r.placeFailAt.IsZero() {
		t.Fatal("success should reset the backoff state")
	}
}

// stripTmuxWrapped removes every DCS tmux; passthrough packet from s,
// unwrapping ESC-doubled payloads with the same rules tmux uses: ESC ESC
// inside the payload is a literal ESC, a lone ESC \ terminates the packet. A
// plain regex is not enough because the doubled inner terminators would
// confuse a non-greedy match.
func stripTmuxWrapped(s string) string {
	const dcs = "\x1bPtmux;"
	var b strings.Builder
	for {
		i := strings.Index(s, dcs)
		if i < 0 {
			b.WriteString(s)
			return b.String()
		}
		b.WriteString(s[:i])
		rest := s[i+len(dcs):]
		j := 0
		for j < len(rest) {
			if rest[j] == 0x1b {
				if j+1 < len(rest) && rest[j+1] == 0x1b {
					j += 2
					continue
				}
				if j+1 < len(rest) && rest[j+1] == '\\' {
					break
				}
			}
			j++
		}
		if j >= len(rest) {
			b.WriteString(rest)
			return b.String()
		}
		s = rest[j+2:]
	}
}

// TestBuildAnimationSequenceTmuxWrapsAllAPCs asserts that in tmux passthrough
// mode every Kitty APC of the assembled animation sequence — including the
// animation control commands SetFrameGap/StartAnimation, not just the
// placement payload — travels inside the DCS passthrough packet: tmux
// consumes bare APCs as pane titles (dropping or polluting them), so the
// animation would silently never start. Tests mutate the kitty package's
// global passthrough mode and must not run in parallel.
func TestBuildAnimationSequenceTmuxWrapsAllAPCs(t *testing.T) {
	kitty.SetTmuxPassthroughForTest(true)
	t.Cleanup(func() { kitty.SetTmuxPassthroughForTest(false) })

	const (
		animID   = 7
		oldAnimD = 5
		row      = 5
		col      = 1
	)
	seq, ok := buildAnimationSequence(animID, oldAnimD, 33, row, col, 20, 1, 0, 74)
	if !ok {
		t.Fatal("expected the placement to succeed")
	}

	if n := strings.Count(seq, "\x1bPtmux;"); n != 1 {
		t.Fatalf("expected exactly one DCS passthrough packet, got %d", n)
	}
	rest := stripTmuxWrapped(seq)
	if strings.Contains(rest, "\x1b_G") {
		t.Fatalf("found a bare Kitty APC outside the passthrough packet: %q", rest)
	}
	// The wrapped payload must carry the outer-terminal absolute CUP
	// (pane top 0 + row 5, pane left 74 + col 1) and the cursor save/restore.
	if !strings.Contains(seq, "\x1b[5;75H") || !strings.Contains(seq, "\x1b7") || !strings.Contains(seq, "\x1b8") {
		t.Fatalf("expected absolute CUP (row 5, col 75) with cursor save/restore in the payload: %q", seq)
	}
}

func TestIsEnabledRequiresTmuxPassthroughOptIn(t *testing.T) {
	prev := configs.AppConfig
	t.Cleanup(func() { configs.AppConfig = prev })

	configs.AppConfig = &configs.Config{}
	configs.AppConfig.Main.Lyric.Cover.Show = true

	kitty.SetTmuxPassthroughForTest(true)
	t.Cleanup(func() { kitty.SetTmuxPassthroughForTest(false) })

	r := &CoverRenderer{kittySupport: true}

	if r.IsEnabled() {
		t.Fatal("expected cover disabled inside tmux without the tmuxPassthrough opt-in")
	}

	configs.AppConfig.Main.Lyric.Cover.TmuxPassthrough = true
	if !r.IsEnabled() {
		t.Fatal("expected cover enabled inside tmux after the tmuxPassthrough opt-in")
	}

	kitty.SetTmuxPassthroughForTest(false)
	configs.AppConfig.Main.Lyric.Cover.TmuxPassthrough = false
	if !r.IsEnabled() {
		t.Fatal("expected cover enabled outside tmux even without tmuxPassthrough")
	}
}

func TestCoverDebugEnabled(t *testing.T) {
	prev := configs.AppConfig
	t.Cleanup(func() { configs.AppConfig = prev })

	configs.AppConfig = nil
	if coverDebugEnabled() {
		t.Fatal("expected debug disabled when AppConfig is nil")
	}

	configs.AppConfig = &configs.Config{}
	if coverDebugEnabled() {
		t.Fatal("expected debug disabled when Main.Debug is false")
	}

	configs.AppConfig.Main.Debug = true
	if !coverDebugEnabled() {
		t.Fatal("expected debug enabled when Main.Debug is true")
	}
}

func TestCoverWriteTraceHelpers(t *testing.T) {
	if coverWriteIsSlow(coverWriteSlowThreshold) {
		t.Fatal("duration equal to the threshold should not be slow")
	}
	if !coverWriteIsSlow(coverWriteSlowThreshold + time.Millisecond) {
		t.Fatal("duration above the threshold should be slow")
	}
	if coverWriteShouldTraceBegin(false, coverWriteTraceMinBytes) {
		t.Fatal("begin/end trace should be off when debug is false")
	}
	if coverWriteShouldTraceBegin(true, coverWriteTraceMinBytes-1) {
		t.Fatal("begin/end trace should skip tiny writes")
	}
	if !coverWriteShouldTraceBegin(true, coverWriteTraceMinBytes) {
		t.Fatal("begin/end trace should run for large writes when debug is on")
	}
}

func TestTmuxImageLimiterTokenBucket(t *testing.T) {
	now := time.Unix(100, 0)
	l := newTmuxImageLimiter(now)

	if decision := l.allow(now, tmuxImageSingleMaxBytes); !decision.allowed {
		t.Fatalf("initial packet should be allowed: %+v", decision)
	}
	if decision := l.allow(now, tmuxImageSingleMaxBytes+1); decision.allowed || decision.reason != "single_packet_limit" {
		t.Fatalf("oversized packet should be rejected: %+v", decision)
	}
	if decision := l.allow(now, tmuxImageSingleMaxBytes); decision.allowed || decision.reason != "rate_limit" {
		t.Fatalf("packet exceeding remaining tokens should be rejected: %+v", decision)
	}
	if decision := l.allow(now.Add(time.Second), tmuxImageSingleMaxBytes); !decision.allowed {
		t.Fatalf("tokens should refill after time advances: %+v", decision)
	}

	state := l.snapshot(now.Add(10 * time.Second))
	if state.tokens != tmuxImageBurstBytes {
		t.Fatalf("tokens should not refill past burst: got %d, want %d", state.tokens, tmuxImageBurstBytes)
	}
}

func TestTmuxImageLimiterCooldown(t *testing.T) {
	now := time.Unix(200, 0)
	l := newTmuxImageLimiter(now)
	slow := coverWriteResult{written: 1, duration: coverWriteSlowThreshold + time.Millisecond}

	want := tmuxSlowCooldownInitial
	for i := 0; i < 5; i++ {
		if pressure := l.report(now, slow, 1); pressure != "slow" {
			t.Fatalf("slow report pressure = %q, want slow", pressure)
		}
		state := l.snapshot(now)
		if state.cooldownRemaining != want {
			t.Fatalf("cooldown %d = %v, want %v", i, state.cooldownRemaining, want)
		}
		if decision := l.allow(now, 1); decision.allowed || decision.reason != "cooldown" {
			t.Fatalf("write during cooldown should be rejected: %+v", decision)
		}
		now = now.Add(want)
		want = min(want*2, tmuxSlowCooldownMax)
	}

	fast := coverWriteResult{written: 1, duration: time.Millisecond}
	if pressure := l.report(now, fast, 1); pressure != "normal" {
		t.Fatalf("fast report pressure = %q, want normal", pressure)
	}
	if state := l.snapshot(now); state.cooldownRemaining != 0 || l.cooldownLevel != 0 {
		t.Fatalf("fast write should reset cooldown: %+v", state)
	}
	l.report(now, slow, 1)
	if state := l.snapshot(now); state.cooldownRemaining != tmuxSlowCooldownInitial {
		t.Fatalf("cooldown after reset = %v, want %v", state.cooldownRemaining, tmuxSlowCooldownInitial)
	}
}

func prepareTmuxWriteTest(t *testing.T) {
	t.Helper()
	origWrite := coverStdoutWrite
	origOffset := coverTmuxPaneOffset
	kitty.SetTmuxPassthroughForTest(true)
	coverTmuxPaneOffset = func() (int, int, bool) { return 2, 3, true }
	t.Cleanup(func() {
		coverStdoutWrite = origWrite
		coverTmuxPaneOffset = origOffset
		kitty.SetTmuxPassthroughForTest(false)
	})
}

func TestWritePositionedTmuxLimiterRejectsWithoutWriting(t *testing.T) {
	prepareTmuxWriteTest(t)
	called := false
	coverStdoutWrite = func(s string) (int, error) {
		called = true
		return len(s), nil
	}

	r := &CoverRenderer{tmuxImageLimiter: newTmuxImageLimiter(time.Now())}
	if r.writePositioned(4, 5, strings.Repeat("x", tmuxImageSingleMaxBytes), true) {
		t.Fatal("oversized wrapped image should be rejected")
	}
	if called {
		t.Fatal("stdout must not be called for a limited image")
	}
	if r.placeBackoff == 0 {
		t.Fatal("limiter rejection should arm placement backoff")
	}
}

func TestWritePositionedTmuxRecordsWrappedBytes(t *testing.T) {
	prepareTmuxWriteTest(t)
	var got string
	coverStdoutWrite = func(s string) (int, error) {
		got = s
		return len(s), nil
	}

	now := time.Now()
	r := &CoverRenderer{tmuxImageLimiter: newTmuxImageLimiter(now)}
	const imageSeq = "\x1b_Gf=100;abc\x1b\\"
	if !r.writePositioned(4, 5, imageSeq, true) {
		t.Fatal("expected tmux image write to succeed")
	}
	want := kitty.Wrap(kitty.BuildTmuxPositionedPayload(2, 3, 4, 5, imageSeq, true))
	if got != want {
		t.Fatalf("stdout payload differs from wrapped positioned payload")
	}
	if state := r.tmuxImageLimiter.snapshot(now); state.admittedBytes != int64(len(want)) {
		t.Fatalf("admitted bytes = %d, want %d", state.admittedBytes, len(want))
	}
}

func TestWritePositionedTmuxWriteFailure(t *testing.T) {
	tests := []struct {
		name  string
		write func(string) (int, error)
	}{
		{name: "short write", write: func(s string) (int, error) { return len(s) - 1, nil }},
		{name: "error", write: func(string) (int, error) { return 0, errors.New("write failed") }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prepareTmuxWriteTest(t)
			coverStdoutWrite = tt.write
			r := &CoverRenderer{tmuxImageLimiter: newTmuxImageLimiter(time.Now())}
			if r.writePositioned(4, 5, "image", false) {
				t.Fatal("failed stdout write must not count as a successful render")
			}
			if state := r.tmuxImageLimiter.snapshot(time.Now()); state.cooldownRemaining <= 0 {
				t.Fatal("failed stdout write should enter congestion cooldown")
			}
		})
	}
}

func TestWritePositionedDirectBypassesLimiter(t *testing.T) {
	kitty.SetTmuxPassthroughForTest(false)
	orig := coverStdoutWrite
	t.Cleanup(func() { coverStdoutWrite = orig })

	called := false
	coverStdoutWrite = func(s string) (int, error) {
		called = true
		return 0, errors.New("best effort direct write")
	}
	r := &CoverRenderer{tmuxImageLimiter: newTmuxImageLimiter(time.Now())}
	if !r.writePositioned(1, 1, strings.Repeat("x", tmuxImageBurstBytes+1), false) {
		t.Fatal("direct writes should preserve best-effort success behavior")
	}
	if !called {
		t.Fatal("direct write should reach stdout regardless of tmux limiter")
	}
	if state := r.tmuxImageLimiter.snapshot(time.Now()); state.admittedBytes != 0 {
		t.Fatalf("direct write unexpectedly consumed limiter tokens: %+v", state)
	}
}

func TestWriteStdoutInvokesStubAndRecordsDuration(t *testing.T) {
	orig := coverStdoutWrite
	t.Cleanup(func() { coverStdoutWrite = orig })

	const payload = "x"
	called := false
	coverStdoutWrite = func(s string) (int, error) {
		called = true
		if s != payload {
			t.Errorf("stub got %q, want %q", s, payload)
		}
		time.Sleep(250 * time.Millisecond)
		return len(s), nil
	}

	r := &CoverRenderer{}
	got := r.writeStdout(payload)
	if !called {
		t.Fatal("expected coverStdoutWrite stub to be invoked")
	}
	if got.written != len(payload) {
		t.Fatalf("written = %d, want %d", got.written, len(payload))
	}
	if !coverWriteIsSlow(got.duration) {
		t.Fatalf("expected write duration to be slow, got %v", got.duration)
	}
}

func TestNewCoverRendererEmptyConfigDoesNotPanic(t *testing.T) {
	prev := configs.AppConfig
	t.Cleanup(func() { configs.AppConfig = prev })

	configs.AppConfig = &configs.Config{}
	r := NewCoverRenderer(nil, nil)
	if r == nil {
		t.Fatal("expected a renderer")
	}
	r.Close()

	configs.AppConfig = nil
	r = NewCoverRenderer(nil, nil)
	if r == nil {
		t.Fatal("expected a renderer with nil AppConfig")
	}
	r.Close()
}

func TestProcessRSSBytesDoesNotPanic(t *testing.T) {
	_ = processRSSBytes()
}

func TestPlaceholderSegmentRequiresTmuxAndImage(t *testing.T) {
	r := &CoverRenderer{
		imageRendered:  true,
		displayImageID: 42,
		cols:           4,
		rows:           3,
		lastStartRow:   10,
		lastStartCol:   5,
	}

	if _, _, ok := r.PlaceholderSegment(10); ok {
		t.Fatal("expected ok=false outside tmux passthrough")
	}

	kitty.SetTmuxPassthroughForTest(true)
	t.Cleanup(func() { kitty.SetTmuxPassthroughForTest(false) })

	startCol, cells, ok := r.PlaceholderSegment(11)
	if !ok {
		t.Fatal("expected placeholder for absRow inside cover rect")
	}
	if startCol != 5 {
		t.Fatalf("startCol = %d, want 5", startCol)
	}
	if got := ansi.StringWidth(cells); got != 4 {
		t.Fatalf("placeholder cells StringWidth = %d, want 4", got)
	}

	if _, _, ok := r.PlaceholderSegment(9); ok {
		t.Fatal("expected ok=false for row above cover")
	}
	if _, _, ok := r.PlaceholderSegment(13); ok {
		t.Fatal("expected ok=false for row below cover")
	}

	r.displayImageID = 0
	if _, _, ok := r.PlaceholderSegment(10); ok {
		t.Fatal("expected ok=false without displayImageID")
	}
}

func TestCoverCloseWithDebugDoesNotHang(t *testing.T) {

	prev := configs.AppConfig
	t.Cleanup(func() { configs.AppConfig = prev })

	configs.AppConfig = &configs.Config{}
	configs.AppConfig.Main.Debug = true
	r := NewCoverRenderer(nil, nil)

	done := make(chan struct{})
	go func() {
		r.Close()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Close hung with debug enabled")
	}
}
