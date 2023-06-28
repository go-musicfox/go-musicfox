package like_list

import (
	"strconv"

	"github.com/buger/jsonparser"
	"github.com/go-musicfox/netease-music/service"
)

type LikeList map[int64]struct{}

var likeList = make(LikeList)

func IsLikeSong(songId int64) bool {
	_, ok := likeList[songId]
	return ok
}

func RefreshLikeList(userId int64) {
	s := &service.LikeListService{UID: strconv.FormatInt(userId, 10)}
	_, resp := s.LikeList()

	likeList = make(LikeList)
	_, _ = jsonparser.ArrayEach(resp, func(value []byte, _ jsonparser.ValueType, _ int, err error) {
		if err != nil {
			return
		}
		if id, err := jsonparser.ParseInt(value); err == nil {
			likeList[id] = struct{}{}
		}
	}, "ids")
}
