package lastfm_go

type geoApi struct {
	params *apiParams
}

// geo.getTopArtists
func (api geoApi) GetTopArtists(args map[string]interface{}) (result GeoGetTopArtists, err error) {
	defer func() { appendCaller(err, "lastfm.Geo.GetTopArtists") }()
	err = callGet("geo.gettopartists", api.params, args, &result, P{
		"plain": []string{"country", "limit", "page"},
	})
	return
}

// geo.getTopTracks
func (api geoApi) GetTopTracks(args map[string]interface{}) (result GeoGetTopTracks, err error) {
	defer func() { appendCaller(err, "lastfm.Geo.GetTopTracks") }()
	err = callGet("geo.gettoptracks", api.params, args, &result, P{
		"plain": []string{"country", "location", "limit", "page"},
	})
	return
}
