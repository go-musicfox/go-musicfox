package netease

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	_struct "github.com/go-musicfox/go-musicfox/utils/struct"
)

func stubPlaylistRequests(
	t *testing.T,
	detailFn func(playlistId int64) (float64, []byte),
	songDetailFn func(ids string) (float64, []byte),
) {
	t.Helper()
	prevDetail := requestPlaylistDetail
	prevSongDetail := requestSongDetail
	requestPlaylistDetail = detailFn
	if songDetailFn != nil {
		requestSongDetail = songDetailFn
	}
	t.Cleanup(func() {
		requestPlaylistDetail = prevDetail
		requestSongDetail = prevSongDetail
	})
}

func TestFetchSongsOfPlaylistRejectsHTML502FallbackCode(t *testing.T) {
	stubPlaylistRequests(t, func(playlistId int64) (float64, []byte) {
		// util.CreateRequest falls back to 200 when JSON "code" cannot be parsed.
		return 200, []byte("<html><head><title>502 Bad Gateway</title></head><body>502</body></html>")
	}, nil)

	codeType, songs := FetchSongsOfPlaylist(12345, false)
	if codeType != _struct.NetworkError {
		t.Fatalf("expected NetworkError for 502 HTML response, got %v", codeType)
	}
	if len(songs) != 0 {
		t.Fatalf("expected no songs on 502 HTML response, got %d", len(songs))
	}
}

func TestFetchSongsOfPlaylistRejectsMissingCodeAndMissingPlaylist(t *testing.T) {
	t.Run("missing code field", func(t *testing.T) {
		stubPlaylistRequests(t, func(playlistId int64) (float64, []byte) {
			return 200, []byte(`{"message":"error"}`)
		}, nil)

		codeType, songs := FetchSongsOfPlaylist(12345, false)
		if codeType != _struct.NetworkError {
			t.Fatalf("expected NetworkError when code field is missing, got %v", codeType)
		}
		if len(songs) != 0 {
			t.Fatalf("expected no songs, got %d", len(songs))
		}
	})

	t.Run("missing playlist object", func(t *testing.T) {
		stubPlaylistRequests(t, func(playlistId int64) (float64, []byte) {
			return 200, []byte(`{"code":200,"playlist":null}`)
		}, nil)

		codeType, songs := FetchSongsOfPlaylist(12345, false)
		if codeType != _struct.UnknownError {
			t.Fatalf("expected UnknownError when playlist is null, got %v", codeType)
		}
		if len(songs) != 0 {
			t.Fatalf("expected no songs, got %d", len(songs))
		}
	})

	t.Run("need login", func(t *testing.T) {
		stubPlaylistRequests(t, func(playlistId int64) (float64, []byte) {
			return 301, []byte(`{"code":301}`)
		}, nil)

		codeType, _ := FetchSongsOfPlaylist(12345, false)
		if codeType != _struct.NeedLogin {
			t.Fatalf("expected NeedLogin, got %v", codeType)
		}
	})
}

func TestFetchSongsOfPlaylistReturnsDirectTracksWithoutExtraRequests(t *testing.T) {
	songDetailCalls := 0
	stubPlaylistRequests(t, func(playlistId int64) (float64, []byte) {
		return 200, []byte(`{
			"code": 200,
			"playlist": {
				"tracks": [
					{"id": 101, "name": "Song A", "dt": 180000, "al": {"id": 1, "name": "Album A"}, "ar": [{"id": 10, "name": "Artist A"}]},
					{"id": 102, "name": "Song B", "dt": 200000, "al": {"id": 2, "name": "Album B"}, "ar": [{"id": 20, "name": "Artist B"}]}
				],
				"trackIds": [{"id": 101}, {"id": 102}]
			}
		}`)
	}, func(ids string) (float64, []byte) {
		songDetailCalls++
		return 520, nil
	})

	for _, getAll := range []bool{false, true} {
		codeType, songs := FetchSongsOfPlaylist(12345, getAll)
		if codeType != _struct.Success {
			t.Fatalf("getAll=%v: expected Success, got %v", getAll, codeType)
		}
		if len(songs) != 2 || songs[0].Id != 101 || songs[1].Id != 102 {
			t.Fatalf("getAll=%v: unexpected songs %+v", getAll, songs)
		}
	}
	if songDetailCalls != 0 {
		t.Fatalf("expected 0 song detail calls when tracks are complete, got %d", songDetailCalls)
	}
}

func TestFetchSongsOfPlaylistCompletesMissingTrackDetailsByIDs(t *testing.T) {
	var requestedBatches []string
	stubPlaylistRequests(t, func(playlistId int64) (float64, []byte) {
		return 200, []byte(`{
			"code": 200,
			"playlist": {
				"tracks": [],
				"trackIds": [{"id": 303}, {"id": 101}, {"id": 202}]
			}
		}`)
	}, func(ids string) (float64, []byte) {
		requestedBatches = append(requestedBatches, ids)
		// Intentionally return songs out of order to verify ordering by trackIds.
		return 200, []byte(`{
			"code": 200,
			"songs": [
				{"id": 101, "name": "Song 101", "dt": 1000, "al": {"id": 1, "name": "Al 1"}, "ar": [{"id": 1, "name": "Ar 1"}]},
				{"id": 202, "name": "Song 202", "dt": 2000, "al": {"id": 2, "name": "Al 2"}, "ar": [{"id": 2, "name": "Ar 2"}]},
				{"id": 303, "name": "Song 303", "dt": 3000, "al": {"id": 3, "name": "Al 3"}, "ar": [{"id": 3, "name": "Ar 3"}]}
			]
		}`)
	})

	codeType, songs := FetchSongsOfPlaylist(12345, false)
	if codeType != _struct.Success {
		t.Fatalf("expected Success, got %v", codeType)
	}
	if len(requestedBatches) != 1 || requestedBatches[0] != "303,101,202" {
		t.Fatalf("unexpected requested batches: %v", requestedBatches)
	}
	if len(songs) != 3 || songs[0].Id != 303 || songs[1].Id != 101 || songs[2].Id != 202 {
		t.Fatalf("expected songs ordered [303, 101, 202], got %+v", songs)
	}
}

func TestFetchSongsOfPlaylistCompletesPartialTracksWhenGetAll(t *testing.T) {
	var requestedBatches []string
	stubPlaylistRequests(t, func(playlistId int64) (float64, []byte) {
		return 200, []byte(`{
			"code": 200,
			"playlist": {
				"tracks": [
					{"id": 101, "name": "Song 101", "dt": 1000, "al": {"id": 1, "name": "Al 1"}, "ar": [{"id": 1, "name": "Ar 1"}]}
				],
				"trackIds": [{"id": 101}, {"id": 202}, {"id": 303}]
			}
		}`)
	}, func(ids string) (float64, []byte) {
		requestedBatches = append(requestedBatches, ids)
		return 200, []byte(`{
			"code": 200,
			"songs": [
				{"id": 202, "name": "Song 202", "dt": 2000},
				{"id": 303, "name": "Song 303", "dt": 3000}
			]
		}`)
	})

	codeType, songs := FetchSongsOfPlaylist(12345, true)
	if codeType != _struct.Success {
		t.Fatalf("expected Success, got %v", codeType)
	}
	if len(requestedBatches) != 1 || requestedBatches[0] != "202,303" {
		t.Fatalf("expected only missing IDs [202,303] to be requested, got %v", requestedBatches)
	}
	if len(songs) != 3 || songs[0].Id != 101 || songs[1].Id != 202 || songs[2].Id != 303 {
		t.Fatalf("unexpected merged songs: %+v", songs)
	}
}

func TestFetchSongsOfPlaylistBatchesOver500TrackIDs(t *testing.T) {
	const total = 505
	var idEntries []string
	for i := 1; i <= total; i++ {
		idEntries = append(idEntries, fmt.Sprintf(`{"id":%d}`, i))
	}
	playlistResp := []byte(`{"code":200,"playlist":{"tracks":[],"trackIds":[` + strings.Join(idEntries, ",") + `]}}`)

	var batchSizes []int
	stubPlaylistRequests(t, func(playlistId int64) (float64, []byte) {
		return 200, playlistResp
	}, func(ids string) (float64, []byte) {
		parts := strings.Split(ids, ",")
		batchSizes = append(batchSizes, len(parts))
		var songEntries []string
		for _, p := range parts {
			if p == "" {
				t.Fatalf("unexpected empty ID in batch %q", ids)
			}
			id, err := strconv.Atoi(p)
			if err != nil {
				t.Fatalf("invalid ID %q: %v", p, err)
			}
			songEntries = append(songEntries, fmt.Sprintf(`{"id":%d,"name":"Song %d"}`, id, id))
		}
		return 200, []byte(`{"code":200,"songs":[` + strings.Join(songEntries, ",") + `]}`)
	})

	codeType, songs := FetchSongsOfPlaylist(12345, false)
	if codeType != _struct.Success {
		t.Fatalf("expected Success, got %v", codeType)
	}
	if len(batchSizes) != 2 || batchSizes[0] != 500 || batchSizes[1] != 5 {
		t.Fatalf("expected batch sizes [500, 5], got %v", batchSizes)
	}
	if len(songs) != total || songs[0].Id != 1 || songs[total-1].Id != total {
		t.Fatalf("expected %d ordered songs, got %d", total, len(songs))
	}
}

func TestFetchSongsOfPlaylistFailsWhenTrackCompletionFails(t *testing.T) {
	stubPlaylistRequests(t, func(playlistId int64) (float64, []byte) {
		return 200, []byte(`{"code":200,"playlist":{"tracks":[],"trackIds":[{"id":101}]}}`)
	}, func(ids string) (float64, []byte) {
		return 200, []byte("<html>502 Bad Gateway</html>")
	})

	codeType, songs := FetchSongsOfPlaylist(12345, false)
	if codeType != _struct.NetworkError {
		t.Fatalf("expected NetworkError when song detail completion returns 502 HTML, got %v", codeType)
	}
	if len(songs) != 0 {
		t.Fatalf("expected no songs on completion failure, got %d", len(songs))
	}
}

func TestFetchSongsOfPlaylistAllowsGenuineEmptyPlaylist(t *testing.T) {
	stubPlaylistRequests(t, func(playlistId int64) (float64, []byte) {
		return 200, []byte(`{"code":200,"playlist":{"tracks":[],"trackIds":[]}}`)
	}, nil)

	codeType, songs := FetchSongsOfPlaylist(12345, false)
	if codeType != _struct.Success {
		t.Fatalf("expected Success for genuine empty playlist, got %v", codeType)
	}
	if len(songs) != 0 {
		t.Fatalf("expected 0 songs, got %d", len(songs))
	}
}
