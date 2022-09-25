package lastfm_go

type userApi struct {
	params *apiParams
}

// user.getArtistTracks
func (api userApi) GetArtistTracks(args map[string]interface{}) (result UserGetArtistTracks, err error) {
	defer func() { appendCaller(err, "lastfm.User.GetArtistTracks") }()
	err = callGet("user.getartisttracks", api.params, args, &result, P{
		"plain": []string{"user", "artist", "startTimeStamp", "page", "endTimeStamp"},
	})
	return
}

// user.getFriends
func (api userApi) GetFriends(args map[string]interface{}) (result UserGetFriends, err error) {
	defer func() { appendCaller(err, "lastfm.User.GetFriends") }()
	err = callGet("user.getfriends", api.params, args, &result, P{
		"plain": []string{"user", "recenttracks", "limit", "page"},
	})
	return
}

// user.getInfo
func (api userApi) GetInfo(args map[string]interface{}) (result UserGetInfo, err error) {
	defer func() { appendCaller(err, "lastfm.User.GetInfo") }()
	if _, ok := args["user"]; !ok && api.params.sk != "" {
		err = callPost("user.getinfo", api.params, args, &result, P{})
	} else {
		err = callGet("user.getinfo", api.params, args, &result, P{
			"plain": []string{"user"},
		})
	}
	return
}

// user.getLovedTracks
func (api userApi) GetLovedTracks(args map[string]interface{}) (result UserGetLovedTracks, err error) {
	defer func() { appendCaller(err, "lastfm.User.GetLovedTracks") }()
	err = callGet("user.getlovedtracks", api.params, args, &result, P{
		"plain": []string{"user", "limit", "page"},
	})
	return
}

// user.getPersonalTags
func (api userApi) GetPersonalTags(args map[string]interface{}) (result UserGetPersonalTags, err error) {
	defer func() { appendCaller(err, "lastfm.User.GetPersonalTags") }()
	err = callGet("user.getPersonalTags", api.params, args, &result, P{
		"plain": []string{"user", "tag", "taggingtype", "limit", "page"},
	})
	return
}

// user.getRecentTracks
func (api userApi) GetRecentTracks(args map[string]interface{}) (result UserGetRecentTracks, err error) {
	defer func() { appendCaller(err, "lastfm.User.GetRecentTracks") }()
	err = callGet("user.getrecenttracks", api.params, args, &result, P{
		"plain": []string{"user", "limit", "page", "from", "extended", "to"},
	})
	return
}

// user.getTopAlbums
func (api userApi) GetTopAlbums(args map[string]interface{}) (result UserGetTopAlbums, err error) {
	defer func() { appendCaller(err, "lastfm.User.GetTopAlbums") }()
	err = callGet("user.gettopalbums", api.params, args, &result, P{
		"plain": []string{"user", "period", "limit", "page"},
	})
	return
}

// user.getTopArtists
func (api userApi) GetTopArtists(args map[string]interface{}) (result UserGetTopArtists, err error) {
	defer func() { appendCaller(err, "lastfm.User.GetTopArtists") }()
	err = callGet("user.gettopartists", api.params, args, &result, P{
		"plain": []string{"user", "period", "limit", "page"},
	})
	return
}

// user.getTopTags
func (api userApi) GetTopTags(args map[string]interface{}) (result UserGetTopTags, err error) {
	defer func() { appendCaller(err, "lastfm.User.GetTopTags") }()
	err = callGet("user.gettoptags", api.params, args, &result, P{
		"plain": []string{"user", "limit"},
	})
	return
}

// user.getTopTracks
func (api userApi) GetTopTracks(args map[string]interface{}) (result UserGetTopTracks, err error) {
	defer func() { appendCaller(err, "lastfm.User.GetTopTracks") }()
	err = callGet("user.gettoptracks", api.params, args, &result, P{
		"plain": []string{"user", "period", "limit", "page"},
	})
	return
}

// user.getWeeklyAlbumChart
func (api userApi) GetWeeklyAlbumChart(args map[string]interface{}) (result UserGetWeeklyAlbumChart, err error) {
	defer func() { appendCaller(err, "lastfm.User.GetWeeklyAlbumChart") }()
	err = callGet("user.getweeklyalbumchart", api.params, args, &result, P{
		"plain": []string{"user", "from", "to"},
	})
	return
}

// user.getWeeklyArtistChart
func (api userApi) GetWeeklyArtistChart(args map[string]interface{}) (result UserGetWeeklyArtistChart, err error) {
	defer func() { appendCaller(err, "lastfm.User.GetWeeklyArtistChart") }()
	err = callGet("user.getweeklyartistchart", api.params, args, &result, P{
		"plain": []string{"user", "from", "to"},
	})
	return
}

// user.getWeeklyChartList
func (api userApi) GetWeeklyChartList(args map[string]interface{}) (result UserGetWeeklyChartList, err error) {
	defer func() { appendCaller(err, "lastfm.User.GetWeeklyChartList") }()
	err = callGet("user.getweeklychartlist", api.params, args, &result, P{
		"plain": []string{"user"},
	})
	return
}

// user.getWeeklyTrackChart
func (api userApi) GetWeeklyTrackChart(args map[string]interface{}) (result UserGetWeeklyTrackChart, err error) {
	defer func() { appendCaller(err, "lastfm.User.GetWeeklyTrackChart") }()
	err = callGet("user.getweeklytrackchart", api.params, args, &result, P{
		"plain": []string{"user", "from", "to"},
	})
	return
}
