package service

import (
	"crypto/md5"
	"encoding/hex"

	"github.com/go-musicfox/netease-music/util"
)

type UserBindingcellphoneService struct {
	Phone       string `json:"phone" form:"phone"`
	Countrycode string `json:"countrycode" form:"countrycode"`
	Captcha     string `json:"captcha" form:"captcha"`
	Password    string `json:"password" form:"password"`
}

func (service *UserBindingcellphoneService) UserBindingcellphone() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	data["phone"] = service.Phone
	if service.Countrycode == "" {
		data["countrycode"] = "86"
	} else {
		data["countrycode"] = service.Countrycode
	}
	data["captcha"] = service.Captcha
	if service.Password == "" {
		data["password"] = ""
	} else {
		sum := md5.Sum([]byte(service.Password))
		data["password"] = hex.EncodeToString(sum[:])
	}

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/api/user/bindingCellphone", data, options)

	return code, reBody
}
