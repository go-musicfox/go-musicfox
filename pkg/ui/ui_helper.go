package ui

import (
	"fmt"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/muesli/termenv"
	"go-musicfox/pkg/configs"
	"go-musicfox/pkg/constants"
	ds2 "go-musicfox/pkg/structs"
	"go-musicfox/utils"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

var (
	termProfile     = termenv.ColorProfile()
	primaryColor    termenv.Color
	primaryColorStr string
)

// GetPrimaryColor get random color
func GetPrimaryColor() termenv.Color {
	if primaryColor != nil {
		return primaryColor
	}

	if configs.ConfigRegistry.MainPrimaryColor == "" || configs.ConfigRegistry.MainPrimaryColor == constants.AppPrimaryRandom {
		rand.Seed(time.Now().UnixNano())
		primaryColorStr = strconv.Itoa(rand.Intn(228-17) + 17)
	} else {
		primaryColorStr = configs.ConfigRegistry.MainPrimaryColor
	}
	primaryColor = termProfile.Color(primaryColorStr)

	return primaryColor
}

// GetPrimaryColorStr get random color string
func GetPrimaryColorStr() string {
	if primaryColorStr != "" {
		return primaryColorStr
	}
	if configs.ConfigRegistry.MainPrimaryColor == "" || configs.ConfigRegistry.MainPrimaryColor == constants.AppPrimaryRandom {
		rand.Seed(time.Now().UnixNano())
		primaryColorStr = strconv.Itoa(rand.Intn(228-17) + 17)
	} else {
		primaryColorStr = configs.ConfigRegistry.MainPrimaryColor
	}
	primaryColor = termProfile.Color(primaryColorStr)

	return primaryColorStr
}

// GetRandomRgbColor get random rgb color
func GetRandomRgbColor(isRange bool) (string, string) {
	rand.Seed(time.Now().UnixNano())
	r := 255 - rand.Intn(100)
	rand.Seed(time.Now().UnixNano() / 2)
	g := 255 - rand.Intn(100)
	rand.Seed(time.Now().UnixNano() / 3)
	b := 255 - rand.Intn(100)

	startColor := fmt.Sprintf("#%x%x%x", r, g, b)
	if !isRange {
		return startColor, ""
	}

	rand.Seed(time.Now().UnixNano() / 5)
	rEnd := 50 + rand.Intn(100)
	rand.Seed(time.Now().UnixNano() / 7)
	gEnd := 50 + rand.Intn(100)
	rand.Seed(time.Now().UnixNano() / 11)
	bEnd := 50 + rand.Intn(100)
	endColor := fmt.Sprintf("#%x%x%x", rEnd, gEnd, bEnd)

	return startColor, endColor
}

// SetFgStyle Return a function that will colorize the foreground of a given string.
func SetFgStyle(content string, color termenv.Color) string {
	return termenv.Style{}.Foreground(color).Styled(content)
}

// SetFgBgStyle Color a string's foreground and background with the given value.
func SetFgBgStyle(content string, fg, bg termenv.Color) string {
	return termenv.Style{}.Foreground(fg).Background(bg).Styled(content)
}

// SetNormalStyle dont set any style
func SetNormalStyle(content string) string {
	seq := strings.Join([]string{"0"}, ";")
	return fmt.Sprintf("%s%sm%s%sm", termenv.CSI, seq, content, termenv.CSI+termenv.ResetSeq)
}

// Generate a blend of colors.
func makeRamp(colorA, colorB string, steps float64) (s []string) {
	cA, _ := colorful.Hex(colorA)
	cB, _ := colorful.Hex(colorB)

	for i := 0.0; i < steps; i++ {
		c := cA.BlendLuv(cB, i/steps)
		s = append(s, colorToHex(c))
	}
	return
}

// Convert a colorful.Color to a hexidecimal format compatible with termenv.
func colorToHex(c colorful.Color) string {
	return fmt.Sprintf("#%s%s%s", colorFloatToHex(c.R), colorFloatToHex(c.G), colorFloatToHex(c.B))
}

// Helper function for converting colors to hex. Assumes a value between 0 and 1.
func colorFloatToHex(f float64) (s string) {
	s = strconv.FormatInt(int64(f*255), 16)
	if len(s) == 1 {
		s = "0" + s
	}
	return
}

// GetViewFromSongs 从歌曲列表获取View
func GetViewFromSongs(songs []ds2.Song) []MenuItem {
	var menus []MenuItem
	for _, song := range songs {
		var artists []string
		for _, artist := range song.Artists {
			artists = append(artists, artist.Name)
		}
		menus = append(menus, MenuItem{Title: utils.ReplaceSpecialStr(song.Name), Subtitle: utils.ReplaceSpecialStr(strings.Join(artists, ","))})
	}

	return menus
}

// GetViewFromAlbums 从歌曲列表获取View
func GetViewFromAlbums(albums []ds2.Album) []MenuItem {
	var menus []MenuItem
	for _, album := range albums {
		var artists []string
		for _, artist := range album.Artists {
			artists = append(artists, artist.Name)
		}
		artistsStr := fmt.Sprintf("[%s]", strings.Join(artists, ","))
		menus = append(menus, MenuItem{Title: utils.ReplaceSpecialStr(album.Name), Subtitle: utils.ReplaceSpecialStr(artistsStr)})
	}

	return menus
}

// GetViewFromPlaylists 从歌单列表获取View
func GetViewFromPlaylists(playlists []ds2.Playlist) []MenuItem {
	var menus []MenuItem
	for _, playlist := range playlists {
		menus = append(menus, MenuItem{Title: utils.ReplaceSpecialStr(playlist.Name)})
	}

	return menus
}

// GetViewFromArtists 从歌手列表获取View
func GetViewFromArtists(artists []ds2.Artist) []MenuItem {
	var menus []MenuItem
	for _, artist := range artists {
		menus = append(menus, MenuItem{Title: utils.ReplaceSpecialStr(artist.Name)})
	}

	return menus
}

// GetViewFromUsers 用户列表获取View
func GetViewFromUsers(users []ds2.User) []MenuItem {
	var menus []MenuItem
	for _, user := range users {
		menus = append(menus, MenuItem{Title: utils.ReplaceSpecialStr(user.Nickname)})
	}

	return menus
}

// GetViewFromDjRadios DjRadio列表获取View
func GetViewFromDjRadios(radios []ds2.DjRadio) []MenuItem {
	var menus []MenuItem
	for _, radio := range radios {
		var dj string
		if radio.Dj.Nickname != "" {
			dj = fmt.Sprintf("[%s]", radio.Dj.Nickname)
		}
		menus = append(menus, MenuItem{Title: utils.ReplaceSpecialStr(radio.Name), Subtitle: utils.ReplaceSpecialStr(dj)})
	}

	return menus
}

// GetViewFromDjCate 分类列表获取View
func GetViewFromDjCate(categories []ds2.DjCategory) []MenuItem {
	var menus []MenuItem
	for _, category := range categories {
		menus = append(menus, MenuItem{Title: utils.ReplaceSpecialStr(category.Name)})
	}

	return menus
}
