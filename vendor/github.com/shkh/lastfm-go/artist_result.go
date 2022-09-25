package lastfm_go

import "encoding/xml"

//////////////
// artist.x //
//////////////

//artist.addTags (empty)

// artist.getCorrection
type ArtistGetCorrection struct {
	XMLName    xml.Name `xml:"corrections"`
	Correction struct {
		Index  string `xml:"index,attr"`
		Artist struct {
			Name string `xml:"name"`
			Mbid string `xml:"mbid"`
			Url  string `xml:"url"`
		} `xml:"artist"`
	} `xml:"correction"`
}

// artist.getInfo
type ArtistGetInfo struct {
	Name   string `xml:"name"`
	Mbid   string `xml:"mbid"`
	Url    string `xml:"url"`
	Images []struct {
		Size string `xml:"size,attr"`
		Url  string `xml:",chardata"`
	} `xml:"image"`
	Streamable string `xml:"streamable"`
	Stats      struct {
		Listeners string `xml:"listeners"`
		Plays     string `xml:"plays"`
	} `xml:"stats"`
	//Similar struct {
	//Artists []struct {
	//Name   string `xml:"name"`
	//Url    string `xml:"url"`
	//Images []struct {
	//Size string `xml:"size,attr"`
	//Url  string `xml:",chardata"`
	//} `xml:"image"`
	//} `xml:"artist"`
	//} `xml:"similar"`
	Similars []struct {
		Name   string `xml:"name"`
		Url    string `xml:"url"`
		Images []struct {
			Size string `xml:"size,attr"`
			Url  string `xml:",chardata"`
		} `xml:"image"`
	} `xml:"similar>artist"`
	Tags []struct {
		Name string `xml:"name"`
		Url  string `xml:"url"`
	} `xml:"tags>tag"`
	Bio struct {
		Links []struct {
			Rel string `xml:"rel,attr"`
			Url string `xml:"href,attr"`
		} `xml:"links>link"`
		Published  string `xml:"published"`
		Summary    string `xml:"summary"`
		Content    string `xml:"content"`
		YearFormed string `xml:"yearformed"`
	} `xml:"bio"`
}

// artist.getSimilar
type ArtistGetSimilar struct {
	XMLName  xml.Name `xml:"similarartists"`
	Artist   string   `xml:"artist,attr"`
	Similars []struct {
		Name   string `xml:"name"`
		Mbid   string `xml:"mbid"`
		Match  string `xml:"match"`
		Url    string `xml:"url"`
		Images []struct {
			Size string `xml:"size,attr"`
			Url  string `xml:",chardata"`
		} `xml:"image"`
		Streamable string `xml:"streamable"`
	} `xml:"artist"`
}

// artist.getTags
type ArtistGetTags struct {
	XMLName xml.Name `xml:"tags"`
	Artist  string   `xml:"artist,attr"`
	Tags    []struct {
		Name string `xml:"name"`
		Url  string `xml:"url"`
	} `xml:"tag"`
}

// artist.getTopAlbums
type ArtistGetTopAlbums struct {
	XMLName    xml.Name `xml:"topalbums"`
	Artist     string   `xml:"artist,attr"`
	Total      int      `xml:"total,attr"`
	Page       int      `xml:"page,attr"`
	PerPage    int      `xml:"perPage,attr"`
	TotalPages int      `xml:"totalPages,attr"`
	Albums     []struct {
		Rank   string `xml:"rank,attr"`
		Name   string `xml:"name"`
		Mbid   string `xml:"mbid"`
		Artist struct {
			Name string `xml:"name"`
			Mbid string `xml:"mbid"`
			Url  string `xml:"url"`
		} `xml:"artist"`
		PlayCount string `xml:"playcount"`
		Url       string `xml:"url"`
		Images    []struct {
			Size string `xml:"size,attr"`
			Url  string `xml:",chardata"`
		} `xml:"image"`
	} `xml:"album"`
}

// artist.getTopTags
type ArtistGetTopTags struct {
	XMLName xml.Name `xml:"toptags"`
	Artist  string   `xml:"artist,attr"`
	Tags    []struct {
		Name  string `xml:"name"`
		Count string `xml:"count"`
		Url   string `xml:"url"`
	} `xml:"tag"`
}

// artist.getTopTracks
type ArtistGetTopTracks struct {
	XMLName    xml.Name `xml:"toptracks"`
	Artist     string   `xml:"artist,attr"`
	Total      int      `xml:"total,attr"`
	Page       int      `xml:"page,attr"`
	PerPage    int      `xml:"perPage,attr"`
	TotalPages int      `xml:"totalPages,attr"`
	Tracks     []struct {
		Rank      string `xml:"rank,attr"`
		Name      string `xml:"name"`
		Duration  string `xml:"duration"`
		PlayCount string `xml:"playcount"`
	} `xml:"track"`
}

//artist.removeTag (empty)

// artist.search
type ArtistSearch struct {
	XMLName    xml.Name `xml:"results"`
	OpenSearch string   `xml:"opensearch,attr"`
	For        string   `xml:"for,attr"`
	Query      struct {
		Role        string `xml:"role,attr"`
		SearchTerms string `xml:"searchTrems,attr"`
		StartPage   int    `xml:"startPage,attr"`
	} `xml:"Query"`
	TotalResults  int `xml:"totalResults"`
	StartIndex    int `xml:"startIndex"`
	ItemsPerPage  int `xml:"itemsPerPage"`
	ArtistMatches []struct {
		Name   string `xml:"name"`
		Mbid   string `xml:"mbid"`
		Url    string `xml:"url"`
		Images []struct {
			Size string `xml:"size,attr"`
			Url  string `xml:",chardata"`
		} `xml:"image"`
		Listeners string `xml:"listeners"`
	} `xml:"artistmatches>artist"`
}
