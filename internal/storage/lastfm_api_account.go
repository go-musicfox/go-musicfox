package storage

import (
	"encoding/json"

	"github.com/go-musicfox/go-musicfox/internal/types"
)

type LastfmApiAccount struct {
	Key    string `json:"key"`
	Secret string `json:"secret"`
}

func (u *LastfmApiAccount) GetDbName() string {
	return types.AppDBName
}

func (u *LastfmApiAccount) GetTableName() string {
	return "default_bucket"
}

func (u *LastfmApiAccount) GetKey() string {
	return "lastfm_api_account"
}

func (u *LastfmApiAccount) InitFromStorage() {
	t := NewTable()
	if jsonStr, err := t.GetByKVModel(u); err == nil {
		_ = json.Unmarshal(jsonStr, u)
	}
}

func (u *LastfmApiAccount) Store() {
	t := NewTable()
	_ = t.SetByKVModel(u, u)
}

func (u *LastfmApiAccount) Clear() {
	t := NewTable()
	_ = t.DeleteByKVModel(u)
}
