package service

import (
	"github.com/go-musicfox/netease-music/util"
)

type DigitalAlbumSalesService struct {
	Ids string `json:"ids" form:"ids"`
}

func (service *DigitalAlbumSalesService) DigitalAlbumSales() (float64, []byte) {

	options := &util.Options{
		Crypto: "weapi",
	}
	data := make(map[string]string)
	data["albumIds"] = service.Ids

	code, reBody, _ := util.CreateRequest("POST", "https://music.163.com/api/vipmall/albumproduct/album/query/sales", data, options)

	return code, reBody
}
