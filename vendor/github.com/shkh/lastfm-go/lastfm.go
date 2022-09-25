package lastfm_go

const (
	UriApiSecBase  = "https://ws.audioscrobbler.com/2.0/"
	UriApiBase     = "http://ws.audioscrobbler.com/2.0/"
	UriBrowserBase = "https://www.last.fm/api/auth/"
)

type P map[string]interface{}

type Api struct {
	params  *apiParams
	Album   *albumApi
	Artist  *artistApi
	Chart   *chartApi
	Geo     *geoApi
	Library *libraryApi
	Tag     *tagApi
	Track   *trackApi
	User    *userApi
}

type apiParams struct {
	apikey    string
	secret    string
	sk        string
	useragent string
}

func New(key, secret string) (api *Api) {
	params := apiParams{key, secret, "", ""}
	api = &Api{
		params:  &params,
		Album:   &albumApi{&params},
		Artist:  &artistApi{&params},
		Chart:   &chartApi{&params},
		Geo:     &geoApi{&params},
		Library: &libraryApi{&params},
		Tag:     &tagApi{&params},
		Track:   &trackApi{&params},
		User:    &userApi{&params},
	}
	return
}

func (api *Api) SetSession(sessionkey string) {
	api.params.sk = sessionkey
}

func (api Api) GetSessionKey() (sk string) {
	sk = api.params.sk
	return
}

func (api *Api) SetUserAgent(useragent string) {
	api.params.useragent = useragent
}
