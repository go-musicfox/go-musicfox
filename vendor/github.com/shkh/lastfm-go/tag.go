package lastfm_go

type tagApi struct {
	params *apiParams
}

// tag.getInfo
func (api tagApi) GetInfo(args map[string]interface{}) (result TagGetInfo, err error) {
	defer func() { appendCaller(err, "lastfm.Tag.GetInfo") }()
	err = callGet("tag.getinfo", api.params, args, &result, P{
		"plain": []string{"lang", "artist", "mbid"},
	})
	return
}

// tag.getSimilar
func (api tagApi) GetSimilar(args map[string]interface{}) (result TagGetSimilar, err error) {
	defer func() { appendCaller(err, "lastfm.Tag.GetSimilar") }()
	err = callGet("tag.getsimilar", api.params, args, &result, P{
		"plain": []string{"tag"},
	})
	return
}

// tag.getTopAlbums
func (api tagApi) GetTopAlbums(args map[string]interface{}) (result TagGetTopAlbums, err error) {
	defer func() { appendCaller(err, "lastfm.Tag.GetTopAlbums") }()
	err = callGet("tag.gettopalbums", api.params, args, &result, P{
		"plain": []string{"tag", "limit", "page"},
	})
	return
}

// tag.getTopArtists
func (api tagApi) GetTopArtists(args map[string]interface{}) (result TagGetTopArtists, err error) {
	defer func() { appendCaller(err, "lastfm.Tag.GetTopArtists") }()
	err = callGet("tag.gettopartists", api.params, args, &result, P{
		"plain": []string{"tag", "limit", "page"},
	})
	return
}

// tag.getTopTags
func (api tagApi) GetTopTags(args map[string]interface{}) (result TagGetTopTags, err error) {
	defer func() { appendCaller(err, "lastfm.Tag.GetTopTags") }()
	err = callGet("tag.gettoptags", api.params, args, &result, P{})
	return
}

// tag.getTopTracks
func (api tagApi) GetTopTracks(args map[string]interface{}) (result TagGetTopTracks, err error) {
	defer func() { appendCaller(err, "lastfm.Tag.GetTopTracks") }()
	err = callGet("tag.gettoptracks", api.params, args, &result, P{
		"plain": []string{"tag", "limit", "page"},
	})
	return
}

// tag.getWeeklyChartList
func (api tagApi) GetTopWeeklyChartList(args map[string]interface{}) (result TagGetWeeklyChartList, err error) {
	defer func() { appendCaller(err, "lastfm.Tag.GetWeeklyChartList") }()
	err = callGet("tag.getweeklychartlist", api.params, args, &result, P{
		"plain": []string{"tag"},
	})
	return
}
