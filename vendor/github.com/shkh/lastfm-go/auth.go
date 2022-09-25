package lastfm_go

import (
	"net/url"
)

// Mobile app style
func (api *Api) Login(username, password string) (err error) {
	defer func() { appendCaller(err, "lastfm.Login") }()

	var result AuthGetMobileSession
	args := P{"username": username, "password": password}
	if err = callPostWithoutSession("auth.getmobilesession", api.params, args, &result, P{
		"plain": []string{"username", "password"},
	}); err != nil {
		return
	}
	api.params.sk = result.Key
	//api.creds.username = result.Name
	return
}

// Desktop app style
func (api Api) GetToken() (token string, err error) {
	defer func() { appendCaller(err, "lastfm.GetToken") }()

	var result AuthGetToken
	if err = callGet("auth.gettoken", api.params, nil, &result, P{}); err != nil {
		return
	}
	token = result.Token
	return
}

func (api Api) GetAuthTokenUrl(token string) (uri string) {
	urlParams := url.Values{}
	urlParams.Add("api_key", api.params.apikey)
	urlParams.Add("token", token)
	uri = constructUrl(UriBrowserBase, urlParams)
	return
}

// Web app style
func (api Api) GetAuthRequestUrl(callback string) (uri string) {
	urlParams := url.Values{}
	urlParams.Add("api_key", api.params.apikey)
	if callback != "" {
		urlParams.Add("cb", callback)
	}
	uri = constructUrl(UriBrowserBase, urlParams)
	return
}

// Desktop and Web app style
func (api *Api) LoginWithToken(token string) (err error) {
	defer func() { appendCaller(err, "lastfm.LoginWithToken") }()

	var result AuthGetSession
	args := P{"token": token}
	if err = callPostWithoutSession("auth.getsession", api.params, args, &result, P{"plain": []string{"token"}}); err != nil {
		return
	}
	api.params.sk = result.Key
	//api.params.username = result.Name
	return
}
