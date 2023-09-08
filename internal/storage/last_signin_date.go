package storage

import (
	"github.com/go-musicfox/go-musicfox/internal/types"
)

type LastSignIn struct{}

func (p LastSignIn) GetDbName() string {
	return types.AppDBName
}

func (p LastSignIn) GetTableName() string {
	return "default_bucket"
}

func (p LastSignIn) GetKey() string {
	return "last_sign_in"
}
