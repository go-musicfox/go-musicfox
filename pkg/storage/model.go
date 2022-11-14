package storage

type Model interface {
	GetDbName() string
	GetTableName() string
}

// IDSetter set id
type IDSetter interface {
	SetID(ID uint64)
}
