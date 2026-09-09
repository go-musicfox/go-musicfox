// Package ui implements musicfox's terminal user interface.
package ui

import (
	"fmt"
	"log/slog"
	"strconv"
	"sync"

	"github.com/anhoder/foxful-cli/model"
	"github.com/buger/jsonparser"
	"github.com/go-musicfox/netease-music/service"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/utils/errorx"
	"github.com/go-musicfox/go-musicfox/utils/slogx"
	_struct "github.com/go-musicfox/go-musicfox/utils/struct"
)

const albumSubscriptionPageSize = 200

type albumSubscriptionState struct {
	mu         sync.RWMutex
	ready      bool
	subscribed map[int64]struct{}
	known      map[int64]struct{}
	client     albumSubscriptionClient
}

func newAlbumSubscriptionState() *albumSubscriptionState {
	return &albumSubscriptionState{
		subscribed: make(map[int64]struct{}),
		known:      make(map[int64]struct{}),
		client:     neteaseAlbumSubscriptionClient{},
	}
}

type albumSubscriptionClient interface {
	list(limit, offset int) (float64, []byte)
	dynamic(albumID int64) (float64, []byte)
	update(albumID int64, subscribe bool) (float64, []byte)
}

type neteaseAlbumSubscriptionClient struct{}

func (neteaseAlbumSubscriptionClient) list(limit, offset int) (float64, []byte) {
	return (&service.AlbumSublistService{
		Limit:  strconv.Itoa(limit),
		Offset: strconv.Itoa(offset),
	}).AlbumSublist()
}

func (neteaseAlbumSubscriptionClient) dynamic(albumID int64) (float64, []byte) {
	return (&service.AlbumDetailDynamicService{ID: strconv.FormatInt(albumID, 10)}).AlbumDetailDynamic()
}

func (neteaseAlbumSubscriptionClient) update(albumID int64, subscribe bool) (float64, []byte) {
	t := "2"
	if subscribe {
		t = "1"
	}
	return (&service.AlbumSubService{ID: strconv.FormatInt(albumID, 10), T: t}).AlbumSub()
}

func (s *albumSubscriptionState) get(albumID int64) (subscribed, known bool) {
	if s == nil || albumID == 0 {
		return false, false
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, subscribed = s.subscribed[albumID]
	_, individuallyKnown := s.known[albumID]
	return subscribed, s.ready || individuallyKnown
}

func (s *albumSubscriptionState) replace(albumIDs map[int64]struct{}) {
	if s == nil {
		return
	}
	s.mu.Lock()
	for albumID := range s.known {
		if _, subscribed := s.subscribed[albumID]; subscribed {
			albumIDs[albumID] = struct{}{}
		} else {
			delete(albumIDs, albumID)
		}
	}
	s.subscribed = albumIDs
	s.ready = true
	s.mu.Unlock()
}

func (s *albumSubscriptionState) set(albumID int64, subscribed bool) {
	if s == nil || albumID == 0 {
		return
	}
	s.mu.Lock()
	if subscribed {
		s.subscribed[albumID] = struct{}{}
	} else {
		delete(s.subscribed, albumID)
	}
	s.known[albumID] = struct{}{}
	s.mu.Unlock()
}

type albumSubscriptionPageFetcher func(limit, offset int) (float64, []byte)

func fetchAlbumSubscriptions(fetch albumSubscriptionPageFetcher) (map[int64]struct{}, error) {
	result := make(map[int64]struct{})
	offset := 0
	for {
		code, response := fetch(albumSubscriptionPageSize, offset)
		if _struct.CheckCode(code) != _struct.Success {
			return nil, fmt.Errorf("album subscription list returned code %.0f", code)
		}

		albums := _struct.GetAlbumsSublist(response)
		for _, album := range albums {
			result[album.Id] = struct{}{}
		}

		hasMore, err := jsonparser.GetBoolean(response, "hasMore")
		if err != nil || !hasMore {
			break
		}
		if len(albums) == 0 {
			return nil, fmt.Errorf("album subscription list hasMore without data at offset %d", offset)
		}
		offset += len(albums)
	}
	return result, nil
}

func (n *Netease) refreshAlbumSubscriptions() {
	if n == nil || n.albumSubscriptions == nil || n.user == nil {
		return
	}
	errorx.Go(func() {
		albumIDs, err := fetchAlbumSubscriptions(n.albumSubscriptions.client.list)
		if err != nil {
			slog.Warn("Failed to refresh album subscription state", slogx.Error(err))
			return
		}
		n.albumSubscriptions.replace(albumIDs)
		n.Rerender(false)
	}, true)
}

func albumSubscriptionFromDynamic(response []byte) (bool, error) {
	isSub, err := jsonparser.GetBoolean(response, "isSub")
	if err != nil {
		return false, fmt.Errorf("parse album subscription state: %w", err)
	}
	return isSub, nil
}

func targetAlbum(n *Netease, isSelected bool, selectedIndex int) (structs.Album, bool) {
	if n == nil || n.player == nil {
		return structs.Album{}, false
	}
	if !isSelected {
		song, ok := getTargetSong(n, false)
		return song.Album, ok && song.Album.Id != 0
	}

	menu := n.MustMain().CurMenu()
	return selectedAlbum(menu, selectedIndex)
}

func selectedAlbum(menu model.Menu, selectedIndex int) (structs.Album, bool) {
	index := menu.RealDataIndex(selectedIndex)
	if albumsMenu, ok := menu.(AlbumsMenu); ok {
		albums := albumsMenu.Albums()
		if index >= 0 && index < len(albums) {
			return albums[index], albums[index].Id != 0
		}
	}
	songsMenu, ok := menu.(SongsMenu)
	if !ok {
		return structs.Album{}, false
	}
	if index < 0 || index >= len(songsMenu.Songs()) {
		return structs.Album{}, false
	}
	album := songsMenu.Songs()[index].Album
	return album, album.Id != 0
}
