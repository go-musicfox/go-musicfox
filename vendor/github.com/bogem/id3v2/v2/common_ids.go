// Copyright 2016 Albert Nigmatzianov. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package id3v2

import "strings"

// Common IDs for ID3v2.3 and ID3v2.4.
var (
	V23CommonIDs = map[string]string{
		"Attached picture":                   "APIC",
		"Chapters":                           "CHAP",
		"Comments":                           "COMM",
		"Album/Movie/Show title":             "TALB",
		"BPM":                                "TBPM",
		"Composer":                           "TCOM",
		"Content type":                       "TCON",
		"Copyright message":                  "TCOP",
		"Date":                               "TDAT",
		"Playlist delay":                     "TDLY",
		"Encoded by":                         "TENC",
		"Lyricist/Text writer":               "TEXT",
		"File type":                          "TFLT",
		"Time":                               "TIME",
		"Content group description":          "TIT1",
		"Title/Songname/Content description": "TIT2",
		"Subtitle/Description refinement":    "TIT3",
		"Initial key":                        "TKEY",
		"Language":                           "TLAN",
		"Length":                             "TLEN",
		"Media type":                         "TMED",
		"Original album/movie/show title":    "TOAL",
		"Original filename":                  "TOFN",
		"Original lyricist/text writer":      "TOLY",
		"Original artist/performer":          "TOPE",
		"Original release year":              "TORY",
		"Popularimeter":                      "POPM",
		"File owner/licensee":                "TOWN",
		"Lead artist/Lead performer/Soloist/Performing group": "TPE1",
		"Band/Orchestra/Accompaniment":                        "TPE2",
		"Conductor/performer refinement":                      "TPE3",
		"Interpreted, remixed, or otherwise modified by":      "TPE4",
		"Part of a set":                "TPOS",
		"Publisher":                    "TPUB",
		"Track number/Position in set": "TRCK",
		"Recording dates":              "TRDA",
		"Internet radio station name":  "TRSN",
		"Internet radio station owner": "TRSO",
		"Size":                         "TSIZ",
		"ISRC":                         "TSRC",
		"Software/Hardware and settings used for encoding": "TSSE",
		"Year":                                     "TYER",
		"User defined text information frame":      "TXXX",
		"Unique file identifier":                   "UFID",
		"Unsynchronised lyrics/text transcription": "USLT",

		// Just for convenience.
		"Artist": "TPE1",
		"Title":  "TIT2",
		"Genre":  "TCON",
	}

	V24CommonIDs = map[string]string{
		"Attached picture":                   "APIC",
		"Chapters":                           "CHAP",
		"Comments":                           "COMM",
		"Album/Movie/Show title":             "TALB",
		"BPM":                                "TBPM",
		"Composer":                           "TCOM",
		"Content type":                       "TCON",
		"Copyright message":                  "TCOP",
		"Encoding time":                      "TDEN",
		"Playlist delay":                     "TDLY",
		"Original release time":              "TDOR",
		"Recording time":                     "TDRC",
		"Release time":                       "TDRL",
		"Tagging time":                       "TDTG",
		"Encoded by":                         "TENC",
		"Lyricist/Text writer":               "TEXT",
		"File type":                          "TFLT",
		"Involved people list":               "TIPL",
		"Content group description":          "TIT1",
		"Title/Songname/Content description": "TIT2",
		"Subtitle/Description refinement":    "TIT3",
		"Initial key":                        "TKEY",
		"Language":                           "TLAN",
		"Length":                             "TLEN",
		"Musician credits list":              "TMCL",
		"Media type":                         "TMED",
		"Mood":                               "TMOO",
		"Original album/movie/show title":    "TOAL",
		"Original filename":                  "TOFN",
		"Original lyricist/text writer":      "TOLY",
		"Original artist/performer":          "TOPE",
		"Popularimeter":                      "POPM",
		"File owner/licensee":                "TOWN",
		"Lead artist/Lead performer/Soloist/Performing group": "TPE1",
		"Band/Orchestra/Accompaniment":                        "TPE2",
		"Conductor/performer refinement":                      "TPE3",
		"Interpreted, remixed, or otherwise modified by":      "TPE4",
		"Part of a set":                "TPOS",
		"Produced notice":              "TPRO",
		"Publisher":                    "TPUB",
		"Track number/Position in set": "TRCK",
		"Internet radio station name":  "TRSN",
		"Internet radio station owner": "TRSO",
		"Album sort order":             "TSOA",
		"Performer sort order":         "TSOP",
		"Title sort order":             "TSOT",
		"ISRC":                         "TSRC",
		"Software/Hardware and settings used for encoding": "TSSE",
		"Set subtitle":                             "TSST",
		"User defined text information frame":      "TXXX",
		"Unique file identifier":                   "UFID",
		"Unsynchronised lyrics/text transcription": "USLT",

		// Deprecated frames of ID3v2.3.
		"Date":                  "TDRC",
		"Time":                  "TDRC",
		"Original release year": "TDOR",
		"Recording dates":       "TDRC",
		"Size":                  "",
		"Year":                  "TDRC",

		// Just for convenience.
		"Artist": "TPE1",
		"Title":  "TIT2",
		"Genre":  "TCON",
	}
)

// parsers is map, where key is ID of frame and value is function for the
// parsing of corresponding frame.
// You should consider that there is no text frame parser. That's why you should
// check at first, if it's a text frame:
//	if strings.HasPrefix(id, "T") {
//  	...
//	}
var parsers = map[string]func(*bufReader, byte) (Framer, error){
	"APIC": parsePictureFrame,
	"CHAP": parseChapterFrame,
	"COMM": parseCommentFrame,
	"POPM": parsePopularimeterFrame,
	"TXXX": parseUserDefinedTextFrame,
	"UFID": parseUFIDFrame,
	"USLT": parseUnsynchronisedLyricsFrame,
}

// mustFrameBeInSequence checks if frame with corresponding ID must
// be added to sequence.
func mustFrameBeInSequence(id string) bool {
	if id != "TXXX" && strings.HasPrefix(id, "T") {
		return false
	}

	switch id {
	case "MCDI", "ETCO", "SYTC", "RVRB", "MLLT", "PCNT", "RBUF", "POSS", "OWNE", "SEEK", "ASPI":
	case "IPLS", "RVAD": // Specific ID3v2.3 frames.
		return false
	}

	return true
}
