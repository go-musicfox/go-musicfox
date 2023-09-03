package storage

import (
	"github.com/go-musicfox/go-musicfox/internal/constants"
)

type LastSignIn struct{}

func (p LastSignIn) GetDbName() string {
	return constants.AppDBName
}

func (p LastSignIn) GetTableName() string {
	return "default_bucket"
}

func (p LastSignIn) GetKey() string {
	return "last_sign_in"
}
