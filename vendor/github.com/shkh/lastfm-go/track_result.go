package lastfm_go

import "encoding/xml"

//track.addTags (empty)

// track.getCorrection
type TrackGetCorrection struct {
	XMLName    xml.Name `xml:"corrections"`
	Correction struct {
		Index           string `xml:"index,attr"`
		ArtistCorrected string `xml:"artistcorrected"`
		TrackCorrected  string `xml:"trackcorrected"`
		Track           struct {
			Name   string `xml:"name"`
			Mbid   string `xml:"mbid"`
			Url    string `xml:"url"`
			Artist struct {
				Name string `xml:"name"`
				Mbid string `xml:"mbid"`
				Url  string `xml:"url"`
			} `xml:"artist"`
		} `xml:"track"`
	} `xml:"correction"`
}

// track.getInfo
type TrackGetInfo struct {
	XMLName    xml.Name `xml:"track"`
	Id         string   `xml:"id"`
	Name       string   `xml:"name"`
	Mbid       string   `xml:"mbid"`
	Url        string   `xml:"url"`
	Duration   string   `xml:"duration"`
	Streamable struct {
		FullTrack  string `xml:"fulltrack,attr"`
		Streamable string `xml:",chardata"`
	} `xml:"streamable"`
	PlayCount     string `xml:"playcount"`
	UserPlayCount string `xml:"userplaycount"`
	UserLoved     string `xml:"userloved"`
	Artist        struct {
		Name string `xml:"name"`
		Mbid string `xml:"mbid"`
		Url  string `xml:"url"`
	} `xml:"artist"`
	Album struct {
		Position string `xml:"position,attr"`
		Artist   string `xml:"artist"`
		Title    string `xml:"title"`
		Mbid     string `xml:"mbid"`
		Url      string `xml:"url"`
		Images   []struct {
			Size string `xml:"size,attr"`
			Url  string `xml:",chardata"`
		} `xml:"image"`
	} `xml:"album"`
	TopTags []struct {
		Name string `xml:"name"`
		Url  string `xml:"url"`
	} `xml:"toptags>tag"`
	Wiki struct {
		Published string `xml:"published"`
		Summary   string `xml:"summary"`
		Content   string `xml:"content"`
	} `xml:"wiki"`
}

// track.getSimilar
type TrackGetSimilar struct {
	XMLName xml.Name `xml:"similartracks"`
	Track   string   `xml:"track,attr"`
	Artist  string   `xml:"artist,attr"`
	Tracks  []struct {
		Name       string `xml:"name"`
		PlayCount  string `xml:"playcount"`
		Mbid       string `xml:"mbid"`
		Match      string `xml:"match"`
		Url        string `xml:"url"`
		Streamable struct {
			FullTrack  string `xml:"fulltrack,attr"`
			Streamable string `xml:",chardata"`
		} `xml:"streamable"`
		Duration string `xml:"duration"`
		Artist   struct {
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

// track.getTags
type TrackGetTags struct {
	XMLName xml.Name `xml:"tags"`
	Artist  string   `xml:"artist,attr"`
	Track   string   `xml:"track,attr"`
	Tags    []struct {
		Name string `xml:"name"`
		Url  string `xml:"url"`
	} `xml:"tag"`
}

// track.getTopTags
type TrackGetTopTags struct {
	XMLName xml.Name `xml:"toptags"`
	Artist  string   `xml:"artist,attr"`
	Track   string   `xml:"track,attr"`
	Tags    []struct {
		Name  string `xml:"name"`
		Count string `xml:"count"`
		Url   string `xml:"url"`
	} `xml:"tag"`
}

//track.love (empty)

//track.removeTag (empty)

// track.scrobble
type TrackScrobble struct {
	XMLName   xml.Name `xml:"scrobbles"`
	Accepted  string   `xml:"accepted,attr"`
	Ignored   string   `xml:"ignored,attr"`
	Scrobbles []struct {
		Track struct {
			Corrected string `xml:"corrected,attr"`
			Name      string `xml:",chardata"`
		} `xml:"track"`
		Artist struct {
			Corrected string `xml:"corrected,attr"`
			Name      string `xml:",chardata"`
		} `xml:"artist"`
		Album struct {
			Corrected string `xml:"corrected,attr"`
			Name      string `xml:",chardata"`
		} `xml:"album"`
		AlbumArtist struct {
			Corrected string `xml:"corrected,attr"`
			Name      string `xml:",chardata"`
		} `xml:"albumArtist"`
		TimeStamp      string `xml:"timestamp"`
		IgnoredMessage struct {
			Corrected string `xml:"corrected,attr"`
			Body      string `xml:",chardata"`
		} `xml:"ignoredMessage"`
	} `xml:"scrobble"`
}

// track.search
type TrackSearch struct {
	XMLName    xml.Name `xml:"results"`
	OpenSearch string   `xml:"opensearch,attr"`
	For        string   `xml:"for,attr"`
	Query      struct {
		Role        string `xml:"role,attr"`
		SearchTerms string `xml:"searchTrems,attr"`
		StartPage   int    `xml:"startPage,attr"`
	} `xml:"Query"`
	TotalResults int `xml:"totalResults"`
	StartIndex   int `xml:"startIndex"`
	ItemsPerPage int `xml:"itemsPerPage"`
	Tracks       []struct {
		Name       string `xml:"name"`
		Mbid       string `xml:"mbid"`
		Artist     string `xml:"artist"`
		Url        string `xml:"url"`
		Streamable struct {
			FullTrack  string `xml:"fulltrack,attr"`
			Streamable string `xml:"streamable"`
		} `xml:"streamable"`
		Listeners string `xml:"listeners"`
		Images    []struct {
			Size string `xml:"size,attr"`
			Url  string `xml:",chardata"`
		} `xml:"image"`
	} `xml:"trackmatches>track"`
}

//track.unlove (empty)

// track.updateNowPlaying
type TrackUpdateNowPlaying struct {
	XMLName xml.Name `xml:"nowplaying"`
	Track   struct {
		Corrected string `xml:"corrected,attr"`
		Name      string `xml:",chardata"`
	} `xml:"track"`
	Artist struct {
		Corrected string `xml:"corrected,attr"`
		Name      string `xml:",chardata"`
	} `xml:"artist"`
	Album struct {
		Corrected string `xml:"corrected,attr"`
		Name      string `xml:",chardata"`
	} `xml:"album"`
	AlbumArtist struct {
		Corrected string `xml:"corrected,attr"`
		Name      string `xml:",chardata"`
	} `xml:"albumArtist"`
	IgnoredMessage struct {
		Corrected string `xml:"corrected,attr"`
		Body      string `xml:",chardata"`
	} `xml:"ignoredMessage"`
}
