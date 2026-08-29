package core

import (
	"testing"
	"time"

	"github.com/go-musicfox/go-musicfox/internal/player"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
)

// stubEngine is a minimal player.Player with a controllable State so tests can
// exercise CommandContext without a real audio engine.
type stubEngine struct {
	state types.State
}

func (s *stubEngine) Play(player.URLMusic)           {}
func (s *stubEngine) CurMusic() player.URLMusic      { return player.URLMusic{} }
func (s *stubEngine) Pause()                         {}
func (s *stubEngine) Resume()                        {}
func (s *stubEngine) Stop()                          {}
func (s *stubEngine) Toggle()                        {}
func (s *stubEngine) Seek(time.Duration)             {}
func (s *stubEngine) PassedTime() time.Duration      { return 0 }
func (s *stubEngine) PlayedTime() time.Duration      { return 0 }
func (s *stubEngine) TimeChan() <-chan time.Duration { return nil }
func (s *stubEngine) State() types.State             { return s.state }
func (s *stubEngine) StateChan() <-chan types.State  { return nil }
func (s *stubEngine) Volume() int                    { return 0 }
func (s *stubEngine) SetVolume(int)                  {}
func (s *stubEngine) UpVolume()                      {}
func (s *stubEngine) DownVolume()                    {}
func (s *stubEngine) Close()                         {}

// stubPlaylistManager returns a fixed playlist/current-index snapshot so tests
// can exercise the Song boundary without mutating a real playlist manager
// (which clamps out-of-range indices instead of exposing them).
type stubPlaylistManager struct {
	playlist []structs.Song
	index    int
}

func (s *stubPlaylistManager) Initialize(index int, pl []structs.Song) error { return nil }
func (s *stubPlaylistManager) SupportedPlayModes() []types.Mode              { return nil }
func (s *stubPlaylistManager) GetPlaylist() []structs.Song                   { return s.playlist }
func (s *stubPlaylistManager) GetCurrentIndex() int                          { return s.index }
func (s *stubPlaylistManager) GetCurrentSong() (structs.Song, error)         { return structs.Song{}, nil }
func (s *stubPlaylistManager) NextSong(manual bool) (structs.Song, error)    { return structs.Song{}, nil }
func (s *stubPlaylistManager) PreviousSong(manual bool) (structs.Song, error) {
	return structs.Song{}, nil
}
func (s *stubPlaylistManager) RemoveSong(index int) (structs.Song, error) { return structs.Song{}, nil }
func (s *stubPlaylistManager) SetPlayMode(mode types.Mode) error          { return nil }
func (s *stubPlaylistManager) GetPlayMode() types.Mode                    { return types.PmOrdered }
func (s *stubPlaylistManager) GetPlayModeName() string                    { return "" }
func (s *stubPlaylistManager) SaveState() error                           { return nil }
func (s *stubPlaylistManager) LoadState() error                           { return nil }

// newCommandContextPlayer builds a Player wired with the stub engine and stub
// playlist manager, plus the given shared user slot.
func newCommandContextPlayer(state types.State, user *structs.User, songs []structs.Song, index int) *Player {
	userSlot := user
	return &Player{
		Player:          &stubEngine{state: state},
		playlistManager: &stubPlaylistManager{playlist: songs, index: index},
		user:            &userSlot,
	}
}

func TestPlayerCommandContextFullSnapshot(t *testing.T) {
	user := &structs.User{UserId: 42, Nickname: "fox"}
	songs := []structs.Song{{
		Id:      1,
		Name:    "Song One",
		Artists: []structs.Artist{{Name: "A1"}, {Name: "A2"}},
		Album:   structs.Album{Name: "Album X"},
	}}
	p := newCommandContextPlayer(types.Playing, user, songs, 0)

	ctx := p.CommandContext()
	if ctx.UserID != 42 || ctx.UserName != "fox" {
		t.Fatalf("user fields = %d/%q, want 42/fox", ctx.UserID, ctx.UserName)
	}
	if !ctx.Playing {
		t.Fatal("Playing = false, want true")
	}
	if ctx.Song == nil {
		t.Fatal("Song = nil, want snapshot")
	}
	if ctx.Song.ID != 1 || ctx.Song.Name != "Song One" || ctx.Song.Artist != "A1,A2" || ctx.Song.Album != "Album X" {
		t.Fatalf("Song = %+v, want {ID:1 Name:Song One Artist:A1,A2 Album:Album X}", ctx.Song)
	}
}

func TestPlayerCommandContextNotPlayingKeepsSong(t *testing.T) {
	songs := []structs.Song{{Id: 7, Name: "Paused Song"}}
	p := newCommandContextPlayer(types.Paused, nil, songs, 0)

	ctx := p.CommandContext()
	if ctx.Playing {
		t.Fatal("Playing = true for Paused state, want false")
	}
	if ctx.Song == nil || ctx.Song.ID != 7 {
		t.Fatalf("Song = %+v, want {ID: 7}", ctx.Song)
	}
}

func TestPlayerCommandContextEmpty(t *testing.T) {
	p := newCommandContextPlayer(types.Stopped, nil, nil, -1)

	ctx := p.CommandContext()
	if ctx.UserID != 0 || ctx.UserName != "" {
		t.Fatalf("user fields = %d/%q, want zero", ctx.UserID, ctx.UserName)
	}
	if ctx.Playing {
		t.Fatal("Playing = true, want false")
	}
	if ctx.Song != nil {
		t.Fatalf("Song = %+v, want nil", ctx.Song)
	}
}

func TestPlayerCommandContextNilReceiver(t *testing.T) {
	var p *Player
	ctx := p.CommandContext() // must not panic
	if ctx.UserID != 0 || ctx.Playing || ctx.Song != nil {
		t.Fatalf("nil receiver snapshot = %+v, want zero value", ctx)
	}
}

func TestPlayerCommandContextIndexOutOfRange(t *testing.T) {
	songs := []structs.Song{{Id: 1}, {Id: 2}}

	// index beyond the playlist length.
	high := newCommandContextPlayer(types.Paused, nil, songs, 5)
	if ctx := high.CommandContext(); ctx.Song != nil {
		t.Fatalf("Song = %+v for out-of-range index, want nil", ctx.Song)
	}

	// negative index.
	neg := newCommandContextPlayer(types.Paused, nil, songs, -1)
	if ctx := neg.CommandContext(); ctx.Song != nil {
		t.Fatalf("Song = %+v for negative index, want nil", ctx.Song)
	}
}
