package lyric

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/pkg/errors"
)

// State represents the complete current state of the lyric service, for consumption by the UI layer.
type State struct {
	Fragments           []LRCFragment
	TranslatedFragments map[int64]string
	CurrentIndex        int
	IsRunning           bool
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

	// Internal state & configuration
	currentIndex    int
	isRunning       bool
	showTranslation bool
	offset          time.Duration

	mu sync.RWMutex
}

// NewService creates a new lyric service.
func NewService(fetcher Fetcher, showTranslation bool, initialOffset time.Duration) *Service {
	return &Service{
		fetcher:         fetcher,
		currentIndex:    -1,
		showTranslation: showTranslation,
		offset:          initialOffset,
		transFragments:  make(map[int64]string),
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

	lrcFile, _ := ReadLRC(strings.NewReader(lrcData.Original))
	s.fragments = lrcFile.fragments

	if s.showTranslation {
		transLrcFile, _ := ReadTranslateLRC(strings.NewReader(lrcData.Translated))
		s.transFragments = transLrcFile.fragments
	}

	s.isRunning = true
	return nil
}

// UpdatePosition updates the current playback position and computes the current lyric index.
func (s *Service) UpdatePosition(duration time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isRunning || len(s.fragments) == 0 {
		return
	}

	timeMs := duration.Milliseconds() + s.offset.Milliseconds()

	newIndex := -1
	for i := len(s.fragments) - 1; i >= 0; i-- {
		if timeMs >= s.fragments[i].StartTimeMs {
			newIndex = i
			break
		}
	}

	s.currentIndex = newIndex
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
	s.currentIndex = -1
	s.isRunning = false
	s.lastLRCData = structs.LRCData{}
	s.currentSongID = 0
	if !preserveConfig {
		s.showTranslation = false
		s.offset = 0
	}
}
