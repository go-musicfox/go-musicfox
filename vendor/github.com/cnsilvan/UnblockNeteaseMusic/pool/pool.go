package pool

import (
	"math/big"
	"sync"
)

var bigIntPool = sync.Pool{
	New: func() interface{} {
		return big.NewInt(0)
	},
}

func GetBigInt() *big.Int {
	b := bigIntPool.Get().(*big.Int)
	b.SetUint64(0)
	return b
}
func PutBigInt(x *big.Int) {
	bigIntPool.Put(x)
}
