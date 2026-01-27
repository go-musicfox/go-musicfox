package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type LoginRefreshService struct {
}

func (service *LoginRefreshService) LoginRefresh() (float64, []byte, error) {

	api := "https://music.163.com/weapi/login/token/refresh"
	data := make(map[string]interface{})
	cookiejar := util.GetGlobalCookieJar()
	util.ApplyRequestStrategy(cookiejar) // 为cookie应用反风控策略
	csrfToken := util.GetCsrfToken(cookiejar)
	data["csrf_token"] = csrfToken
	code, bodyBytes, err := util.CallWeapi(api, data, cookiejar)
	if err != nil {
		return code, bodyBytes, err
	}
	return code, bodyBytes, nil
}
