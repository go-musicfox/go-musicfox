package service

import (
	"encoding/json"

	"github.com/go-musicfox/netease-music/util"
)

type YunbeiService struct {
}

// YunbeiSignResult 云贝签到结果
type YunbeiSignResult struct {
	Code float64 `json:"code"`
	Data struct {
		Sign      bool `json:"sign"`      //是否签到成功
		YunbeiNum int  `json:"yunbeiNum"` //获得云贝数量
	} `json:"data"`
	Message string `json:"message"`
}

// YunbeiInfoResult 云贝信息结果
type YunbeiInfoResult struct {
	Code       float64 `json:"code"`
	Level      int     `json:"level"` // 用户等级
	MobileSign bool    `json:"mobileSign"`
	PCSign     bool    `json:"pcSign"`
	VIPType    int     `json:"viptype"`    // vip类型
	ExpireTime int     `json:"expiretime"` // 过期时间
	UserPoint  struct {
		Balance      int `json:"balance"`      // 云贝余额
		BlockBalance int `json:"blockBalance"` //失效云贝余额
		Status       int `json:"status"`
		UpdateTime   int `json:"updateTime"` //更新时间
		UserID       int `json:"userId"`     // 用户id
	} `json:"userPoint"`
}

// YunbeiTasksResult 云贝任务列表结果
type YunbeiTasksResult struct {
	Code    float64     `json:"code"`
	Message string      `json:"message"`
	Data    []tasksData `json:"data"`
}

type tasksData struct {
	TaskId           int         `json:"taskId"` // 任务id
	UserTaskId       int         `json:"userTaskId"`
	TaskName         string      `json:"taskName"`  // 任务名称
	TaskPoint        int         `json:"taskPoint"` // 完成可得云贝
	WebPicUrl        string      `json:"webPicUrl"`
	CompletedIconUrl interface{} `json:"completedIconUrl"`
	BackgroundPicUrl interface{} `json:"backgroundPicUrl"`
	WordsPicUrl      interface{} `json:"wordsPicUrl"`
	Link             string      `json:"link"`
	LinkText         string      `json:"linkText"`
	Completed        bool        `json:"completed"`
	CompletedPoint   int         `json:"completedPoint"`
	Status           int         `json:"status"`
	TargetStatus     interface{} `json:"targetStatus"`
	TargetPoint      int         `json:"targetPoint"`
	TargetUserTaskId int         `json:"targetUserTaskId"`
	TaskDescription  string      `json:"taskDescription"`
	Position         int         `json:"position"`
	ActionType       int         `json:"actionType"`
	TaskType         string      `json:"taskType"`
	ExtInfoMap       interface{} `json:"extInfoMap"`
	Period           int         `json:"period"`
	SubAction        string      `json:"subAction"`
}

// Sign 云贝签到
func (service *YunbeiService) Sign() (result YunbeiSignResult, err error) {
	data := map[string]interface{}{}
	api := "https://music.163.com/api/pointmall/user/sign"
	_, bytesData, err := util.CallWeapi(api, data)
	if err != nil {
		return result, err
	}
	err = json.Unmarshal(bytesData, &result)
	if err != nil {
		return result, err
	}
	return result, nil
}

// Info 云贝信息
func (service *YunbeiService) Info() (result YunbeiInfoResult, err error) {
	data := map[string]interface{}{}
	api := "https://music.163.com/api/v1/user/info"
	_, bytesData, err := util.CallWeapi(api, data)
	if err != nil {
		return result, err
	}
	err = json.Unmarshal(bytesData, &result)
	if err != nil {
		return result, err
	}
	return result, nil
}

// Tasks 云贝任务列表
func (service *YunbeiService) Tasks() (result YunbeiTasksResult, err error) {
	data := map[string]interface{}{}
	api := "https://music.163.com/api/usertool/task/list/all"
	_, bytesData, err := util.CallWeapi(api, data)
	if err != nil {
		return result, err
	}
	err = json.Unmarshal(bytesData, &result)
	if err != nil {
		return result, err
	}
	return result, nil
}
