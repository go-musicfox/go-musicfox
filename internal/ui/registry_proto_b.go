// PROTO: Phase 3.0 falsification prototype — Candidate B (generic registry).
//
// This file is part of the Phase 3.0 prototype comparing two shapes for a
// menu/page provider registry (key -> parameterized factory). Do NOT build on
// this code; it exists only as evidence for the A/B adjudication.
//
// Candidate B: providers are registered with a typed parameter object
//
//	RegisterMenuB(key, func(base baseMenu, opts T) (Menu, error))
//
// so call sites pass a struct literal whose fields are checked at compile time;
// the registry stores menuFactory[T] behind an `any` and BuildMenuB[T] recovers
// it with a single runtime type assertion (the ONLY assertion, hidden inside
// the registry, not at call sites or in providers).
package ui

import (
	"fmt"

	"github.com/anhoder/foxful-cli/model"

	"github.com/go-musicfox/go-musicfox/internal/structs"
)

// Parameter-object types for the 5 sample menus (Candidate B). These become the
// discoverable "contract" of each menu provider.
type (
	PlaylistDetailOpts struct{ PlaylistID int64 }
	ArtistDetailOpts   struct {
		ArtistID int64
		Name     string
	}
	SearchResultOpts      struct{ SearchType SearchType }
	AddToUserPlaylistOpts struct {
		UserID int64
		Song   structs.Song
		IsAdd  bool
	}
	NoArgMenuOpts struct{} // placeholder for no-arg menus (e.g. Ranks)
	LoginPageOpts struct{ Netease *Netease }
)

// menuFactory[T] is a typed menu provider stored behind `any` in the registry.
type menuFactory[T any] struct {
	Key   string
	Build func(base baseMenu, opts T) (Menu, error)
}

// menuRegistryB is the prototype registry for Candidate B.
var menuRegistryB = map[string]any{}

// RegisterMenuB registers a typed menu provider under key (panics on
// programmer error: duplicate key). T must be identical (not merely assignable)
// to the T used at the BuildMenuB call site.
func RegisterMenuB[T any](key string, f func(base baseMenu, opts T) (Menu, error)) {
	if key == "" || f == nil {
		panic("RegisterMenuB: empty key or nil factory")
	}
	if _, dup := menuRegistryB[key]; dup {
		panic("RegisterMenuB: duplicate key " + key)
	}
	menuRegistryB[key] = menuFactory[T]{Key: key, Build: f}
}

// BuildMenuB resolves the typed factory for key and invokes it with opts.
// The type-parameter T is normally inferred from the opts argument; a mismatch
// with the registered T is a runtime error here rather than a compile error.
func BuildMenuB[T any](key string, base baseMenu, opts T) (Menu, error) {
	factory, ok := menuRegistryB[key].(menuFactory[T])
	if !ok {
		return nil, fmt.Errorf("menu %q not registered or opts type mismatch (want %T)", key, opts)
	}
	return factory.Build(base, opts)
}

// mustBuildMenuB is the bootstrap-path escape hatch (see mustBuildMenuA).
func mustBuildMenuB[T any](key string, base baseMenu, opts T) Menu {
	menu, err := BuildMenuB(key, base, opts)
	if err != nil {
		panic(fmt.Sprintf("mustBuildMenuB(%q): %v", key, err))
	}
	return menu
}

// pageFactory[T] is a typed page provider stored behind `any` in the registry.
type pageFactory[T any] struct {
	Key   string
	Build func(opts T) (model.Page, error)
}

// pageRegistryB is the prototype page registry for Candidate B.
var pageRegistryB = map[string]any{}

// RegisterPageB registers a typed page provider under key.
func RegisterPageB[T any](key string, f func(opts T) (model.Page, error)) {
	if key == "" || f == nil {
		panic("RegisterPageB: empty key or nil factory")
	}
	if _, dup := pageRegistryB[key]; dup {
		panic("RegisterPageB: duplicate key " + key)
	}
	pageRegistryB[key] = pageFactory[T]{Key: key, Build: f}
}

// BuildPageB resolves the typed page factory for key and invokes it with opts.
func BuildPageB[T any](key string, opts T) (model.Page, error) {
	factory, ok := pageRegistryB[key].(pageFactory[T])
	if !ok {
		return nil, fmt.Errorf("page %q not registered or opts type mismatch (want %T)", key, opts)
	}
	return factory.Build(opts)
}

func init() {
	RegisterMenuB("playlist_detail", func(base baseMenu, opts PlaylistDetailOpts) (Menu, error) {
		return NewPlaylistDetailMenu(base, opts.PlaylistID), nil
	})

	RegisterMenuB("artist_detail", func(base baseMenu, opts ArtistDetailOpts) (Menu, error) {
		return NewArtistDetailMenu(base, opts.ArtistID, opts.Name), nil
	})

	RegisterMenuB("search_result", func(base baseMenu, opts SearchResultOpts) (Menu, error) {
		return NewSearchResultMenu(base, opts.SearchType), nil
	})

	RegisterMenuB("add_to_user_playlist", func(base baseMenu, opts AddToUserPlaylistOpts) (Menu, error) {
		return NewAddToUserPlaylistMenu(base, opts.UserID, opts.Song, opts.IsAdd), nil
	})

	// No-arg menu: needs a placeholder opts type (NoArgMenuOpts) so the generic
	// signature still holds; the call site must spell NoArgMenuOpts{} explicitly.
	RegisterMenuB("ranks", func(base baseMenu, _ NoArgMenuOpts) (Menu, error) {
		return NewRanksMenu(base), nil
	})

	RegisterPageB("login", func(opts LoginPageOpts) (model.Page, error) {
		return NewLoginPage(opts.Netease), nil
	})
}
