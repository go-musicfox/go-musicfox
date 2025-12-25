package structs

// LRCData contains the original and translated lyrics data fetched from the API.
type LRCData struct {
	Original   string
	Translated string
	// Raw YRC (word-by-word) lyric data returned by Netease "lyric/new" API
	Yrc string
	// Raw translated LRC aligned to YRC returned by Netease "lyric/new" API
	Ytlrc string
	// Raw romanized LRC aligned to YRC returned by Netease "lyric/new" API
	Yromalrc string
}
