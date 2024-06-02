package netease

import (
	"strconv"
)

func WebUrlOfPlaylist(playlistId int64) string {
	return "https://music.163.com/#/my/m/music/playlist?id=" + strconv.FormatInt(playlistId, 10)
}

func WebUrlOfSong(songId int64) string {
	return "https://music.163.com/#/song?id=" + strconv.FormatInt(songId, 10)
}

func WebUrlOfArtist(artistId int64) string {
	return "https://music.163.com/#/artist?id=" + strconv.FormatInt(artistId, 10)
}

func WebUrlOfAlbum(artistId int64) string {
	return "https://music.163.com/#/album?id=" + strconv.FormatInt(artistId, 10)
}
