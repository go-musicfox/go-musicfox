package lyric

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"sort"
	"strings"
)

// YRCWord represents a single word in YRC (word-by-word) lyric with timing and metadata.
type YRCWord struct {
	Word      string `json:"word"`
	StartTime int64  `json:"startTime"` // ms
	EndTime   int64  `json:"endTime"`   // ms
}

// YRCLine represents a complete line of YRC lyric with words and optional translations.
type YRCLine struct {
	StartTime       int64
	EndTime         int64
	Words           []YRCWord
	TranslatedLyric string // Translated text for the entire line
	RomanLyric      string // Romanized text for the entire line
	IsBG            bool   // Background lyric flag
	IsDuet          bool   // Duet (singing with someone else) flag
}

// ParseYRC parses the YRC JSON string returned by Netease API and returns a slice of YRCLine.
// YRC format from Netease has TWO formats:
// 1. Pure JSON: {"t":timestamp,"c":[{"tx":"word","tr":[offset,duration]}...]}
// 2. LRC-style: [startTime,duration](wordTime,duration,0)word(wordTime,duration,0)word...
func ParseYRC(yrcData string) ([]YRCLine, error) {
	if yrcData == "" {
		return nil, nil
	}

	lines := strings.Split(yrcData, "\n")
	result := make([]YRCLine, 0, len(lines))

	for lineNum, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Try parsing as pure JSON first (for metadata lines like composer, lyricist)
		if strings.HasPrefix(line, "{") {
			if yrcLine := parseJSONYRCLine(line); yrcLine != nil {
				result = append(result, *yrcLine)
				continue
			}
		}

		// Try parsing as LRC-style format [time,duration](time,dur,0)word(time,dur,0)word...
		if strings.HasPrefix(line, "[") {
			if yrcLine := parseLRCStyleYRCLine(line); yrcLine != nil {
				result = append(result, *yrcLine)
				continue
			}
		}

		slog.Debug("failed to parse YRC line", "lineNum", lineNum, "preview", line[:min(80, len(line))])
	}

	// Sort by start time
	sort.Slice(result, func(i, j int) bool {
		return result[i].StartTime < result[j].StartTime
	})

	return result, nil
}

// parseJSONYRCLine parses pure JSON format: {"t":time,"c":[{"tx":"word","tr":[offset,duration]}]}
func parseJSONYRCLine(line string) *YRCLine {
	var lineData struct {
		T int64 `json:"t"` // Start time (ms)
		C []struct {
			Tx string  `json:"tx"`           // Text (word)
			Tr []int64 `json:"tr,omitempty"` // [start_offset, duration]
		} `json:"c"`
	}

	if err := json.Unmarshal([]byte(line), &lineData); err != nil {
		return nil
	}

	yrcLine := YRCLine{
		StartTime: lineData.T,
		Words:     make([]YRCWord, 0, len(lineData.C)),
	}

	var currentTime int64 = lineData.T
	for _, wordData := range lineData.C {
		word := YRCWord{
			Word:      wordData.Tx,
			StartTime: currentTime,
		}

		if len(wordData.Tr) >= 2 {
			word.StartTime = lineData.T + wordData.Tr[0]
			word.EndTime = word.StartTime + wordData.Tr[1]
			currentTime = word.EndTime
		} else {
			// No timing info, estimate
			word.EndTime = currentTime + 500 // 500ms default
			currentTime = word.EndTime
		}

		yrcLine.Words = append(yrcLine.Words, word)
	}

	// Set line end time
	if len(yrcLine.Words) > 0 {
		yrcLine.EndTime = yrcLine.Words[len(yrcLine.Words)-1].EndTime
	} else {
		yrcLine.EndTime = yrcLine.StartTime
	}

	return &yrcLine
}

// parseLRCStyleYRCLine parses LRC-style format: [startTime,duration](wordTime,duration,0)word(wordTime,duration,0)word...
func parseLRCStyleYRCLine(line string) *YRCLine {
	// Extract [startTime,duration] at the beginning
	closeBracket := strings.Index(line, "]")
	if closeBracket == -1 {
		return nil
	}

	// Parse line timing [startTime,duration]
	timingStr := line[1:closeBracket]
	timeParts := strings.Split(timingStr, ",")
	if len(timeParts) < 2 {
		return nil
	}

	var lineStartTime, lineDuration int64
	if _, err := fmt.Sscanf(timeParts[0], "%d", &lineStartTime); err != nil {
		return nil
	}
	if _, err := fmt.Sscanf(timeParts[1], "%d", &lineDuration); err != nil {
		return nil
	}

	yrcLine := YRCLine{
		StartTime: lineStartTime,
		EndTime:   lineStartTime + lineDuration,
		Words:     []YRCWord{},
	}

	// Parse word segments: (wordTime,duration,0)word
	content := line[closeBracket+1:]

	// Use regex to match pattern: (time,duration,0)word
	// But since words can be any Unicode character, we parse manually
	i := 0
	for i < len(content) {
		// Find next (
		if content[i] != '(' {
			i++
			continue
		}

		// Extract timing info (time,duration,0)
		closeParenIdx := strings.Index(content[i:], ")")
		if closeParenIdx == -1 {
			break
		}
		closeParenIdx += i

		timingInfo := content[i+1 : closeParenIdx]
		parts := strings.Split(timingInfo, ",")
		if len(parts) < 2 {
			i = closeParenIdx + 1
			continue
		}

		var wordTime, wordDuration int64
		if _, err := fmt.Sscanf(parts[0], "%d", &wordTime); err != nil {
			i = closeParenIdx + 1
			continue
		}
		if _, err := fmt.Sscanf(parts[1], "%d", &wordDuration); err != nil {
			i = closeParenIdx + 1
			continue
		}

		// Extract the word (next character or characters until next '(')
		i = closeParenIdx + 1
		wordStart := i
		for i < len(content) && content[i] != '(' {
			i++
		}

		wordText := content[wordStart:i]
		if wordText == "" {
			continue
		}

		word := YRCWord{
			Word:      wordText,
			StartTime: wordTime,
			EndTime:   wordTime + wordDuration,
		}
		yrcLine.Words = append(yrcLine.Words, word)
	}

	if len(yrcLine.Words) == 0 {
		return nil
	}

	return &yrcLine
}

// parseTimeTag parses LRC timestamp like "01:23.45" to milliseconds
func parseTimeTag(tag string) int64 {
	parts := strings.Split(tag, ":")
	if len(parts) != 2 {
		return -1
	}

	var min, sec int
	var ms int
	if _, err := fmt.Sscanf(parts[0], "%d", &min); err != nil {
		return -1
	}
	if strings.Contains(parts[1], ".") {
		if _, err := fmt.Sscanf(parts[1], "%d.%d", &sec, &ms); err != nil {
			return -1
		}
	} else {
		if _, err := fmt.Sscanf(parts[1], "%d", &sec); err != nil {
			return -1
		}
	}

	return int64(min)*60000 + int64(sec)*1000 + int64(ms)*10
}

// AlignTranslationToYRC aligns translated LRC lyrics with YRC lines by matching timestamps.// AlignTranslationToYRC merges translated LRC lines into YRC lines by matching timing.
// This aligns line-level translated text (from ytlrc.lyric) with YRC word data.
func AlignTranslationToYRC(yrcLines []YRCLine, translatedLRC string) []YRCLine {
	if translatedLRC == "" || len(yrcLines) == 0 {
		return yrcLines
	}

	// Parse translated LRC to get a map of time -> translation text
	transFile, err := ReadTranslateLRC(strings.NewReader(translatedLRC))
	if err != nil || transFile == nil {
		slog.Debug("failed to parse translated LRC for YRC alignment", "error", err)
		return yrcLines
	}

	// Match YRC lines with translated fragments by start time
	for i := range yrcLines {
		if trans, ok := transFile.fragments[yrcLines[i].StartTime]; ok && trans != "" {
			yrcLines[i].TranslatedLyric = trans
		}
	}

	return yrcLines
}

// AlignRomanToYRC merges romanized LRC lines into YRC lines by matching timing.
func AlignRomanToYRC(yrcLines []YRCLine, romanLRC string) []YRCLine {
	if romanLRC == "" || len(yrcLines) == 0 {
		return yrcLines
	}

	// Parse roman LRC to get a map of time -> roman text
	romanFile, err := ReadTranslateLRC(strings.NewReader(romanLRC))
	if err != nil || romanFile == nil {
		slog.Debug("failed to parse roman LRC for YRC alignment", "error", err)
		return yrcLines
	}

	// Match YRC lines with roman fragments by start time
	for i := range yrcLines {
		if roman, ok := romanFile.fragments[yrcLines[i].StartTime]; ok && roman != "" {
			yrcLines[i].RomanLyric = roman
		}
	}

	return yrcLines
}

// FindYRCLineAtTimeMs finds the YRC line index that is active at the given time (ms).
// Returns -1 if no line is active.
func FindYRCLineAtTimeMs(yrcLines []YRCLine, timeMs int64) int {
	for i := len(yrcLines) - 1; i >= 0; i-- {
		if timeMs >= yrcLines[i].StartTime {
			return i
		}
	}
	return -1
}

// FindYRCWordAtTimeMs finds the word index within a YRC line that is active at the given time (ms).
// Returns -1 if no word is active.
func FindYRCWordAtTimeMs(words []YRCWord, timeMs int64) int {
	for i := len(words) - 1; i >= 0; i-- {
		if timeMs >= words[i].StartTime {
			return i
		}
	}
	return -1
}
