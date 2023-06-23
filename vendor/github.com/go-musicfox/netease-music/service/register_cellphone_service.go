package service

import (
	"crypto/md5"
	"encoding/hex"
	"net/http"

	"github.com/go-musicfox/netease-music/util"
)

type RegisterCellphoneService struct {
	Phone    string `json:"phone" form:"phone"`
	Captcha  string `json:"captcha" form:"captcha"`
	Password string `json:"password" form:"password"`
	Nickname string `json:"nickname" form:"nickname"`
}

func (service *RegisterCellphoneService) RegisterCellphone() (float64, []byte) {

	cookiesOS := &http.Cookie{Name: "os", Value: "pc"}

	options := &util.Options{
		Crypto:  "weapi",
		Cookies: []*http.Cookie{cookiesOS},
	}
	data := make(map[string]string)

	data["phone"] = service.Phone
	h := md5.New()
	h.Write([]byte(service.Password))
	data["password"] = hex.EncodeToString(h.Sum(nil))
	data["captcha"] = service.Captcha
	data["nickname"] = service.Nickname

	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/weapi/register/cellphone`, data, options)

	return code, reBody
}
