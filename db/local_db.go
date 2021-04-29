package db

import (
	"errors"
	"fmt"
	"github.com/boltdb/bolt"
	"go-musicfox/utils"
	"os"
)

type LocalDB struct {
	*bolt.DB
}

// NewLocalDB 创建本地数据库
func NewLocalDB(dbName string) (*LocalDB, error) {
	projectPath := utils.GetLocalDataDir()

	dbDir := fmt.Sprintf("%s/db", projectPath)
	if _, err := os.Stat(dbDir); err != nil {
		err = os.MkdirAll(dbDir, 0755)
	}
	path := fmt.Sprintf("%s/%s.db", dbDir, dbName)

	boltDB, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}

	db := &LocalDB{
		DB: boltDB,
	}

	return db, err
}

var DBManager *LocalDBManager

type LocalDBManager struct {
	localDBs map[string]*LocalDB
}

// GetDBFromCache 从缓存中获取 LocalDB
func (dm *LocalDBManager) GetDBFromCache(db interface{}) (localDB *LocalDB, err error) {
	var dbName string
	switch dbWithType := db.(type) {
	case []byte:
		dbName = string(dbWithType)
	case string:
		dbName = dbWithType
	case Model:
		dbName = dbWithType.GetDbName()
	default:
		return nil, errors.New("param(db) expect a string or db.Model")
	}

	if dm.localDBs == nil {
		dm.localDBs = map[string]*LocalDB{}
	}

	localDB, ok := dm.localDBs[dbName]
	if !ok {
		localDB, err = NewLocalDB(dbName)
		if err != nil {
			return nil, err
		}
		dm.localDBs[dbName] = localDB
	}

	return localDB, nil
}