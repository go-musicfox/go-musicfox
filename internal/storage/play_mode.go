package storage

import (
	"github.com/go-musicfox/go-musicfox/internal/constants"
)

type PlayMode struct{}

func (p PlayMode) GetDbName() string {
	return constants.AppDBName
}

func (p PlayMode) GetTableName() string {
	return "default_bucket"
}

func (p PlayMode) GetKey() string {
	return "play_mode_int"
}
