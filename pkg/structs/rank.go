package structs

import (
	"encoding/json"

	"github.com/pkg/errors"
)

type Rank struct {
	Id        int64  `json:"id"`
	Name      string `json:"name"`
	Frequency string `json:"updateFrequency"` // 更新频率
}

// NewRankFromJson 获取排行榜信息
func NewRankFromJson(jsonBytes []byte) (Rank, error) {
	var rank Rank
	if len(jsonBytes) == 0 {
		return rank, errors.New("json is empty")
	}

	err := json.Unmarshal(jsonBytes, &rank)
	if err != nil {
		return rank, err
	}

	return rank, nil
}
