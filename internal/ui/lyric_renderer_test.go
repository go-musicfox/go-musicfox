package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/lyric"
	"github.com/go-musicfox/go-musicfox/internal/ui/kitty"
)

func TestYRCWordTimings(t *testing.T) {
	line := lyric.YRCLine{Words: []lyric.YRCWord{
		{Word: "甲", StartTime: 0, EndTime: 500},
		{Word: "乙", StartTime: 500, EndTime: 1000},
		{Word: "丙", StartTime: 1000, EndTime: 1500},
	}}

	words, currentWord, playedWords := yrcWordTimings(line, 750)

	if currentWord != 1 {
		t.Errorf("current word = %d, want 1", currentWord)
	}
	if playedWords != 2 {
		t.Errorf("played words = %d, want 2", playedWords)
	}
	if words[0].state != wordStatePlayed || words[0].interpolation != 1 {
		t.Errorf("first word = %+v, want completed", words[0])
	}
	if words[1].state != wordStatePlaying || words[1].interpolation != 0.5 {
		t.Errorf("second word = %+v, want halfway playing", words[1])
	}
	if words[2].state != wordStateNotPlayed {
		t.Errorf("third word state = %v, want not played", words[2].state)
	}
}

func TestBuildYRCLineStringPreservesTextAndTranslation(t *testing.T) {
	previousConfig := configs.AppConfig
	configs.AppConfig = &configs.Config{}
	t.Cleanup(func() { configs.AppConfig = previousConfig })

	line := lyric.YRCLine{
		Words:           []lyric.YRCWord{{Word: "甲", StartTime: 0, EndTime: 500}, {Word: "乙", StartTime: 500, EndTime: 1000}},
		TranslatedLyric: "translation",
	}

	rendered := (&LyricRenderer{}).buildYRCLineString(line, 100, true)
	if got, want := ansi.Strip(rendered), "甲乙 [translation]"; got != want {
		t.Errorf("visible lyric = %q, want %q", got, want)
	}
}

func TestPlaceholderRowsPreserveGeometry(t *testing.T) {
	kitty.SetTmuxPassthroughForTest(true)
	t.Cleanup(func() { kitty.SetTmuxPassthroughForTest(false) })

	cover := &CoverRenderer{
		imageRendered:  true,
		displayImageID: 42,
		cols:           4,
		rows:           2,
		lastStartRow:   10,
		lastStartCol:   3,
	}
	lr := &LyricRenderer{netease: &Netease{coverRenderer: cover}}

	var pad strings.Builder
	lr.writePaddingRow(&pad, 10, 20)
	gotPad := pad.String()
	if ansi.StringWidth(strings.TrimSuffix(gotPad, "\n")) != 20 {
		t.Fatalf("padding row width must match the pane, got %q", gotPad)
	}
	if !strings.ContainsRune(gotPad, '\U0010EEEE') {
		t.Fatal("padding row missing placeholder cells")
	}

	var lead strings.Builder
	lr.writeLyricLeadingPadding(&lead, 11, 12)
	gotLead := lead.String()
	if ansi.StringWidth(gotLead) != 12 {
		t.Fatalf("leading padding width must match the text position, got %q", gotLead)
	}

	var noPH strings.Builder
	lr.writePaddingRow(&noPH, 99, 20)
	if strings.ContainsRune(noPH.String(), '\U0010EEEE') {
		t.Fatal("rows outside the cover must not contain placeholders")
	}
}
