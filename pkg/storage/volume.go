package storage

import (
	"github.com/go-musicfox/go-musicfox/pkg/constants"
)

type VolumeStorable interface {
	Volume() int
	SetVolume(volume int)
}

type Volume struct{}

func (v Volume) GetDbName() string {
	return constants.AppDBName
}

func (v Volume) GetTableName() string {
	return "default_bucket"
}

func (v Volume) GetKey() string {
	return "volume"
}
