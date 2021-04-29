package db

type KVModel interface {
	Model
	GetKey() string
}