package service

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/go-musicfox/netease-music/util"
)

type AdListeningRightsGainService struct {
	ReqUid                  string `json:"reqUid" form:"reqUid"`
	ExposureTime            string `json:"exposureTime" form:"exposureTime"`
	Uid                     string `json:"uid" form:"uid"`
	ClickTime               string `json:"clickTime" form:"clickTime"`
	ExtraRightsType         string `json:"extraRightsType" form:"extraRightsType"`
	PlayContinuously        string `json:"playContinuously" form:"playContinuously"`
	Source                  string `json:"source" form:"source"`
	CreativeType            string `json:"creativeType" form:"creativeType"`
	RightsGainMethod        string `json:"rightsGainMethod" form:"rightsGainMethod"`
	ExtraRightsGainMethod   string `json:"extraRightsGainMethod" form:"extraRightsGainMethod"`
	ExtraRightsGainDuration string `json:"extraRightsGainDuration" form:"extraRightsGainDuration"`
	NextRightsGainDuration  string `json:"nextRightsGainDuration" form:"nextRightsGainDuration"`
	RightsGainType          string `json:"rightsGainType" form:"rightsGainType"`
	RightsGainDuration      string `json:"rightsGainDuration" form:"rightsGainDuration"`
	GainMethodStep          string `json:"gainMethodStep" form:"gainMethodStep"`
	GeneralRightsInfo       string `json:"generalRightsInfo" form:"generalRightsInfo"`
	RightsExtJson           string `json:"rightsExtJson" form:"rightsExtJson"`
	AppInfo                 string `json:"appInfo" form:"appInfo"`
	ContextInfo             string `json:"contextInfo" form:"contextInfo"`
	Installed               string `json:"installed" form:"installed"`
	SniffTime               string `json:"sniffTime" form:"sniffTime"`
}

func adGainInt(value string) (int, bool) {
	if value == "" {
		return 0, false
	}
	v, err := strconv.Atoi(value)
	return v, err == nil
}

func (service *AdListeningRightsGainService) AdListeningRightsGain() (float64, []byte) {

	now := time.Now().UnixMilli()
	gainMethod, ok := adGainInt(service.RightsGainMethod)
	if !ok || gainMethod == 0 {
		gainMethod = 2
	}
	creativeType, ok := adGainInt(service.CreativeType)
	if !ok || creativeType == 0 {
		creativeType = 36
	}
	exposure := service.ExposureTime
	if exposure == "" {
		exposure = strconv.FormatInt(now, 10)
	}
	click := service.ClickTime
	if click == "" {
		click = strconv.FormatInt(now, 10)
	}

	param := map[string]any{
		"reqUid":           service.ReqUid,
		"exposureTime":     exposure,
		"clickTime":        click,
		"creativeType":     creativeType,
		"rightsGainMethod": gainMethod,
		"playContinuously": service.PlayContinuously != "" && service.PlayContinuously != "0" && service.PlayContinuously != "false",
	}
	if service.GeneralRightsInfo != "" {
		param["generalRightsInfo"] = service.GeneralRightsInfo
	}
	if v, ok := adGainInt(service.Uid); ok {
		param["userId"] = v
	}
	for key, value := range map[string]string{
		"extraRightsType":         service.ExtraRightsType,
		"source":                  service.Source,
		"extraRightsGainMethod":   service.ExtraRightsGainMethod,
		"extraRightsGainDuration": service.ExtraRightsGainDuration,
		"nextRightsGainDuration":  service.NextRightsGainDuration,
		"rightsGainType":          service.RightsGainType,
		"rightsGainDuration":      service.RightsGainDuration,
		"gainMethodStep":          service.GainMethodStep,
		"installed":               service.Installed,
	} {
		if v, ok := adGainInt(value); ok {
			param[key] = v
		}
	}
	if service.RightsExtJson != "" {
		param["rightsExtJson"] = service.RightsExtJson
	}
	if service.ContextInfo != "" {
		param["contextInfo"] = service.ContextInfo
	}
	if service.AppInfo != "" {
		var appInfo any
		if json.Unmarshal([]byte(service.AppInfo), &appInfo) == nil {
			param["appInfo"] = appInfo
		}
	}
	if gainMethod == 3 || gainMethod == 6 {
		sniff := service.SniffTime
		if sniff == "" {
			sniff = strconv.FormatInt(now, 10)
		}
		if v, ok := adGainInt(sniff); ok {
			param["sniffTime"] = v
		}
	}
	reqParam, _ := json.Marshal(param)
	data := map[string]string{"reqParam": string(reqParam)}

	options := &util.Options{
		Crypto:     "xeapi",
		Url:        "/api/ad/listening/rights/gain",
		CheckToken: "v3",
	}
	code, reBody, _ := util.CreateRequest("POST", "https://interface3.music.163.com/xeapi/ad/listening/rights/gain", data, options)

	return code, reBody
}
