package lyric

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Parses an LRC file
// https://en.wikipedia.org/wiki/LRC_(file_format)

type LRCFile struct {
	fragments []LRCFragment
}

type LRCFragment struct {
	StartTimeMs int64
	Content     string
}

func OpenLRCFile(filePath string) (lrcFile *LRCFile, err error) {
	file, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer file.Close()
	lrcFile, err = ReadLRC(file)
	return
}

func ReadLRC(reader io.Reader) (lrcFile *LRCFile, err error) {
	var fragments []LRCFragment

	lineNo := 1
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		var lineFragments []LRCFragment
		lineFragments, err = readLRCLine(scanner.Text(), lineNo)
		if err != nil {
			continue
		}
		for _, fragment := range lineFragments {
			fragments = append(fragments, fragment)
		}
		lineNo++
	}

	sort.Slice(fragments, func(i, j int) bool {
		return fragments[i].StartTimeMs < fragments[j].StartTimeMs
	})

	lrcFile = &LRCFile{
		fragments: fragments,
	}
	return
}

func readLRCLine(line string, lineNo int) (fragments []LRCFragment, err error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return
	}

	tm, err := parseLRCTime(line, "[", "]")
	if err != nil {
		err = fmt.Errorf("Error on line %d: %v", lineNo, err)
		return
	}

	closeIndex := strings.Index(line, "]")
	line = line[closeIndex+1:]

	var extraTms []time.Time
	for {
		extraTm, tmErr := parseLRCTime(line, "[", "]")
		if tmErr != nil {
			break
		}
		extraTms = append(extraTms, extraTm)
		line = line[closeIndex+1:]
	}

	line = strings.TrimSpace(line)
	if len(extraTms) > 0 {
		fragments = append(fragments, LRCFragment{
			StartTimeMs: getMillisecondsFromTime(tm),
			Content:     line,
		})

		for _, extraTm := range extraTms {
			fragments = append(fragments, LRCFragment{
				StartTimeMs: getMillisecondsFromTime(extraTm),
				Content:     line,
			})
		}
		return
	}

	lineFragments, err := parseContentLine(line, tm)
	for _, fragment := range lineFragments {
		fragments = append(fragments, fragment)
	}
	return
}

func parseLRCTime(line, openChar, closeChar string) (tm time.Time, err error) {
	closeIndex := strings.Index(line, closeChar)
	if line[0:1] != openChar || closeIndex < 0 || strings.Index(line, ":") < 0 || strings.Index(line, ".") < 0 {
		err = errors.New("brackets missing")
		return
	}

	_, err = strconv.Atoi(line[1:3])
	if err != nil {
		// A tag line
		return
	}

	part1 := strings.Split(line[1:closeIndex], ":")
	if len(part1) < 2 {
		err = errors.New("brackets missing")
		return
	}

	part2 := part1[1]
	part3 := strings.Split(part2, ".")
	if len(part3) < 2 {
		err = errors.New("brackets missing")
		return
	}

	minutes, err := strconv.Atoi(part1[0])
	seconds, err := strconv.Atoi(part3[0])
	milliseconds, err := strconv.ParseFloat(fmt.Sprintf("0.%s", part3[1]), 64)
	if err != nil {
		return
	}

	tm = time.Unix(int64(minutes*60+seconds), int64(math.Ceil(milliseconds*float64(time.Second))))

	return
}

func parseContentLine(line string, tm time.Time) (fragments []LRCFragment, err error) {
	if !strings.Contains(line, "<") {
		fragments = append(fragments, LRCFragment{
			StartTimeMs: getMillisecondsFromTime(tm),
			Content:     line,
		})
		return
	}

	previousTm := tm
	startIndex := 0
	lastIndex := 0
	for {
		idx := strings.Index(line[lastIndex:], "<")
		closeIndex := strings.Index(line[lastIndex:], ">")
		if idx < 0 || closeIndex < 0 {
			break
		}
		idx += lastIndex
		closeIndex += lastIndex

		splitTm, tmErr := parseLRCTime(line[idx:], "<", ">")
		if tmErr == nil {
			fragments = append(fragments, LRCFragment{
				StartTimeMs: getMillisecondsFromTime(previousTm),
				Content:     strings.TrimSpace(line[startIndex:idx]),
			})
			startIndex = closeIndex + 1
			previousTm = splitTm
		}
		lastIndex = closeIndex + 1
	}

	fragments = append(fragments, LRCFragment{
		StartTimeMs: getMillisecondsFromTime(previousTm),
		Content:     strings.TrimSpace(line[startIndex:]),
	})
	return
}

func getMillisecondsFromTime(tm time.Time) (ms int64) {
	ms = tm.Sub(time.Unix(0, 0)).Nanoseconds() / int64(time.Millisecond)
	return
}
