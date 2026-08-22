// PROTO: Phase 3.0 falsification prototype — Candidate A (variadic registry).
//
// This file is part of the Phase 3.0 prototype comparing two shapes for a
// menu/page provider registry (key -> parameterized factory). Do NOT build on
// this code; it exists only as evidence for the A/B adjudication.
//
// Candidate A: providers are registered with an untyped variadic signature
//
//	Build(base baseMenu, args ...any) (Menu, error)
//
// so call sites pass positional arguments (mirroring today's constructors) and
// each provider closure is responsible for asserting arg count/type/order.
package ui

import (
	"fmt"

	"github.com/anhoder/foxful-cli/model"

	"github.com/go-musicfox/go-musicfox/internal/structs"
)

// MenuProviderA is a key -> parameterized menu factory (variadic shape).
type MenuProviderA struct {
	Key   string
	Build func(base baseMenu, args ...any) (Menu, error)
}

// menuRegistryA is the prototype registry for Candidate A.
var menuRegistryA = map[string]*MenuProviderA{}

// RegisterMenuA registers a menu provider under Key (panics on programmer
// error: nil provider or duplicate/empty key).
func RegisterMenuA(p *MenuProviderA) {
	if p == nil || p.Key == "" {
		panic("RegisterMenuA: nil provider or empty key")
	}
	if _, dup := menuRegistryA[p.Key]; dup {
		panic("RegisterMenuA: duplicate key " + p.Key)
	}
	menuRegistryA[p.Key] = p
}

// BuildMenuA resolves a provider by key and invokes it with the given args.
// Arg type/count/order mistakes surface here as runtime errors, not compile
// errors (the provider's assertions are the only guard).
func BuildMenuA(key string, base baseMenu, args ...any) (Menu, error) {
	p, ok := menuRegistryA[key]
	if !ok {
		return nil, fmt.Errorf("menu %q not registered", key)
	}
	return p.Build(base, args...)
}

// mustBuildMenuA is the bootstrap-path escape hatch: registration errors with
// a static, in-code registry are programmer errors, so panic instead of
// threading an error through a constructor that has no error channel.
func mustBuildMenuA(key string, base baseMenu, args ...any) Menu {
	menu, err := BuildMenuA(key, base, args...)
	if err != nil {
		panic(fmt.Sprintf("mustBuildMenuA(%q): %v", key, err))
	}
	return menu
}

// PageProviderA is a key -> parameterized page factory (variadic shape).
type PageProviderA struct {
	Key   string
	Build func(args ...any) (model.Page, error)
}

// pageRegistryA is the prototype page registry for Candidate A.
var pageRegistryA = map[string]*PageProviderA{}

// RegisterPageA registers a page provider under Key.
func RegisterPageA(p *PageProviderA) {
	if p == nil || p.Key == "" {
		panic("RegisterPageA: nil provider or empty key")
	}
	if _, dup := pageRegistryA[p.Key]; dup {
		panic("RegisterPageA: duplicate key " + p.Key)
	}
	pageRegistryA[p.Key] = p
}

// BuildPageA resolves a page provider by key and invokes it with the given args.
func BuildPageA(key string, args ...any) (model.Page, error) {
	p, ok := pageRegistryA[key]
	if !ok {
		return nil, fmt.Errorf("page %q not registered", key)
	}
	return p.Build(args...)
}

func init() {
	RegisterMenuA(&MenuProviderA{
		Key: "playlist_detail",
		Build: func(base baseMenu, args ...any) (Menu, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("playlist_detail: want 1 arg (playlistId int64), got %d", len(args))
			}
			playlistId, ok := args[0].(int64)
			if !ok {
				return nil, fmt.Errorf("playlist_detail: args[0] must be int64, got %T", args[0])
			}
			return NewPlaylistDetailMenu(base, playlistId), nil
		},
	})

	RegisterMenuA(&MenuProviderA{
		Key: "artist_detail",
		Build: func(base baseMenu, args ...any) (Menu, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("artist_detail: want 2 args (artistId int64, name string), got %d", len(args))
			}
			artistId, ok := args[0].(int64)
			if !ok {
				return nil, fmt.Errorf("artist_detail: args[0] must be int64, got %T", args[0])
			}
			name, ok := args[1].(string)
			if !ok {
				return nil, fmt.Errorf("artist_detail: args[1] must be string, got %T", args[1])
			}
			return NewArtistDetailMenu(base, artistId, name), nil
		},
	})

	RegisterMenuA(&MenuProviderA{
		Key: "search_result",
		Build: func(base baseMenu, args ...any) (Menu, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("search_result: want 1 arg (searchType SearchType), got %d", len(args))
			}
			searchType, ok := args[0].(SearchType)
			if !ok {
				return nil, fmt.Errorf("search_result: args[0] must be SearchType, got %T", args[0])
			}
			return NewSearchResultMenu(base, searchType), nil
		},
	})

	RegisterMenuA(&MenuProviderA{
		Key: "add_to_user_playlist",
		Build: func(base baseMenu, args ...any) (Menu, error) {
			if len(args) != 3 {
				return nil, fmt.Errorf("add_to_user_playlist: want 3 args (userId int64, song structs.Song, isAdd bool), got %d", len(args))
			}
			userId, ok := args[0].(int64)
			if !ok {
				return nil, fmt.Errorf("add_to_user_playlist: args[0] must be int64, got %T", args[0])
			}
			song, ok := args[1].(structs.Song)
			if !ok {
				return nil, fmt.Errorf("add_to_user_playlist: args[1] must be structs.Song, got %T", args[1])
			}
			isAdd, ok := args[2].(bool)
			if !ok {
				return nil, fmt.Errorf("add_to_user_playlist: args[2] must be bool, got %T", args[2])
			}
			return NewAddToUserPlaylistMenu(base, userId, song, isAdd), nil
		},
	})

	// No-arg menu: provider takes zero args; the variadic shape needs no
	// placeholder type, the call site is just BuildMenuA("ranks", base).
	RegisterMenuA(&MenuProviderA{
		Key: "ranks",
		Build: func(base baseMenu, args ...any) (Menu, error) {
			if len(args) != 0 {
				return nil, fmt.Errorf("ranks: want 0 args, got %d", len(args))
			}
			return NewRanksMenu(base), nil
		},
	})

	RegisterPageA(&PageProviderA{
		Key: "login",
		Build: func(args ...any) (model.Page, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("login: want 1 arg (netease *Netease), got %d", len(args))
			}
			netease, ok := args[0].(*Netease)
			if !ok {
				return nil, fmt.Errorf("login: args[0] must be *Netease, got %T", args[0])
			}
			return NewLoginPage(netease), nil
		},
	})
}
