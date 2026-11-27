package service

import (
	"crypto/md5"
	"encoding/base64"
	"net/http"

	"github.com/go-musicfox/netease-music/util"
)

const anonymousIDXORKey = "3go8&$8*3*3h0k(2)2"

func anonymousEncodeID(deviceID string) string {
	xored := make([]byte, len(deviceID))
	for i := range deviceID {
		xored[i] = deviceID[i] ^ anonymousIDXORKey[i%len(anonymousIDXORKey)]
	}
	sum := md5.Sum(xored)
	return base64.StdEncoding.EncodeToString(sum[:])
}

func anonymousUsername(deviceID string) string {
	encoded := deviceID + " " + anonymousEncodeID(deviceID)
	return base64.StdEncoding.EncodeToString([]byte(encoded))
}

type RegisterAnonimousService struct {
	DeviceID string `json:"deviceId" form:"deviceId"`
}

func (service *RegisterAnonimousService) RegisterAnonimous() (float64, []byte) {

	deviceID := service.DeviceID
	if deviceID == "" {
		deviceID = util.GenerateSDeviceId()
	}
	options := &util.Options{
		Crypto: "xeapi",
		Url:    "/api/register/anonimous",
		Cookies: []*http.Cookie{
			{Name: "deviceId", Value: deviceID},
			{Name: "sDeviceId", Value: deviceID},
		},
	}
	data := map[string]string{"username": anonymousUsername(deviceID)}

	code, reBody, _ := util.CreateRequest("POST", "https://interface3.music.163.com/xeapi/register/anonimous", data, options)

	return code, reBody
}
