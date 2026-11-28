package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type DeviceinfoCenterUploadService struct {
	DeviceName string `json:"deviceName" form:"deviceName"`
	Name       string `json:"name" form:"name"`
}

func (service *DeviceinfoCenterUploadService) DeviceinfoCenterUpload() (float64, []byte) {

	options := &util.Options{
		Crypto: "eapi",
		Url:    "/api/deviceinfo/center/upload",
	}
	data := make(map[string]string)
	deviceName := service.DeviceName
	if deviceName == "" {
		deviceName = service.Name
	}
	data["deviceName"] = deviceName

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/eapi/deviceinfo/center/upload", data, options)

	return code, reBody
}
