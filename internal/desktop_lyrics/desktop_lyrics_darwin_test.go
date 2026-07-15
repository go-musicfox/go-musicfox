//go:build darwin

package desktop_lyrics

import (
	"math"
	"testing"
	"time"
)

func TestUTF16Length(t *testing.T) {
	tests := []struct {
		text string
		want int
	}{
		{"ASCII", 5},
		{"歌词", 2},
		{"a😀b", 4},
		{"𠀀", 2},
	}

	for _, tt := range tests {
		if got := utf16Length(tt.text); got != tt.want {
			t.Errorf("utf16Length(%q) = %d, want %d", tt.text, got, tt.want)
		}
	}
}

func TestAdvanceScroll(t *testing.T) {
	state := &scrollState{
		active:     true,
		maxOffset:  10,
		pauseTimer: 0.1,
	}

	advanceScroll(state, 0.05)
	assertScrollState(t, state, 0, 0.05)

	advanceScroll(state, 0.05)
	assertScrollState(t, state, 0, 0)

	advanceScroll(state, 0.1)
	assertScrollState(t, state, 5, 0)

	advanceScroll(state, 0.1)
	assertScrollState(t, state, 10, scrollEndPause)

	advanceScroll(state, scrollEndPause)
	assertScrollState(t, state, 10, 0)

	advanceScroll(state, scrollTickInterval)
	assertScrollState(t, state, 0, scrollInitialDelay)
}

func TestBeginScrollPreservesUnchangedLine(t *testing.T) {
	lastTick := time.Now()
	state := &scrollState{
		active:     true,
		text:       "a long lyric",
		offset:     24,
		maxOffset:  80,
		pauseTimer: 0.2,
		lastTick:   lastTick,
	}

	if beginScroll(state, "a long lyric", 80) {
		t.Fatal("beginScroll reset an unchanged lyric")
	}
	if state.offset != 24 || state.pauseTimer != 0.2 || !state.lastTick.Equal(lastTick) {
		t.Fatal("unchanged lyric did not preserve scroll progress")
	}

	if !beginScroll(state, "new lyric", 100) {
		t.Fatal("beginScroll did not reset a changed lyric")
	}
	if state.offset != 0 || state.pauseTimer != scrollInitialDelay || !state.lastTick.IsZero() {
		t.Fatal("changed lyric did not reset scroll state")
	}
}

func TestNextWordAnimationDelay(t *testing.T) {
	line := LyricLine{Words: []LyricWord{
		{Word: "first", StartTime: 100, EndTime: 260},
		{Word: "second", StartTime: 1000, EndTime: 1160},
	}}
	tests := []struct {
		timeMs     int64
		wantDelay  float64
		wantActive bool
	}{
		{timeMs: 0, wantDelay: 0.1, wantActive: true},
		{timeMs: 100, wantDelay: scrollTickInterval, wantActive: true},
		{timeMs: 259, wantDelay: scrollTickInterval, wantActive: true},
		{timeMs: 260, wantDelay: 0.74, wantActive: true},
		{timeMs: 1160, wantDelay: 0, wantActive: false},
	}

	for _, tt := range tests {
		gotDelay, gotActive := nextWordAnimationDelay(line, tt.timeMs)
		if gotActive != tt.wantActive || math.Abs(gotDelay-tt.wantDelay) > 1e-9 {
			t.Errorf("nextWordAnimationDelay(%d) = (%v, %t), want (%v, %t)", tt.timeMs, gotDelay, gotActive, tt.wantDelay, tt.wantActive)
		}
	}
}

func TestWordHighlightProgress(t *testing.T) {
	word := LyricWord{StartTime: 1000, EndTime: 1800}
	tests := []struct {
		timeMs int64
		want   float64
	}{
		{timeMs: 1000, want: 0},
		{timeMs: 1200, want: 0.25},
		{timeMs: 1400, want: 0.5},
		{timeMs: 1600, want: 0.75},
		{timeMs: 1800, want: 1},
	}

	for _, tt := range tests {
		if got := wordHighlightProgress(word, tt.timeMs); math.Abs(got-tt.want) > 1e-9 {
			t.Errorf("wordHighlightProgress(%d) = %v, want %v", tt.timeMs, got, tt.want)
		}
	}

	fallback := LyricWord{StartTime: 1000, EndTime: 1000}
	if got := wordHighlightProgress(fallback, 1080); math.Abs(got-0.5) > 1e-9 {
		t.Errorf("wordHighlightProgress fallback = %v, want 0.5", got)
	}
}

func TestWordHighlightRangeForProgress(t *testing.T) {
	tests := []struct {
		word                         string
		progress                     float64
		wantHighlighted, wantCurrent int
		wantCurrentProgress          float64
	}{
		{word: "你好世界", progress: 0, wantHighlighted: 0, wantCurrent: 0, wantCurrentProgress: 0},
		{word: "你好世界", progress: 0.125, wantHighlighted: 0, wantCurrent: 1, wantCurrentProgress: 0.5},
		{word: "你好世界", progress: 0.5, wantHighlighted: 2, wantCurrent: 0, wantCurrentProgress: 0},
		{word: "A😀é", progress: 0.5, wantHighlighted: 1, wantCurrent: 2, wantCurrentProgress: 0.5},
		{word: "A😀é", progress: 1, wantHighlighted: 5, wantCurrent: 0, wantCurrentProgress: 0},
	}

	for _, tt := range tests {
		got := wordHighlightRangeForProgress(tt.word, tt.progress)
		if got.highlightedLength != tt.wantHighlighted || got.transitioningLength != tt.wantCurrent || math.Abs(got.transitionProgress-tt.wantCurrentProgress) > 1e-9 {
			t.Errorf("wordHighlightRangeForProgress(%q, %v) = %+v, want highlighted=%d current=%d progress=%v", tt.word, tt.progress, got, tt.wantHighlighted, tt.wantCurrent, tt.wantCurrentProgress)
		}
	}
}

func TestConstrainWindowWidth(t *testing.T) {
	tests := []struct {
		width float64
		want  float64
	}{
		{width: 80, want: 100},
		{width: 320, want: 320},
		{width: 720, want: 500},
	}

	for _, tt := range tests {
		if got := constrainWindowWidth(tt.width, 100, 500); got != tt.want {
			t.Errorf("constrainWindowWidth(%v) = %v, want %v", tt.width, got, tt.want)
		}
	}
}

func TestWindowResizeWidth(t *testing.T) {
	tests := []struct {
		start, target, elapsed float64
		want                   float64
		done                   bool
	}{
		{start: 100, target: 300, elapsed: 0, want: 100, done: false},
		{start: 100, target: 300, elapsed: 0.05, want: 131.25, done: false},
		{start: 100, target: 300, elapsed: 0.1, want: 200, done: false},
		{start: 100, target: 300, elapsed: 0.2, want: 300, done: true},
		{start: 300, target: 100, elapsed: 0.1, want: 200, done: false},
	}

	for _, tt := range tests {
		got, done := windowResizeWidth(tt.start, tt.target, tt.elapsed)
		if math.Abs(got-tt.want) > 1e-9 || done != tt.done {
			t.Errorf("windowResizeWidth(%v, %v, %v) = (%v, %t), want (%v, %t)", tt.start, tt.target, tt.elapsed, got, done, tt.want, tt.done)
		}
	}
}

func assertScrollState(t *testing.T, state *scrollState, wantOffset, wantPause float64) {
	t.Helper()
	if math.Abs(state.offset-wantOffset) > 1e-9 {
		t.Errorf("offset = %v, want %v", state.offset, wantOffset)
	}
	if math.Abs(state.pauseTimer-wantPause) > 1e-9 {
		t.Errorf("pauseTimer = %v, want %v", state.pauseTimer, wantPause)
	}
}
