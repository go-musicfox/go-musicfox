package service

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/buger/jsonparser"
	"github.com/go-musicfox/netease-music/util"
)

type VipSignService struct {
}

func (service *VipSignService) VipSign() (float64, []byte) {

	taskCode, taskBody, _ := util.CreateRequest("POST", "https://music.163.com/api/vip-center-bff/task/sign", map[string]string{}, &util.Options{Crypto: "weapi"})

	checkinData := map[string]string{
		"signDayTime": strconv.FormatInt(time.Now().UnixMilli(), 10),
		"type":        "1",
	}
	checkinCode, checkinBody, _ := util.CreateRequest(
		"POST",
		"https://music.163.com/eapi/vipnewcenter/app/level/user/checkin/history/detail",
		checkinData,
		&util.Options{Crypto: "eapi", Url: "/api/vipnewcenter/app/level/user/checkin/history/detail"},
	)

	taskBusiness, _ := jsonparser.GetInt(taskBody, "code")
	checkinBusiness, _ := jsonparser.GetInt(checkinBody, "code")
	signed := taskCode == 200 && checkinCode == 200 && taskBusiness == 200 && checkinBusiness == 200
	message := "黑胶乐签打卡失败"
	if signed {
		message = "黑胶乐签打卡成功"
	}
	result, _ := json.Marshal(map[string]any{
		"code":          200,
		"taskSign":      json.RawMessage(taskBody),
		"checkinDetail": json.RawMessage(checkinBody),
		"signed":        signed,
		"message":       message,
	})
	return 200, result
}
