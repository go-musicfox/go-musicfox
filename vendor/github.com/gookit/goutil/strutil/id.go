package strutil

import (
	"strconv"
	"time"

	"github.com/gookit/goutil/mathutil"
)

// global id:
//
//	https://github.com/rs/xid
//	https://github.com/satori/go.uuid
var (
	DefMinInt = 1000
	DefMaxInt = 9999
)

// MicroTimeID generate.
// return like: 16074145697981929446(len: 20)
func MicroTimeID() string {
	ms := time.Now().UnixNano() / 1000
	ri := mathutil.RandomInt(DefMinInt, DefMaxInt)

	return strconv.FormatInt(ms, 10) + strconv.FormatInt(int64(ri), 10)
}

// MicroTimeHexID generate.
// return like: 5b5f0588af1761ad3(len: 16-17)
func MicroTimeHexID() string {
	ms := time.Now().UnixNano() / 1000
	ri := mathutil.RandomInt(DefMinInt, DefMaxInt)

	return strconv.FormatInt(ms, 16) + strconv.FormatInt(int64(ri), 16)
}
