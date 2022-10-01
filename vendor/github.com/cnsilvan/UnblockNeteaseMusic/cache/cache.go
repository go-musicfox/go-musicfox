package cache

import (
	"fmt"
	"github.com/cnsilvan/UnblockNeteaseMusic/common"
	"strconv"
	"sync"
)

var cache sync.Map

func GetPlatFormIdTag(key common.PlatformIdTag) string {
	var id int64 = 100001
	k := string(key)
	if value, ok := cache.LoadOrStore("PlatFormIdTagFor"+k, id+1); ok {
		id = value.(int64)
		cache.Store("PlatFormIdTagFor"+k, id+1)
	}

	return strconv.FormatInt(id, 10)
}
func PutSong(key common.SearchMusic, value *common.Song) {
	cache.Store(fmt.Sprintf("%+v", key), value)
}
func GetSong(key common.SearchMusic) (*common.Song, bool) {
	var song *common.Song
	if value, ok := cache.Load(fmt.Sprintf("%+v", key)); ok {
		song, ok = value.(*common.Song)
		return song, ok
	}
	return song, false
}
func Delete(key common.SearchMusic) {
	cache.Delete(fmt.Sprintf("%+v", key))
}
