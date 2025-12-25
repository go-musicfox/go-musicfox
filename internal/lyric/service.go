package lyric

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/pkg/errors"
)

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// State represents the complete current state of the lyric service, for consumption by the UI layer.
type State struct {
	Fragments           []LRCFragment
	TranslatedFragments map[int64]string
	CurrentIndex        int
	IsRunning           bool
	// Word-by-word lyric data
	YRCLines     []YRCLine
	YRCLineIndex int  // Current YRC line index
	YRCEnabled   bool // Whether YRC mode is active
}

// FormatAsLRC serializes the State into a string that conforms to the LRC file format standard.
// This is a data representation method, not a UI rendering method.
func (s State) FormatAsLRC() string {
	if !s.IsRunning || len(s.Fragments) == 0 {
		return "[00:00.00]No lyrics available.~"
	}

	var builder strings.Builder
	for _, line := range s.Fragments {
		at := time.Duration(line.StartTimeMs) * time.Millisecond
		builder.WriteString(fmt.Sprintf("[%02d:%05.2f]", at/time.Minute, (at % time.Minute).Seconds()))
		builder.WriteString(line.Content)
		if trans, ok := s.TranslatedFragments[line.StartTimeMs]; ok && trans != "" {
			builder.WriteString(" [")
			builder.WriteString(trans)
			builder.WriteString("]")
		}
		builder.WriteString("\n")
	}
	return builder.String()
}

// Service handles all business logic related to lyrics.
type Service struct {
	fetcher Fetcher

	// Raw data cache
	lastLRCData   structs.LRCData
	currentSongID int64

	// Parsed data
	fragments      []LRCFragment
	transFragments map[int64]string
	yrcLines       []YRCLine // Word-by-word lyric data
	yrcIndex       int       // Current word-by-word lyric line index

	// Internal state & configuration
	currentIndex    int
	isRunning       bool
	showTranslation bool
	showYRC         bool // Enable word-by-word lyric mode
	offset          time.Duration
	skipParseErr    bool

	mu sync.RWMutex
}

// NewService creates a new lyric service.
func NewService(fetcher Fetcher, showTranslation bool, initialOffset time.Duration, skipParseErr bool) *Service {
	return &Service{
		fetcher:         fetcher,
		currentIndex:    -1,
		yrcIndex:        -1,
		showTranslation: showTranslation,
		showYRC:         false,
		offset:          initialOffset,
		transFragments:  make(map[int64]string),
		skipParseErr:    skipParseErr,
	}
}

// SetSong loads lyrics for a new song.
func (s *Service) SetSong(ctx context.Context, song structs.Song) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.resetState(true) // Preserve configuration on reset

	lrcData, err := s.fetcher.GetLyric(ctx, song.Id)
	if err != nil {
		return errors.Wrap(err, "failed to fetch lyric data")
	}

	s.lastLRCData = lrcData
	s.currentSongID = song.Id

	// 网易云新版 API 可能返回混合格式：前面几行 JSON 格式元数据 + 后面传统 LRC 格式
	lrcOriginal := lrcData.Original
	hasJsonLines := strings.Contains(lrcOriginal, `{"t":`)
	hasLrcLines := strings.Contains(lrcOriginal, "[00:")

	if hasJsonLines && hasLrcLines {
		// 混合格式：逐行解析
		slog.Debug("[LRC] Detected mixed format (JSON + traditional LRC)")
		lines := strings.Split(lrcOriginal, "\n")
		var jsonLines []string
		var lrcLines []string

		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			if strings.HasPrefix(line, "{") {
				jsonLines = append(jsonLines, line)
			} else if strings.HasPrefix(line, "[") {
				lrcLines = append(lrcLines, line)
			}
		}

		// 解析 JSON 格式的行（元数据）
		if len(jsonLines) > 0 {
			yrcLines, err := ParseYRC(strings.Join(jsonLines, "\n"))
			if err == nil {
				for _, yrcLine := range yrcLines {
					var lineText strings.Builder
					for _, word := range yrcLine.Words {
						lineText.WriteString(word.Word)
					}
					s.fragments = append(s.fragments, LRCFragment{
						StartTimeMs: yrcLine.StartTime,
						Content:     lineText.String(),
					})
				}
			}
		}

		// 解析传统 LRC 格式的行（歌词）
		if len(lrcLines) > 0 {
			lrcFile, err := ReadLRC(strings.NewReader(strings.Join(lrcLines, "\n")))
			if err == nil && lrcFile != nil {
				s.fragments = append(s.fragments, lrcFile.fragments...)
			} else if err != nil {
				slog.Debug("[LRC] Failed to parse traditional LRC lines", "error", err)
			}
		}

		// 按时间排序
		sort.Slice(s.fragments, func(i, j int) bool {
			return s.fragments[i].StartTimeMs < s.fragments[j].StartTimeMs
		})

		slog.Debug("[LRC] Parsed mixed format", "totalFragments", len(s.fragments))
	} else if hasJsonLines {
		// 纯 JSON 格式
		slog.Debug("[LRC] Detected pure JSON format LRC")
		yrcLines, err := ParseYRC(lrcOriginal)
		if err == nil && len(yrcLines) > 0 {
			for _, yrcLine := range yrcLines {
				var lineText strings.Builder
				for _, word := range yrcLine.Words {
					lineText.WriteString(word.Word)
				}
				s.fragments = append(s.fragments, LRCFragment{
					StartTimeMs: yrcLine.StartTime,
					Content:     lineText.String(),
				})
			}
			slog.Debug("[LRC] Converted JSON LRC to fragments", "count", len(s.fragments))
		} else {
			slog.Warn("[LRC] Failed to parse JSON format LRC", "error", err)
		}
	} else {
		// 传统 LRC 格式 [00:00.00]歌词
		lrcFile, err := ReadLRC(strings.NewReader(lrcData.Original))
		if err != nil {
			if !s.skipParseErr {
				return errors.Wrap(err, "failed to parse original lyric")
			}
			slog.Debug("ignoring lyric parsing error", "error", err)
		}

		if lrcFile != nil {
			s.fragments = lrcFile.fragments
		}
	}

	if s.showTranslation {
		if trans, err := ReadTranslateLRC(strings.NewReader(lrcData.Translated)); err == nil {
			s.transFragments = trans.fragments
		} else {
			if s.skipParseErr && trans != nil {
				s.transFragments = trans.fragments
				slog.Debug("ignoring lyric translation parsing error", "error", err)
			} else {
				slog.Error("failed to parse lyric translationd", "error", err)
			}
		}
	}

	// Parse YRC (word-by-word) lyrics if available
	if lrcData.Yrc != "" {
		slog.Debug("[YRC] Starting to parse YRC data", "dataLen", len(lrcData.Yrc))
		slog.Debug("[YRC] Raw YRC data", "data", lrcData.Yrc)
		yrcLines, err := ParseYRC(lrcData.Yrc)
		if err == nil && len(yrcLines) > 0 {
			slog.Debug("[YRC] Successfully parsed YRC", "lines", len(yrcLines), "firstLineWords", len(yrcLines[0].Words))
			s.yrcLines = yrcLines

			// Optionally align translation and roman lyrics to YRC
			if lrcData.Ytlrc != "" {
				s.yrcLines = AlignTranslationToYRC(s.yrcLines, lrcData.Ytlrc)
			}
			if lrcData.Yromalrc != "" {
				s.yrcLines = AlignRomanToYRC(s.yrcLines, lrcData.Yromalrc)
			}

			slog.Debug("[YRC] Successfully stored YRC lines", "totalLines", len(s.yrcLines))
		} else if err != nil {
			slog.Warn("[YRC] Failed to parse YRC lyric", "error", err)
		} else {
			slog.Warn("[YRC] Parsed YRC but got 0 lines")
		}
	} else {
		slog.Debug("[YRC] No YRC data in lrcData")
	}

	s.isRunning = true
	return nil
}

// UpdatePosition updates the current playback position and computes the current lyric index.
func (s *Service) UpdatePosition(duration time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isRunning {
		return
	}

	timeMs := duration.Milliseconds() + s.offset.Milliseconds()

	// Update LRC index
	if len(s.fragments) > 0 {
		newIndex := -1
		for i := len(s.fragments) - 1; i >= 0; i-- {
			if timeMs >= s.fragments[i].StartTimeMs {
				newIndex = i
				break
			}
		}
		s.currentIndex = newIndex
	}

	// Update YRC line index
	if len(s.yrcLines) > 0 {
		s.yrcIndex = FindYRCLineAtTimeMs(s.yrcLines, timeMs)
	}
}

// EnableTranslation dynamically enables or disables lyric translation.
func (s *Service) EnableTranslation(enable bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.showTranslation == enable {
		return
	}
	s.showTranslation = enable

	if !s.isRunning {
		return
	}

	if enable {
		transLrcFile, _ := ReadTranslateLRC(strings.NewReader(s.lastLRCData.Translated))
		s.transFragments = transLrcFile.fragments
	} else {
		s.transFragments = make(map[int64]string)
	}
}

// EnableYRC dynamically enables or disables word-by-word lyric (YRC) mode.
func (s *Service) EnableYRC(enable bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.showYRC == enable {
		return
	}
	s.showYRC = enable
}

// SetOffset dynamically sets the lyric time offset.
func (s *Service) SetOffset(offset time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.offset = offset
}

// State returns the current state of the lyric service in a thread-safe manner.
func (s *Service) State() State {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return State{
		Fragments:           s.fragments,
		TranslatedFragments: s.transFragments,
		CurrentIndex:        s.currentIndex,
		IsRunning:           s.isRunning,
		YRCLines:            s.yrcLines,
		YRCLineIndex:        s.yrcIndex,
		YRCEnabled:          s.showYRC && len(s.yrcLines) > 0,
	}
}

// Stop stops the lyric service and clears its state.
func (s *Service) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.resetState(false) // Full reset
}

// resetState clears the internal state.
func (s *Service) resetState(preserveConfig bool) {
	s.fragments = nil
	s.transFragments = make(map[int64]string)
	s.yrcLines = nil
	s.currentIndex = -1
	s.yrcIndex = -1
	s.isRunning = false
	s.lastLRCData = structs.LRCData{}
	s.currentSongID = 0
	if !preserveConfig {
		s.showTranslation = false
		s.showYRC = false
		s.offset = 0
	}
}
