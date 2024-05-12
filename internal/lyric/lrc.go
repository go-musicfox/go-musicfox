package lyric

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
)

// Parses an LRC file
// https://en.wikipedia.org/wiki/LRC_(file_format)

type LRCFile struct {
	fragments []LRCFragment

	text string
	l    sync.Mutex
}

func (f *LRCFile) AsText(t ...*TranslateLRCFile) string {
	if f == nil || len(f.fragments) == 0 {
		return "[00:00.00]暂无歌词~"
	}

	f.l.Lock()
	defer f.l.Unlock()

	if f.text != "" {
		return f.text
	}

	var trans *TranslateLRCFile
	if len(t) > 0 {
		trans = t[0]
	}

	var builder strings.Builder
	for _, line := range f.fragments {
		at := time.Duration(line.StartTimeMs) * time.Millisecond
		builder.WriteString(fmt.Sprintf("[%02d:%05.2f]", int(at.Minutes()), at.Seconds()))
		builder.WriteString(line.Content)
		if trans != nil && trans.fragments[line.StartTimeMs] != "" {
			builder.WriteString(" [")
			builder.WriteString(trans.fragments[line.StartTimeMs])
			builder.WriteString("]")
		}
		builder.WriteString("\n")
	}
	f.text = builder.String()

	return f.text
}

type TranslateLRCFile struct {
	fragments map[int64]string
}

func (tf *TranslateLRCFile) FindByTimeMs(timeMs int64) string {
	return tf.fragments[timeMs]
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
		fragments = append(fragments, lineFragments...)
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

func ReadTranslateLRC(reader io.Reader) (f *TranslateLRCFile, err error) {
	lrcFile, e := ReadLRC(reader)
	if e != nil {
		return nil, e
	}
	f = &TranslateLRCFile{fragments: map[int64]string{}}
	for _, fragment := range lrcFile.fragments {
		f.fragments[fragment.StartTimeMs] = fragment.Content
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
		err = fmt.Errorf("error on line %d: %v", lineNo, err)
		return
	}

	closeIndex := strings.Index(line, "]")
	line = line[closeIndex+1:]

	var extraTms []time.Duration
	for {
		extraTm, tmErr := parseLRCTime(line, "[", "]")
		if tmErr != nil {
			break
		}
		extraTms = append(extraTms, extraTm)
		closeIndex = strings.Index(line, "]")
		line = line[closeIndex+1:]
	}

	line = strings.TrimSpace(line)
	if len(extraTms) > 0 {
		fragments = append(fragments, LRCFragment{
			StartTimeMs: tm.Milliseconds(),
			Content:     line,
		})

		for _, extraTm := range extraTms {
			fragments = append(fragments, LRCFragment{
				StartTimeMs: extraTm.Milliseconds(),
				Content:     line,
			})
		}
		return
	}

	lineFragments, err := parseContentLine(line, tm)
	fragments = append(fragments, lineFragments...)
	return
}

func parseLRCTime(line, openChar, closeChar string) (tm time.Duration, err error) {
	defer func() {
		e := recover()
		if e != nil {
			err = errors.New(fmt.Sprint(e))
		}
	}()

	left := strings.Index(line, openChar)
	right := strings.Index(line, closeChar)
	if left < 0 || right < 0 {
		err = errors.New("brackets missing")
		return
	}
	for left > right && right < len(line) {
		line = line[right+1:]
		right = strings.Index(line, closeChar)
	}

	timeStr := line[left+1 : right]
	t := strings.Split(timeStr, ":")
	if len(t) > 1 && t[0] != "" && t[1] != "" {
		minutes, err1 := strconv.Atoi(t[0])
		seconds, err2 := strconv.ParseFloat(t[1], 64)
		if err1 != nil || err2 != nil {
			err = errors.New("format error")
			return
		}
		milliseconds := minutes*60000 + int(math.Floor(seconds*1000))
		tm = time.Duration(milliseconds) * time.Millisecond
		return
	}

	err = errors.New("brackets missing")
	return
}

func parseContentLine(line string, tm time.Duration) (fragments []LRCFragment, err error) {
	if !strings.Contains(line, "<") {
		fragments = append(fragments, LRCFragment{
			StartTimeMs: tm.Milliseconds(),
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
				StartTimeMs: previousTm.Milliseconds(),
				Content:     strings.TrimSpace(line[startIndex:idx]),
			})
			startIndex = closeIndex + 1
			previousTm = splitTm
		}
		lastIndex = closeIndex + 1
	}

	fragments = append(fragments, LRCFragment{
		StartTimeMs: previousTm.Milliseconds(),
		Content:     strings.TrimSpace(line[startIndex:]),
	})
	return
}
