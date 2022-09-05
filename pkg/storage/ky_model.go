package storage

type KVModel interface {
	Model
	GetKey() string
}
