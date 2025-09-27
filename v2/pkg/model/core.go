// Package model defines the core data structures and enums for go-musicfox v2
package model

import (
	"time"
)

const (
	// UnknownValue represents an unknown or invalid enum value
	UnknownValue = "unknown"
)

// PlayStatus represents the current playback status
type PlayStatus int

const (
	PlayStatusStopped PlayStatus = iota
	PlayStatusPlaying
	PlayStatusPaused
	PlayStatusBuffering
	PlayStatusError
)

// String returns the string representation of PlayStatus
func (ps PlayStatus) String() string {
	switch ps {
	case PlayStatusStopped:
		return "stopped"
	case PlayStatusPlaying:
		return "playing"
	case PlayStatusPaused:
		return "paused"
	case PlayStatusBuffering:
		return "buffering"
	case PlayStatusError:
		return "error"
	default:
		return UnknownValue
	}
}

// PlayMode represents the playback mode
type PlayMode int

const (
	PlayModeSequential PlayMode = iota
	PlayModeRepeatOne
	PlayModeRepeatAll
	PlayModeShuffle
)

// String returns the string representation of PlayMode
func (pm PlayMode) String() string {
	switch pm {
	case PlayModeSequential:
		return "sequential"
	case PlayModeRepeatOne:
		return "repeat_one"
	case PlayModeRepeatAll:
		return "repeat_all"
	case PlayModeShuffle:
		return "shuffle"
	default:
		return UnknownValue
	}
}

// Quality represents the audio quality level
type Quality int

const (
	QualityLow Quality = iota
	QualityMedium
	QualityHigh
	QualityLossless
)

// String returns the string representation of Quality
func (q Quality) String() string {
	switch q {
	case QualityLow:
		return "low"
	case QualityMedium:
		return "medium"
	case QualityHigh:
		return "high"
	case QualityLossless:
		return "lossless"
	default:
		return UnknownValue
	}
}

// Song represents a music track
type Song struct {
	ID        string            `json:"id" validate:"required"`
	Title     string            `json:"title" validate:"required"`
	Artist    string            `json:"artist" validate:"required"`
	Album     string            `json:"album"`
	Duration  time.Duration     `json:"duration"`
	Source    string            `json:"source" validate:"required"`
	URL       string            `json:"url"`
	CoverURL  string            `json:"cover_url"`
	Quality   Quality           `json:"quality"`
	Metadata  map[string]string `json:"metadata"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// NewSong creates a new Song instance with default values
func NewSong(id, title, artist, source string) *Song {
	now := time.Now()
	return &Song{
		ID:        id,
		Title:     title,
		Artist:    artist,
		Source:    source,
		Quality:   QualityMedium,
		Metadata:  make(map[string]string),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Playlist represents a collection of songs
type Playlist struct {
	ID          string    `json:"id" validate:"required"`
	Name        string    `json:"name" validate:"required"`
	Description string    `json:"description"`
	Songs       []*Song   `json:"songs"`
	Source      string    `json:"source" validate:"required"`
	CreatedBy   string    `json:"created_by"`
	IsPublic    bool      `json:"is_public"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// NewPlaylist creates a new Playlist instance with default values
func NewPlaylist(id, name, source, createdBy string) *Playlist {
	now := time.Now()
	return &Playlist{
		ID:        id,
		Name:      name,
		Source:    source,
		CreatedBy: createdBy,
		Songs:     make([]*Song, 0),
		IsPublic:  false,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// AddSong adds a song to the playlist
func (p *Playlist) AddSong(song *Song) {
	p.Songs = append(p.Songs, song)
	p.UpdatedAt = time.Now()
}

// RemoveSong removes a song from the playlist by ID
func (p *Playlist) RemoveSong(songID string) bool {
	for i, song := range p.Songs {
		if song.ID == songID {
			p.Songs = append(p.Songs[:i], p.Songs[i+1:]...)
			p.UpdatedAt = time.Now()
			return true
		}
	}
	return false
}

// GetSongCount returns the number of songs in the playlist
func (p *Playlist) GetSongCount() int {
	return len(p.Songs)
}

// PlayerState represents the current state of the music player
type PlayerState struct {
	Status      PlayStatus    `json:"status"`
	CurrentSong *Song         `json:"current_song"`
	Position    time.Duration `json:"position"`
	Duration    time.Duration `json:"duration"`
	Volume      float64       `json:"volume" validate:"min=0,max=1"`
	IsMuted     bool          `json:"is_muted"`
	PlayMode    PlayMode      `json:"play_mode"`
	Queue       []*Song       `json:"queue"`
	History     []*Song       `json:"history"`
}

// NewPlayerState creates a new PlayerState instance with default values
func NewPlayerState() *PlayerState {
	return &PlayerState{
		Status:   PlayStatusStopped,
		Volume:   0.8,
		IsMuted:  false,
		PlayMode: PlayModeSequential,
		Queue:    make([]*Song, 0),
		History:  make([]*Song, 0),
	}
}

// IsPlaying returns true if the player is currently playing
func (ps *PlayerState) IsPlaying() bool {
	return ps.Status == PlayStatusPlaying
}

// IsPaused returns true if the player is paused
func (ps *PlayerState) IsPaused() bool {
	return ps.Status == PlayStatusPaused
}

// User represents a user of the application
type User struct {
	ID       string `json:"id" validate:"required"`
	Username string `json:"username" validate:"required"`
	Email    string `json:"email" validate:"email"`
	Avatar   string `json:"avatar"`
}

// AppState represents the overall application state
type AppState struct {
	Player      *PlayerState      `json:"player"`
	CurrentView string            `json:"current_view"`
	User        *User             `json:"user"`
	Config      map[string]string `json:"config"`
	Plugins     []string          `json:"plugins"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// NewAppState creates a new AppState instance with default values
func NewAppState() *AppState {
	return &AppState{
		Player:      NewPlayerState(),
		CurrentView: "main",
		Config:      make(map[string]string),
		Plugins:     make([]string, 0),
		UpdatedAt:   time.Now(),
	}
}

// UpdateConfig updates a configuration value
func (as *AppState) UpdateConfig(key, value string) {
	as.Config[key] = value
	as.UpdatedAt = time.Now()
}

// GetConfig retrieves a configuration value
func (as *AppState) GetConfig(key string) (string, bool) {
	value, exists := as.Config[key]
	return value, exists
}

// AddPlugin adds a plugin to the application state
func (as *AppState) AddPlugin(pluginName string) {
	for _, plugin := range as.Plugins {
		if plugin == pluginName {
			return // Plugin already exists
		}
	}
	as.Plugins = append(as.Plugins, pluginName)
	as.UpdatedAt = time.Now()
}

// RemovePlugin removes a plugin from the application state
func (as *AppState) RemovePlugin(pluginName string) bool {
	for i, plugin := range as.Plugins {
		if plugin == pluginName {
			as.Plugins = append(as.Plugins[:i], as.Plugins[i+1:]...)
			as.UpdatedAt = time.Now()
			return true
		}
	}
	return false
}