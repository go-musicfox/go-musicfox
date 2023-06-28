package service

import (
	"encoding/json"

	"github.com/go-musicfox/netease-music/util"
)

type DigitalAlbumOrderingService struct {
	ID            string `json:"id" form:"id"`
	PaymentMethod string `json:"payment" form:"payment"`
	Quantity      string `json:"quantity" form:"quantity"`
}

func (service *DigitalAlbumOrderingService) DigitalAlbumOrdering() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	data["business"] = "Album"
	data["paymentMethod"] = service.PaymentMethod
	data["from"] = "web"

	var digitalResources []map[string]string
	dMap := make(map[string]string)
	dMap["business"] = "Album"
	dMap["resourceID"] = service.ID
	dMap["quantity"] = service.Quantity
	digitalResources = append(digitalResources, dMap)
	dig, _ := json.Marshal(digitalResources)

	data["digitalResources"] = string(dig)

	code, reBody, _ := util.CreateRequest("POST", `https://music.163.com/api/ordering/web/digital`, data, options)

	return code, reBody
}
