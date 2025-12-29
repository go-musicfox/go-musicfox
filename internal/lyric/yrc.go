package lyric

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"sort"
	"strings"
)

type YRCWord struct {
	Word      string `json:"word"`
	StartTime int64  `json:"startTime"` // 毫秒
	EndTime   int64  `json:"endTime"`   // 毫秒
}

// YRCLine 表示包含词语和可选翻译的完整 YRC 歌词行。
type YRCLine struct {
	StartTime       int64
	EndTime         int64
	Words           []YRCWord
	TranslatedLyric string // 整行的翻译文本
	RomanLyric      string // 整行的罗马音文本
	IsBG            bool   // 背景歌词标志
	IsDuet          bool   // 二重唱（与他人合唱）标志
}

// ParseYRC解析API返回的YRC JSON字符串并返回YRCLine切片。
// YRC 格式有两种：
// 1. 纯 JSON：{"t":timestamp,"c":[{"tx":"word","tr":[offset,duration]}...]}
// 2. LRC 风格：[startTime,duration](wordTime,duration,0)word(wordTime,duration,0)word...
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

		// 首先尝试解析为纯 JSON（用于作曲家、作词家等元数据行）
		if strings.HasPrefix(line, "{") {
			if yrcLine := parseJSONYRCLine(line); yrcLine != nil {
				result = append(result, *yrcLine)
				continue
			}
		}

		// 尝试解析为 LRC 风格格式 [time,duration](time,dur,0)word(time,dur,0)word...
		if strings.HasPrefix(line, "[") {
			if yrcLine := parseLRCStyleYRCLine(line); yrcLine != nil {
				result = append(result, *yrcLine)
				continue
			}
		}

		slog.Debug("failed to parse YRC line", "lineNum", lineNum, "preview", line[:min(80, len(line))])
	}

	// 按开始时间排序
	sort.Slice(result, func(i, j int) bool {
		return result[i].StartTime < result[j].StartTime
	})

	return result, nil
}

// parseJSONYRCLine 解析纯 JSON 格式：{"t":time,"c":[{"tx":"word","tr":[offset,duration]}]}
func parseJSONYRCLine(line string) *YRCLine {
	var lineData struct {
		T int64 `json:"t"` // 开始时间（毫秒）
		C []struct {
			Tx string  `json:"tx"`           // 文本（词）
			Tr []int64 `json:"tr,omitempty"` // [起始偏移, 时长]
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
			// 没有时间信息，估算
			word.EndTime = currentTime + 500 // 默认 500 毫秒
			currentTime = word.EndTime
		}

		yrcLine.Words = append(yrcLine.Words, word)
	}

	// 设置行的结束时间
	if len(yrcLine.Words) > 0 {
		yrcLine.EndTime = yrcLine.Words[len(yrcLine.Words)-1].EndTime
	} else {
		yrcLine.EndTime = yrcLine.StartTime
	}

	return &yrcLine
}

// parseLRCStyleYRCLine 解析 LRC 风格格式：[startTime,duration](wordTime,duration,0)word(wordTime,duration,0)word...
func parseLRCStyleYRCLine(line string) *YRCLine {
	// 提取开头的 [startTime,duration]
	closeBracket := strings.Index(line, "]")
	if closeBracket == -1 {
		return nil
	}

	// 解析行的时间 [startTime,duration]
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

	// 解析词段：(wordTime,duration,0)word
	content := line[closeBracket+1:]

	// 使用正则表达式匹配模式：(time,duration,0)word
	// 但由于词可以是任何 Unicode 字符，我们手动解析
	i := 0
	for i < len(content) {
		// 查找下一个 (
		if content[i] != '(' {
			i++
			continue
		}

		// 提取时间信息 (time,duration,0)
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

		// 提取词（下一个字符或直到下一个 '(' 的字符）
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

// parseTimeTag 将 LRC 时间戳如 "01:23.45" 解析为毫秒
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

// AlignTranslationToYRC 通过匹配时间将翻译的 LRC 行合并到 YRC 行中。
// 这将行级别的翻译文本（来自 ytlrc.lyric）与 YRC 词数据对齐。
func AlignTranslationToYRC(yrcLines []YRCLine, translatedLRC string) []YRCLine {
	if translatedLRC == "" || len(yrcLines) == 0 {
		return yrcLines
	}

	// 解析翻译的 LRC 以获取时间 -> 翻译文本的映射
	transFile, err := ReadTranslateLRC(strings.NewReader(translatedLRC))
	if err != nil || transFile == nil {
		slog.Debug("failed to parse translated LRC for YRC alignment", "error", err)
		return yrcLines
	}

	// 按开始时间将 YRC 行与翻译片段匹配
	for i := range yrcLines {
		if trans, ok := transFile.fragments[yrcLines[i].StartTime]; ok && trans != "" {
			yrcLines[i].TranslatedLyric = trans
		}
	}

	return yrcLines
}

// AlignRomanToYRC 通过匹配时间将罗马音 LRC 行合并到 YRC 行中。
func AlignRomanToYRC(yrcLines []YRCLine, romanLRC string) []YRCLine {
	if romanLRC == "" || len(yrcLines) == 0 {
		return yrcLines
	}

	// 解析罗马音 LRC 以获取时间 -> 罗马音文本的映射
	romanFile, err := ReadTranslateLRC(strings.NewReader(romanLRC))
	if err != nil || romanFile == nil {
		slog.Debug("failed to parse roman LRC for YRC alignment", "error", err)
		return yrcLines
	}

	// 按开始时间将 YRC 行与罗马音片段匹配
	for i := range yrcLines {
		if roman, ok := romanFile.fragments[yrcLines[i].StartTime]; ok && roman != "" {
			yrcLines[i].RomanLyric = roman
		}
	}

	return yrcLines
}

// FindYRCLineAtTimeMs 查找给定时间（毫秒）处活动的 YRC 行索引。
// 如果没有活动行则返回 -1。
func FindYRCLineAtTimeMs(yrcLines []YRCLine, timeMs int64) int {
	for i := len(yrcLines) - 1; i >= 0; i-- {
		if timeMs >= yrcLines[i].StartTime {
			return i
		}
	}
	return -1
}

// FindYRCWordAtTimeMs 查找 YRC 行中给定时间（毫秒）处活动的词索引。
// 如果没有活动词则返回 -1。
func FindYRCWordAtTimeMs(words []YRCWord, timeMs int64) int {
	for i := len(words) - 1; i >= 0; i-- {
		if timeMs >= words[i].StartTime {
			return i
		}
	}
	return -1
}
