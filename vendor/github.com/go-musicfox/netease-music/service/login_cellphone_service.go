package service

import (
	"crypto/md5"
	"encoding/hex"

	"github.com/go-musicfox/netease-music/util"
)

type LoginCellphoneService struct {
	Phone       string `json:"phone" form:"phone"`
	Countrycode string `json:"countrycode" form:"countrycode"`
	Password    string `json:"password" form:"password"`
	Md5password string `json:"md5_password" form:"md5_password"`
	Captcha     string `json:"captcha" from:"captcha"`
	CsrfToken   string `json:"csrf_token" from:"csrf_token"`
}

// LoginCellphone 使用手机号和密码登录
//
// 返回：
//   - code: 状态码
//   - bodyBytes：返回的响应体
//   - err：错误内容
func (service *LoginCellphoneService) LoginCellphone() (float64, []byte, error) {
	data := make(map[string]interface{})

	data["phone"] = service.Phone
	if service.Countrycode != "" {
		data["countrycode"] = service.Countrycode
	} else {
		data["countrycode"] = "86"
	}

	if service.Captcha != "" {
		data["captcha"] = service.Captcha
	}

	data["csrf_token"] = service.CsrfToken

	if service.Password != "" {
		h := md5.New()
		h.Write([]byte(service.Password))
		data["password"] = hex.EncodeToString(h.Sum(nil))
	} else {
		data["password"] = service.Md5password
	}
	data["rememberLogin"] = "true"

	api := "https://music.163.com/weapi/login/cellphone"
	cookieJar := util.GetGlobalCookieJar()
	util.ApplyRequestStrategy(cookieJar)
	code, bodyBytes, err := util.CallWeapi(api, data, cookieJar)
	return code, bodyBytes, err
}

// web端登录安全检查,需要获取checkToken的值
func (service *LoginCellphoneService) loginSecure() (float64, []byte, error) {
	data := make(map[string]interface{})
	data["phone"] = service.Phone
	if service.Countrycode != "" {
		data["countrycode"] = service.Countrycode
	} else {
		data["countrycode"] = "86"
	}
	data["checkToken"] = "" // 需要动态生成
	api := "https://music.163.com/api/user/login/secure"
	code, bodyBytes, err := util.CallWeapi(api, data)
	return code, bodyBytes, err
}
