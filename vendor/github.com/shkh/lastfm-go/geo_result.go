package lastfm_go

import "encoding/xml"

// geo.getTopArtists
type GeoGetTopArtists struct {
	XMLName    xml.Name `xml:"topartists"`
	Country    string   `xml:"country,attr"`
	Total      int      `xml:"total,attr"`
	Page       int      `xml:"page,attr"`
	PerPage    int      `xml:"perPage,attr"`
	TotalPages int      `xml:"totalPages,attr"`
	Artists    []struct {
		Rank       string `xml:"rank,attr"`
		Name       string `Xml:"name"`
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

// geo.getTopTracks
type GeoGetTopTracks struct {
	XMLName    xml.Name `xml:"toptracks"`
	Country    string   `xml:"country,attr"`
	Total      int      `xml:"total,attr"`
	Page       int      `xml:"page,attr"`
	PerPage    int      `xml:"perPage,attr"`
	TotalPages int      `xml:"totalPages,attr"`
	Tracks     []struct {
		Rank       string `xml:"rank,attr"`
		Name       string `xml:"name"`
		Duration   string `xml:"duration"`
		Listeners  string `xml:"listeners"`
		Mbid       string `xml:"mbid"`
		Url        string `xml:"url"`
		Streamable struct {
			FullTrack  string `xml:"fulltrack,attr"`
			Streamable string `xml:",chardata"`
		} `xml:"streamable"`
		Artist struct {
			Name string `xml:"artist"`
			Mbid string `xml:"mbid"`
			Url  string `xml:"url"`
		} `xml:"artist"`
		Images []struct {
			Size string `xml:"size,attr"`
			Url  string `xml:",chardata"`
		} `xml:"image"`
	} `xml:"track"`
}
