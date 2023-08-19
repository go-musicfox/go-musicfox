package storage

import (
	"encoding/json"

	"github.com/go-musicfox/go-musicfox/utils"
	"github.com/pkg/errors"
	"go.etcd.io/bbolt"
)

type IteratorCallback func(k, v []byte) error

type table struct{}

// NewTable 创建table
func NewTable() *table {
	return &table{}
}

// AllMap traverse all data
func (table *table) AllMap(model Model, callback IteratorCallback) (err error) {
	localDB, err := DBManager.GetDBFromCache(model)
	if err != nil {
		return
	}

	err = localDB.View(func(tx *bbolt.Tx) (err error) {
		bucketName := model.GetTableName()
		bucket := tx.Bucket([]byte(bucketName))
		if err = checkBucket(bucket, bucketName); err != nil {
			return err
		}

		err = bucket.ForEach(func(k, v []byte) error {
			return callback(k, v)
		})
		return
	})
	return
}

// IncrAdd add one line, return increment id
func (table *table) IncrAdd(model Model, data IDSetter) (id uint64, err error) {
	localDB, err := DBManager.GetDBFromCache(model)
	if err != nil {
		return
	}

	err = localDB.Update(func(tx *bbolt.Tx) error {
		bucketName := model.GetTableName()
		bucket, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			return err
		}

		id, err = bucket.NextSequence()
		if err != nil {
			return err
		}
		data.SetID(id)

		buf, err := json.Marshal(data)
		if err != nil {
			return err
		}

		// Persist bytes to users bucket.
		return bucket.Put(utils.IDToBin(id), buf)
	})

	return id, err
}

// Set edit one line
func (table *table) Set(model Model, key []byte, data interface{}) (err error) {
	localDB, err := DBManager.GetDBFromCache(model)
	if err != nil {
		return
	}

	err = localDB.Update(func(tx *bbolt.Tx) error {
		bucketName := model.GetTableName()
		bucket, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			return err
		}

		buf, err := json.Marshal(data)
		if err != nil {
			return err
		}

		// Persist bytes to users bucket.
		return bucket.Put(key, buf)
	})

	return
}

// SetByID edit one line by ID
func (table *table) SetByID(model Model, ID uint64, data interface{}) error {
	return table.Set(model, utils.IDToBin(ID), data)
}

// SetByKVModel edit one line by KVModel
func (table *table) SetByKVModel(model KVModel, data interface{}) error {
	return table.Set(model, []byte(model.GetKey()), data)
}

// Delete delete one line
func (table *table) Delete(model Model, key []byte) (err error) {
	localDB, err := DBManager.GetDBFromCache(model)
	if err != nil {
		return
	}

	err = localDB.Update(func(tx *bbolt.Tx) error {
		bucketName := model.GetTableName()
		bucket, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			return err
		}

		return bucket.Delete(key)
	})

	return
}

// DeleteByID delete one line by ID
func (table *table) DeleteByID(model Model, ID uint64) error {
	return table.Delete(model, utils.IDToBin(ID))
}

// DeleteByKVModel delete one line by KVModel
func (table *table) DeleteByKVModel(model KVModel) error {
	return table.Delete(model, []byte(model.GetKey()))
}

// Get 通过key获取value
func (table *table) Get(model Model, key []byte) (value []byte, err error) {
	db, err := DBManager.GetDBFromCache(model)
	if err != nil {
		return
	}

	err = db.View(func(tx *bbolt.Tx) error {
		bucketName := model.GetTableName()
		bucket := tx.Bucket([]byte(bucketName))
		if err = checkBucket(bucket, bucketName); err != nil {
			return err
		}

		value = bucket.Get(key)
		return nil
	})
	return
}

// GetByID 通过ID获取value
func (table *table) GetByID(model Model, ID uint64) ([]byte, error) {
	return table.Get(model, utils.IDToBin(ID))
}

// GetByKVModel 通过KVModel获取value
func (table *table) GetByKVModel(model KVModel) ([]byte, error) {
	return table.Get(model, []byte(model.GetKey()))
}

func checkBucket(bucket *bbolt.Bucket, bucketName string) error {
	if bucket == nil {
		return errors.Errorf("Bucket(%s) not exists!", bucketName)
	}
	return nil
}
