package service

import "encoding/json"

type BaseServiceResult struct {
}

// Format 将结构体或map格式化为json字符串形式
func Format(v interface{}) string {
	bytesData, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		panic(err)
	}
	return string(bytesData)
}
