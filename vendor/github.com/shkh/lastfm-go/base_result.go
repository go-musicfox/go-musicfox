package lastfm_go

import "encoding/xml"

type Base struct {
	XMLName xml.Name `xml:"lfm"`
	Status  string   `xml:"status,attr"`
	Inner   []byte   `xml:",innerxml"`
}

type ApiError struct {
	Code    int    `xml:"code,attr"`
	Message string `xml:",chardata"`
}
