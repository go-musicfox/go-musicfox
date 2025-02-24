package storage

import (
	"encoding/json"

	"github.com/go-musicfox/go-musicfox/internal/types"
)

type LastfmUser struct {
	Id         string `json:"id"`
	Name       string `json:"name"`
	RealName   string `json:"real_name"`
	Url        string `json:"url"`
	ApiKey     string `json:"api_key"`
	SessionKey string `json:"session_key"`
}

func (u *LastfmUser) GetDbName() string {
	return types.AppDBName
}

func (u *LastfmUser) GetTableName() string {
	return "default_bucket"
}

func (u *LastfmUser) GetKey() string {
	return "lastfm_user"
}

func (u *LastfmUser) InitFromStorage() {
	t := NewTable()
	if jsonStr, err := t.GetByKVModel(u); err == nil {
		_ = json.Unmarshal(jsonStr, u)
	}
}

func (u *LastfmUser) Store() {
	t := NewTable()
	_ = t.SetByKVModel(u, u)
}

func (u *LastfmUser) Clear() {
	t := NewTable()
	_ = t.DeleteByKVModel(u)
}
