package lastfm_go

////////////
// auth.x //
////////////

// auth.getMobileSession
type AuthGetMobileSession struct {
	Name       string `xml:"name"` //username
	Key        string `xml:"key"`  //session key
	Subscriber bool   `xml:"subscriber"`
}

// auth.getToken
type AuthGetToken struct {
	Token string `xml:",chardata"`
}

// auth.getSession
type AuthGetSession AuthGetMobileSession
