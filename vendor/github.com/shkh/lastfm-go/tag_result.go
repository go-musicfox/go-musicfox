package lastfm_go

import "encoding/xml"

// tag.getinfo
type TagGetInfo struct {
	XMLName    xml.Name `tag`
	Name       string   `xml:"name"`
	Url        string   `xml:"url"`
	Reach      string   `xml:"reach"`
	Taggings   string   `xml:"taggings"`
	Streamable string   `xml:"streamable"`
	Wiki       struct {
		Published string `xml:"published"`
		Summary   string `xml:"summary"`
		Content   string `xml:"content"`
	} `xml:"wiki"`
}

// tag.getSimilar
type TagGetSimilar struct {
	XMLName xml.Name `xml:"similartags"`
	Tag     string   `xml:"tag,attr"`
	Tags    []struct {
		Name       string `xml:"name"`
		Url        string `xml:"url"`
		Streamable string `xml:"streamable"`
	} `xml:"tag"`
}

// tag.getTopAlbums
type TagGetTopAlbums struct {
	XMLName    xml.Name `xml:"topalbums"`
	Tag        string   `xml:"tag,attr"`
	Total      int      `xml:"total,attr"`
	Page       int      `xml:"page,attr"`
	PerPage    int      `xml:"perPage,attr"`
	TotalPages int      `xml:"totalPages,attr"`
	Albums     []struct {
		Rank   string `xml:"rank,attr"`
		Name   string `xml:"name"`
		Url    string `xml:"url"`
		Artist struct {
			Name string `xml:"name"`
			Mbid string `xml:"mbid"`
			Url  string `xml:"url"`
		} `xml:"artist"`
		Images []struct {
			Size string `xml:"size,attr"`
			Url  string `xml:",chardata"`
		} `xml:"image"`
	} `xml:"album"`
}

// tag.getTopArtists
type TagGetTopArtists struct {
	XMLName xml.Name `xml:"topartists"`
	Tag     string   `xml:"tag,attr"`
	//Total      string   `xml:"total,attr"`
	//Page       string   `xml:"page,attr"`
	//PerPage    string   `xml:"perPage,attr"`
	//TotalPages string   `xml:"totalPages"`
	Artists []struct {
		Rank       string `xml:"rank,attr"`
		Name       string `xml:"name"`
		Url        string `xml:"url"`
		Streamable string `xml:"streamable"`
		Images     []struct {
			Size string `xml:"size,attr"`
			Url  string `xml:",chardata"`
		} `xml:"image"`
	} `xml:"artist"`
}

// tag.getTopTags
type TagGetTopTags struct {
	XMLName xml.Name `xml:"toptags"`
	Tags    []struct {
		Name  string `xml:"name"`
		Count string `xml:"count"`
		Url   string `xml:"url"`
	} `xml:"tag"`
}

// tag.getTopTracks
type TagGetTopTracks struct {
	XMLName    xml.Name `xml:"toptracks"`
	Tag        string   `xml:"tag,attr"`
	Total      int      `xml:"total,attr"`
	Page       int      `xml:"page,attr"`
	PerPage    int      `xml:"perPage,attr"`
	TotalPages int      `xml:"totalPages,attr"`
	Tracks     []struct {
		Rank       string `xml:"rank,attr"`
		Name       string `xml:"name"`
		Duration   string `xml:"duration"`
		Mbid       string `xml:"mbid"`
		Url        string `xml:"url"`
		Streamable struct {
			FullTrack  string `xml:"fulltrack,attr"`
			Streamable string `xml:"streamable"`
		} `xml:"streamable"`
		Artist struct {
			Name string `xml:"name"`
			Mbid string `xml:"mbid"`
			Url  string `xml:"url"`
		} `xml:"artist"`
		Images []struct {
			Size string `xml:"size,attr"`
			Url  string `xml:",chardata"`
		} `xml:"image"`
	} `xml:"track"`
}

// tag.getWeeklyChartList
type TagGetWeeklyChartList struct {
	XMLName xml.Name `xml:"weeklychartlist"`
	Tag     string   `xml:"tag,attr"`
	Charts  []struct {
		From string `xml:"from,attr"`
		To   string `xml:"to,attr"`
	} `xml:"chart"`
}
