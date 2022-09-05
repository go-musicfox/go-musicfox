package storage

import "go-musicfox/constants"

type LastSignIn struct {}

func (p LastSignIn) GetDbName() string {
	return constants.AppDBName
}

func (p LastSignIn) GetTableName() string {
	return "default_bucket"
}

func (p LastSignIn) GetKey() string {
	return "last_sign_in"
}
