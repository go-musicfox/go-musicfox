package ui

import (
	"reflect"
	"testing"
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
