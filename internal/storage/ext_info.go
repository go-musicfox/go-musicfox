package storage

import (
	"github.com/go-musicfox/go-musicfox/internal/types"
)

type ExtInfo struct {
	StorageVersion string `json:"storage_version"`
}

func (e ExtInfo) GetDbName() string {
	return types.AppDBName
}

func (e ExtInfo) GetTableName() string {
	return "default_bucket"
}

func (e ExtInfo) GetKey() string {
	return "ext_info"
}
