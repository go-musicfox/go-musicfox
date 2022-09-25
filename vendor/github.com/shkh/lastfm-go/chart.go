package lastfm_go

type chartApi struct {
	params *apiParams
}

// chart.getTopArtists
func (api chartApi) GetTopArtists(args map[string]interface{}) (result ChartGetTopArtists, err error) {
	defer func() { appendCaller(err, "lastfm.Chart.GetTopArtists") }()
	err = callGet("chart.gettopartists", api.params, args, &result, P{
		"plain": []string{"page", "limit"},
	})
	return
}

// chart.getTopTags
func (api chartApi) GetTopTags(args map[string]interface{}) (result ChartGetTopTags, err error) {
	defer func() { appendCaller(err, "lastfm.Chart.GetTopTags") }()
	err = callGet("chart.gettoptags", api.params, args, &result, P{
		"plain": []string{"page", "limit"},
	})
	return
}

// chart.getTopTracks
func (api chartApi) GetTopTracks(args map[string]interface{}) (result ChartGetTopTracks, err error) {
	defer func() { appendCaller(err, "lastfm.Chart.GetTopTracks") }()
	err = callGet("chart.gettoptracks", api.params, args, &result, P{
		"plain": []string{"page", "limit"},
	})
	return
}
