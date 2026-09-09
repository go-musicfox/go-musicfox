package ui

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/anhoder/foxful-cli/model"

	"github.com/go-musicfox/go-musicfox/internal/structs"
)

func TestFetchAlbumSubscriptionsPaginates(t *testing.T) {
	var offsets []int
	got, err := fetchAlbumSubscriptions(func(limit, offset int) (float64, []byte) {
		if limit != albumSubscriptionPageSize {
			t.Fatalf("limit = %d, want %d", limit, albumSubscriptionPageSize)
		}
		offsets = append(offsets, offset)
		switch offset {
		case 0:
			return 200, []byte(`{"data":[{"id":11},{"id":22}],"hasMore":true}`)
		case 2:
			return 200, []byte(`{"data":[{"id":33}],"hasMore":false}`)
		default:
			t.Fatalf("unexpected offset %d", offset)
			return 0, nil
		}
	})
	if err != nil {
		t.Fatalf("fetchAlbumSubscriptions() error = %v", err)
	}
	if !reflect.DeepEqual(offsets, []int{0, 2}) {
		t.Fatalf("offsets = %v, want [0 2]", offsets)
	}
	for _, albumID := range []int64{11, 22, 33} {
		if _, ok := got[albumID]; !ok {
			t.Errorf("album %d missing from result", albumID)
		}
	}
}

func TestAlbumSubscriptionStateTracksKnownAlbums(t *testing.T) {
	state := newAlbumSubscriptionState()
	if _, known := state.get(11); known {
		t.Fatal("new state unexpectedly knows album 11")
	}

	state.set(11, true)
	if subscribed, known := state.get(11); !known || !subscribed {
		t.Fatalf("album 11 state = (%v, %v), want subscribed and known", subscribed, known)
	}

	state.replace(map[int64]struct{}{22: {}})
	if subscribed, known := state.get(11); !known || !subscribed {
		t.Fatalf("refresh overwrote newer album 11 state = (%v, %v)", subscribed, known)
	}
	if subscribed, known := state.get(33); !known || subscribed {
		t.Fatalf("ready cache state for album 33 = (%v, %v), want unsubscribed and known", subscribed, known)
	}
}

func TestAlbumSubscriptionFromDynamic(t *testing.T) {
	for _, tc := range []struct {
		name string
		body string
		want bool
	}{
		{name: "subscribed", body: `{"code":200,"isSub":true}`, want: true},
		{name: "not subscribed", body: `{"code":200,"isSub":false}`, want: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := albumSubscriptionFromDynamic([]byte(tc.body))
			if err != nil {
				t.Fatalf("albumSubscriptionFromDynamic() error = %v", err)
			}
			if got != tc.want {
				t.Fatalf("albumSubscriptionFromDynamic() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestAlbumSearchResultUsesAlbumTarget(t *testing.T) {
	menu := &SearchResultMenu{
		searchType: StAlbum,
		result: []structs.Album{
			{Id: 3210056, Name: "HEAR YOU"},
		},
	}

	if isSongsProvider(menu) {
		t.Fatal("album search result was classified as a song provider")
	}
	if !isAlbumsProvider(menu) {
		t.Fatal("album search result was not classified as an album provider")
	}
	album, ok := selectedAlbum(menu, 0)
	if !ok || album.Id != 3210056 || album.Name != "HEAR YOU" {
		t.Fatalf("selectedAlbum() = (%+v, %v), want HEAR YOU album", album, ok)
	}
}

// Exercise the menu callback and deferred write path without real account requests.
func TestAlbumSubscriptionActionPreservesIntent(t *testing.T) {
	for _, tc := range []struct {
		name                            string
		known, cached, current, desired bool
	}{
		{name: "unknown already subscribed", current: true, desired: true},
		{name: "unknown not subscribed", desired: true},
		{name: "known not subscribed", known: true, desired: true},
		{name: "known subscribed", known: true, cached: true, current: true},
		{name: "stale unsubscribed cache", known: true, current: true, desired: true},
		{name: "stale subscribed cache", known: true, cached: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			app, n := newFormPageTestApp(t)
			n.player = &Player{}
			n.user = &structs.User{UserId: 1}
			n.albumSubscriptions = newAlbumSubscriptionState()
			client := &intentAlbumSubscriptionClient{current: tc.current}
			n.albumSubscriptions.client = client
			if tc.known {
				n.albumSubscriptions.set(3210056, tc.cached)
			}
			app.MustMain().EnterMenu(&intentAlbumMenu{}, &model.MenuItem{Title: "Albums"})
			action := buildAlbumSubscriptionAction(n, true, 0)
			wantTitle := iconHeartFilled + "收藏专辑"
			if !tc.desired {
				wantTitle = iconHeartOutline + "取消收藏专辑"
			}
			if action.title.Title != wantTitle {
				t.Errorf("title = %q, want %q", action.title.Title, wantTitle)
			}
			action.page()
			app.MustMain().Update(model.TickMain(0)(), app)
			if client.reads != 1 {
				t.Fatalf("dynamic requests = %d, want 1", client.reads)
			}
			var wantWrites []bool
			if tc.current != tc.desired {
				wantWrites = []bool{tc.desired}
			}
			if !reflect.DeepEqual(client.writes, wantWrites) {
				t.Errorf("writes = %v, want %v", client.writes, wantWrites)
			}
			if got, known := n.albumSubscriptions.get(3210056); !known || got != tc.desired {
				t.Errorf("cached state = (%v, %v), want (%v, true)", got, known, tc.desired)
			}
		})
	}
}

type intentAlbumMenu struct{ pageLayoutTestMenu }

func (*intentAlbumMenu) Albums() []structs.Album {
	return []structs.Album{{Id: 3210056, Name: "HEAR YOU"}}
}

type intentAlbumSubscriptionClient struct {
	albumSubscriptionClient
	current bool
	reads   int
	writes  []bool
}

func (c *intentAlbumSubscriptionClient) dynamic(albumID int64) (float64, []byte) {
	c.reads++
	if albumID != 3210056 {
		return 400, nil
	}
	return 200, fmt.Appendf(nil, `{"isSub":%t}`, c.current)
}
func (c *intentAlbumSubscriptionClient) update(albumID int64, subscribed bool) (float64, []byte) {
	c.writes = append(c.writes, subscribed)
	if albumID != 3210056 {
		return 400, nil
	}
	return 200, []byte(`{"code":200}`)
}

func (*intentAlbumMenu) IsPlayable() bool  { return false }
func (*intentAlbumMenu) IsLocatable() bool { return false }
