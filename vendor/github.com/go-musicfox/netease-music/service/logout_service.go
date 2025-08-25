package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type LogoutService struct {
}

// Logout 注销登录
func (service *LogoutService) Logout() (float64, []byte, error) {
	api := "https://music.163.com/weapi/logout"
	data := make(map[string]interface{})
	cookiejar := util.GetGlobalCookieJar()
	csrfToken := util.GetCsrfToken(cookiejar)
	data["csrf_token"] = csrfToken
	code, bodyBytes, err := util.CallWeapi(api, data)
	if err != nil {
		return code, bodyBytes, err
	}
	return code, bodyBytes, nil
}
