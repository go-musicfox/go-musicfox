package service

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/anhoder/netease-music/util"
	"net/http"
)

type LoginEmailService struct {
	Email       string `json:"email" form:"email"`
	Password    string `json:"password" form:"password"`
	Md5password string `json:"md5_password" form:"md5_password"`
}

func (service *LoginEmailService) LoginEmail() (float64, []byte) {

	cookiesOS := &http.Cookie{Name: "os", Value: "pc"}

	options := &util.Options{
		Crypto:  "weapi",
		Ua:      "pc",
		Cookies: []*http.Cookie{cookiesOS},
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
	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/weapi/login`, data, options)


	return code, reBody
}
