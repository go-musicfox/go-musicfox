package lastfm_go

import "encoding/xml"

// library.getArtists
type LibraryGetArtists struct {
	XMLName    xml.Name `xml:"artists"`
	User       string   `xml:"user,attr"`
	Total      int      `xml:"total,attr"`
	Page       int      `xml:"page,attr"`
	PerPage    int      `xml:"perPage,attr"`
	TotalPages int      `xml:"totalPages,attr"`
	Artists    []struct {
		Name       string `xml:"name"`
		PlayCount  string `xml:"playcount"`
		TagCount   string `xml:"tagcount"`
		Mbid       string `xml:"mbid"`
		Url        string `xml:"url"`
		Streamable string `xml:"streamable"`
		Images     []struct {
			Size string `xml:"size,attr"`
			Url  string `xml:",chardata"`
		} `xml:"image"`
	} `xml:"artist"`
}
