package service

import (
	"bytes"
	"github.com/buger/jsonparser"
	"math"
	"strconv"
	"strings"
	"sync"
)

type PlaylistTrackAllService struct {
	Id string `json:"id" form:"id"`
	S  string `json:"s" form:"s"`
}

func (service *PlaylistTrackAllService) AllTracks() (float64, []byte) {
	playlistDetailService := &PlaylistDetailService{
		Id: service.Id,
		S:  service.S,
	}
	code, reBody := playlistDetailService.PlaylistDetail()
	if code != 200 {
		return code, reBody
	}

	var trackIds []int64
	if _, err := jsonparser.ArrayEach(reBody, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if id, err := jsonparser.GetInt(value, "id"); err == nil {
			trackIds = append(trackIds, id)
		}
	}, "playlist", "trackIds"); err != nil {
		return code, reBody
	}

	count := len(trackIds)
	page := int(math.Ceil(float64(count) / 500))
	var (
		tracks = make([][]byte, page)
		wg     sync.WaitGroup
	)
	for i := 0; i < page; i++ {
		var b strings.Builder
		for j := 0; j < 500; j++ {
			index := i*500 + j
			if index >= count {
				break
			}
			b.WriteString(strconv.FormatInt(trackIds[index], 10))
			if j != 499 {
				b.WriteString(`,`)
			}
		}

		wg.Add(1)
		go func(wg *sync.WaitGroup, page int, ids string) {
			s := SongDetailService{Ids: ids}
			_, resp := s.SongDetail()
			var bf [][]byte
			_, _ = jsonparser.ArrayEach(resp, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
				bf = append(bf, value)
			}, "songs")
			tracks[page] = bytes.Join(bf, []byte(","))
			wg.Done()
		}(&wg, i, b.String())
	}
	wg.Wait()

	var bf = bytes.NewBufferString("[")
	for i, track := range tracks {
		if track == nil {
			continue
		}
		bf.Write(track)
		if i != page-1 {
			bf.WriteString(",")
		}
	}
	bf.WriteString("]")

	if r, err := jsonparser.Set(reBody, bf.Bytes(), "playlist", "tracks"); err == nil {
		reBody = r
	}

	return code, reBody
}
