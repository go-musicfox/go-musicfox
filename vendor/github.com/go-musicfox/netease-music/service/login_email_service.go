package service

import (
	"crypto/md5"
	"encoding/hex"
	"net/http"

	"github.com/go-musicfox/netease-music/util"
)

type LoginEmailService struct {
	Email       string `json:"email" form:"email"`
	Password    string `json:"password" form:"password"`
	Md5password string `json:"md5_password" form:"md5_password"`
}

func (service *LoginEmailService) LoginEmail() (float64, []byte) {
	options := &util.Options{
		Crypto: "weapi",
		Ua:     "pc",
		Cookies: []*http.Cookie{
			{Name: "os", Value: "ios"},
			{Name: "appver", Value: "8.7.01"},
		},
	}
	data := make(map[string]string)

	data["username"] = service.Email
	if service.Password != "" {
		h := md5.New()
		h.Write([]byte(service.Password))
		data["password"] = hex.EncodeToString(h.Sum(nil))
	} else {
		data["password"] = service.Md5password
	}
	data["rememberLogin"] = "true"

	//reBody, cookies := util.CreateRequest("POST", `https://www.httpbin.org/post`, data, options)
	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/api/login`, data, options)

	return code, reBody
}
