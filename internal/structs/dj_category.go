package structs

import (
	"github.com/buger/jsonparser"
	"github.com/pkg/errors"
)

type DjCategory struct {
	Id   int64
	Name string
}

// NewDjCategoryFromJson 初始化分类
func NewDjCategoryFromJson(json []byte) (DjCategory, error) {
	var category DjCategory

	if len(json) == 0 {
		return category, errors.New("json is empty")
	}

	id, err := jsonparser.GetInt(json, "id")
	if err != nil {
		return category, err
	}
	category.Id = id

	if name, err := jsonparser.GetString(json, "name"); err == nil {
		category.Name = name
	}

	return category, nil
}
