package utils

import (
	"strings"

	"github.com/go-musicfox/go-musicfox/pkg/structs"

	"github.com/buger/jsonparser"
)

type ResCode uint8

const (
	Success ResCode = iota
	UnknownError
	NetworkError
	NeedLogin
	PasswordError
)

// CheckCode 验证响应码
func CheckCode(code float64) ResCode {
	switch code {
	case 301, 302, 20001:
		return NeedLogin
	case 520:
		return NetworkError
	case 200:
		return Success
	}

	return PasswordError
}

// CheckUserInfo 验证用户信息
func CheckUserInfo(user *structs.User) ResCode {
	if user == nil || user.UserId == 0 {
		return NeedLogin
	}

	return Success
}

// ReplaceSpecialStr 替换特殊字符
func ReplaceSpecialStr(str string) string {
	replaceStr := map[string]string{
		"“": "\"",
		"”": "\"",
		"·": ".",
	}
	for oldStr, newStr := range replaceStr {
		str = strings.ReplaceAll(str, oldStr, newStr)
	}

	return str
}

// GetDailySongs 获取每日歌曲列表
func GetDailySongs(data []byte) (list []structs.Song) {
	_, _ = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if song, err := structs.NewSongFromShortNameSongsJson(value); err == nil {
			list = append(list, song)
		}

	}, "data", "dailySongs")

	return
}

// GetRecentSongs 获取每日歌曲列表
func GetRecentSongs(data []byte) (list []structs.Song) {
	_, _ = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if t, _ := jsonparser.GetString(value, "resourceType"); t != "SONG" {
			return
		}

		var (
			dataJson []byte
			song     structs.Song
		)
		if dataJson, _, _, err = jsonparser.Get(value, "data"); err != nil {
			return
		}
		if song, err = structs.NewSongFromShortNameSongsJson(dataJson); err == nil {
			list = append(list, song)
		}
	}, "data", "list")
	return
}

// GetDailyPlaylists 获取每日推荐歌单
func GetDailyPlaylists(data []byte) (list []structs.Playlist) {

	_, _ = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if playlist, err := structs.NewPlaylistFromJson(value); err == nil {
			list = append(list, playlist)
		}
	}, "recommend")

	return
}

// GetSongsOfPlaylist 获取播放列表的歌曲
func GetSongsOfPlaylist(data []byte) (list []structs.Song) {
	_, _ = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if song, err := structs.NewSongFromShortNameSongsJson(value); err == nil {
			list = append(list, song)
		}

	}, "playlist", "tracks")

	return
}

// GetSongsOfAlbum 获取专辑的歌曲
func GetSongsOfAlbum(data []byte) (list []structs.Song) {
	_, _ = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if song, err := structs.NewSongFromAlbumSongsJson(value); err == nil {
			list = append(list, song)
		}

	}, "songs")

	return
}

// GetPlaylists 获取播放列表
func GetPlaylists(data []byte) (list []structs.Playlist) {

	_, _ = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if playlist, err := structs.NewPlaylistFromJson(value); err == nil {
			list = append(list, playlist)
		}
	}, "playlist")

	return
}

// GetPlaylistsFromHighQuality 获取精品歌单
func GetPlaylistsFromHighQuality(data []byte) (list []structs.Playlist) {

	_, _ = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if playlist, err := structs.NewPlaylistFromJson(value); err == nil {
			list = append(list, playlist)
		}
	}, "playlists")

	return
}

// GetFmSongs 获取每日歌曲列表
func GetFmSongs(data []byte) (list []structs.Song) {
	_, _ = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if song, err := structs.NewSongFromFmJson(value); err == nil {
			list = append(list, song)
		}

	}, "data")

	return
}

// GetIntelligenceSongs 获取心动模式歌曲列表
func GetIntelligenceSongs(data []byte) (list []structs.Song) {
	_, _ = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if song, err := structs.NewSongFromIntelligenceJson(value); err == nil {
			list = append(list, song)
		}

	}, "data")

	return
}

// GetNewAlbums 获取最新专辑列表
func GetNewAlbums(data []byte) (albums []structs.Album) {
	_, _ = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {

		if album, err := structs.NewAlbumFromAlbumJson(value); err == nil {
			albums = append(albums, album)
		}

	}, "albums")

	return
}

// GetTopAlbums 获取专辑列表
func GetTopAlbums(data []byte) (albums []structs.Album) {
	_, _ = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {

		if album, err := structs.NewAlbumFromAlbumJson(value); err == nil {
			albums = append(albums, album)
		}

	}, "monthData")

	return
}

// GetArtistHotAlbums 获取歌手热门专辑列表
func GetArtistHotAlbums(data []byte) (albums []structs.Album) {
	_, _ = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {

		if album, err := structs.NewAlbumFromAlbumJson(value); err == nil {
			albums = append(albums, album)
		}

	}, "hotAlbums")

	return
}

// GetSongsOfSearchResult 获取搜索结果的歌曲
func GetSongsOfSearchResult(data []byte) (list []structs.Song) {
	_, _ = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if song, err := structs.NewSongFromShortNameSongsJson(value); err == nil {
			list = append(list, song)
		}

	}, "result", "songs")

	return
}

// GetAlbumsOfSearchResult 获取搜索结果的专辑
func GetAlbumsOfSearchResult(data []byte) (list []structs.Album) {
	_, _ = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if album, err := structs.NewAlbumFromAlbumJson(value); err == nil {
			list = append(list, album)
		}

	}, "result", "albums")

	return
}

// GetPlaylistsOfSearchResult 获取搜索结果的歌单
func GetPlaylistsOfSearchResult(data []byte) (list []structs.Playlist) {
	_, _ = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if playlist, err := structs.NewPlaylistFromJson(value); err == nil {
			list = append(list, playlist)
		}

	}, "result", "playlists")

	return
}

// GetArtistsOfSearchResult 获取搜索结果的歌手
func GetArtistsOfSearchResult(data []byte) (list []structs.Artist) {
	_, _ = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if artist, err := structs.NewArtist(value); err == nil {
			list = append(list, artist)
		}

	}, "result", "artists")

	return
}

// GetArtistsOfTopArtists 获取热门歌手
func GetArtistsOfTopArtists(data []byte) (list []structs.Artist) {
	_, _ = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if artist, err := structs.NewArtist(value); err == nil {
			list = append(list, artist)
		}

	}, "artists")

	return
}

// GetSongsOfArtist 获取歌手的歌曲
func GetSongsOfArtist(data []byte) (list []structs.Song) {
	_, _ = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if song, err := structs.NewSongFromArtistSongsJson(value); err == nil {
			list = append(list, song)
		}

	}, "songs")

	return
}

// GetUsersOfSearchResult 从搜索结果中获取用户列表
func GetUsersOfSearchResult(data []byte) (list []structs.User) {
	_, _ = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if song, err := structs.NewUserFromSearchResultJson(value); err == nil {
			list = append(list, song)
		}

	}, "result", "userprofiles")

	return
}

// GetDjRadiosOfSearchResult 从搜索结果中获取电台列表
func GetDjRadiosOfSearchResult(data []byte) (list []structs.DjRadio) {
	_, _ = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if radio, err := structs.NewDjRadioFromJson(value); err == nil {
			list = append(list, radio)
		}

	}, "result", "djRadios")

	return
}

// GetDjRadios 获取电台列表
func GetDjRadios(data []byte) (list []structs.DjRadio) {
	_, _ = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if radio, err := structs.NewDjRadioFromJson(value); err == nil {
			list = append(list, radio)
		}

	}, "djRadios")

	return
}

// GetDjRadiosOfToday 获取今日优选电台列表
func GetDjRadiosOfToday(data []byte) (list []structs.DjRadio) {
	_, _ = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if radio, err := structs.NewDjRadioFromJson(value); err == nil {
			list = append(list, radio)
		}

	}, "data")

	return
}

// GetDjRadiosOfTopDj 获取热门电台列表
func GetDjRadiosOfTopDj(data []byte) (list []structs.DjRadio) {
	_, _ = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if radio, err := structs.NewDjRadioFromJson(value); err == nil {
			list = append(list, radio)
		}

	}, "toplist")

	return
}

// GetSongsOfDjRadio 获取电台节目列表的歌曲
func GetSongsOfDjRadio(data []byte) (list []structs.Song) {
	_, _ = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if song, err := structs.NewSongFromDjRadioProgramJson(value); err == nil {
			list = append(list, song)
		}
	}, "programs")

	return
}

// GetSongsOfDjRank 获取电台节目排行榜列表的歌曲
func GetSongsOfDjRank(data []byte) (list []structs.Song) {
	_, _ = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if song, err := structs.NewSongFromDjRankProgramJson(value); err == nil {
			list = append(list, song)
		}
	}, "toplist")

	return
}

// GetSongsOfDjHoursRank 获取电台节目24小时排行榜列表的歌曲
func GetSongsOfDjHoursRank(data []byte) (list []structs.Song) {
	_, _ = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if song, err := structs.NewSongFromDjRankProgramJson(value); err == nil {
			list = append(list, song)
		}
	}, "data", "list")

	return
}

// GetRanks 获取排行榜
func GetRanks(data []byte) (list []structs.Rank) {

	_, _ = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if rank, err := structs.NewRankFromJson(value); err == nil {
			list = append(list, rank)
		}
	}, "list")

	return
}

// GetSongsOfCloud 获取云盘的歌曲
func GetSongsOfCloud(data []byte) (list []structs.Song) {
	_, _ = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if song, err := structs.NewSongFromCloudJson(value); err == nil {
			list = append(list, song)
		}
	}, "data")

	return
}

// GetDjCategory 获取电台分类
func GetDjCategory(data []byte) (list []structs.DjCategory) {
	_, _ = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if cate, err := structs.NewDjCategoryFromJson(value); err == nil {
			list = append(list, cate)
		}

	}, "categories")

	return
}
