package storage

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/pkg/errors"
	"go.etcd.io/bbolt"

	"github.com/go-musicfox/go-musicfox/utils/app"
)

type LocalDB struct {
	*bbolt.DB
	isTemporary bool
	path        string
}

func (m *LocalDB) Close() error {
	err := m.DB.Close()
	if err != nil {
		return err
	}
	if m.isTemporary {
		err := os.Remove(m.path)
		if err != nil {
			return err
		}
	}
	return nil
}

// NewLocalDB 创建本地数据库
func NewLocalDB(dbName string) (*LocalDB, error) {
	dbDir := app.DBDir()
	if _, err := os.Stat(dbDir); err != nil {
		_ = os.MkdirAll(dbDir, 0755)
	}

	temporaryDB := false
	path := fmt.Sprintf("%s/%s.db", dbDir, dbName)
	options := bbolt.DefaultOptions
	options.Timeout = 500 * time.Millisecond

	for {
		boltDB, err := bbolt.Open(path, 0600, options)
		if err == nil {
			// Success. Just return the DB.
			db := &LocalDB{
				DB:          boltDB,
				isTemporary: temporaryDB,
				path:        path,
			}
			return db, nil
		}
		// If the default database can't be opened because of a timeout, it is likely because another instance
		// of musicfox is already running and has a file lock on the default database. This error is recoverable, and
		// we can try copy the content of the database to a temporary file and open the temporary file as our database
		// instead. Otherwise, it is not possible to recover from this error. We just return nil.
		recoverableError := errors.Is(err, bbolt.ErrTimeout) && !temporaryDB
		if !recoverableError {
			return nil, err
		}
		sourceFile, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer sourceFile.Close()
		targetFile, err := os.CreateTemp("", fmt.Sprintf("%s*.db", dbName))
		if err != nil {
			return nil, err
		}
		defer targetFile.Close()
		_, err = io.Copy(targetFile, sourceFile)
		if err != nil {
			return nil, err
		}
		// Try to open the database again with the new file
		path = targetFile.Name()
		temporaryDB = true
	}
}

var DBManager *LocalDBManager

type LocalDBManager struct {
	localDBs map[string]*LocalDB
}

func (dm *LocalDBManager) Close() error {
	for _, db := range dm.localDBs {
		err := db.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// GetDBFromCache 从缓存中获取 LocalDB
func (dm *LocalDBManager) GetDBFromCache(db any) (localDB *LocalDB, err error) {
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
