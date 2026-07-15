package lyric

import "testing"

func TestProgressYRCLineAtTimeMs(t *testing.T) {
	line := YRCLine{Words: []YRCWord{
		{Word: "甲", StartTime: 100, EndTime: 300},
		{Word: "乙", StartTime: 500, EndTime: 700},
		{Word: "丙", StartTime: 700, EndTime: 900},
	}}

	tests := []struct {
		name              string
		timeMs            int64
		completedWords    int
		currentWord       int
		currentWordFactor float64
	}{
		{"before first word", 99, 0, -1, 0},
		{"mid first word", 200, 0, 0, 0.5},
		{"between words", 400, 1, -1, 0},
		{"mid second word", 600, 1, 1, 0.5},
		{"after line", 1000, 3, -1, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ProgressYRCLineAtTimeMs(line, tt.timeMs)
			if got.CompletedWords != tt.completedWords {
				t.Errorf("completed words = %d, want %d", got.CompletedWords, tt.completedWords)
			}
			if got.CurrentWord != tt.currentWord {
				t.Errorf("current word = %d, want %d", got.CurrentWord, tt.currentWord)
			}
			if got.CurrentProgress != tt.currentWordFactor {
				t.Errorf("current word progress = %v, want %v", got.CurrentProgress, tt.currentWordFactor)
			}
		})
	}
}

func TestProgressYRCLineAtTimeMsHandlesZeroDurationWord(t *testing.T) {
	line := YRCLine{Words: []YRCWord{{Word: "甲", StartTime: 100, EndTime: 100}}}
	got := ProgressYRCLineAtTimeMs(line, 100)

	if got.CompletedWords != 1 || got.CurrentWord != -1 {
		t.Errorf("zero duration word progress = %+v, want completed word", got)
	}
}
