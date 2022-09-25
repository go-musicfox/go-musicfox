package lastfm_go

import "encoding/xml"

// chart.getTopArtists
type ChartGetTopArtists struct {
	XMLName    xml.Name `xml:"artists"`
	Total      int      `xml:"total,attr"`
	Page       int      `xml:"page,attr"`
	PerPage    int      `xml:"perPage,attr"`
	TotalPages int      `xml:"totalPages,attr"`
	Artists    []struct {
		Name       string `xml:"name"`
		PlayCount  string `xml:"playcount"`
		Listeners  string `xml:"listeners"`
		Mbid       string `xml:"mbid"`
		Url        string `xml:"url"`
		Streamable string `xml:"streamable"`
		Images     []struct {
			Size string `xml:"size,attr"`
			Url  string `xml:",chardata"`
		} `xml:"image"`
	} `xml:"artist"`
}

// chart.getTopTags
type ChartGetTopTags struct {
	XMLName    xml.Name `xml:"tags"`
	Total      int      `xml:"total,attr"`
	Page       int      `xml:"page,attr"`
	PerPage    int      `xml:"perPage,attr"`
	TotalPages int      `xml:"totalPages,attr"`
	Tags       []struct {
		Name       string `xml:"name"`
		Url        string `xml:"url"`
		Reach      string `xml:"reach"`
		Taggings   string `xml:"taggings"`
		Streamable string `xml:"streamable"`
		Wiki       struct {
			Published string `xml:"published"`
			Summary   string `xml:"summary"`
			Content   string `xml:"content"`
		} `xml:"wiki"`
	} `xml:"tag"`
}

// chart.getTopTracks
type ChartGetTopTracks struct {
	XMLName    xml.Name `xml:"tracks"`
	Total      int      `xml:"total,attr"`
	Page       int      `xml:"page,attr"`
	PerPage    int      `xml:"perPage,attr"`
	TotalPages int      `xml:"totalPages,attr"`
	Tracks     []struct {
		Name       string `xml:"name"`
		Duration   string `xml:"duration"`
		PlayCount  string `xml:"playcount"`
		Listeners  string `xml:"listeners"`
		Mbid       string `xml:"mbid"`
		Url        string `xml:"url"`
		Streamable struct {
			FullTrack  string `xml:"fulltrack,attr"`
			Streamable string `xml:",chardata"`
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
