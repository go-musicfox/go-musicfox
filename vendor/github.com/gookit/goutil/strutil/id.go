package strutil

import (
	"hash/crc32"
	"math/rand"
	"os"
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

// DatetimeNo generate. can use for order-no.
//
//   - No prefix, return like: 2023041410484904074285478388(len: 28)
//   - With prefix, return like: prefix2023041410484904074285478388(len: 28 + len(prefix))
func DatetimeNo(prefix string) string {
	nt := time.Now()
	pl := len(prefix)
	bs := make([]byte, 0, 28+pl)
	if pl > 0 {
		bs = append(bs, prefix...)
	}

	// micro datatime
	bs = nt.AppendFormat(bs, "20060102150405.000000")
	bs[14+pl] = '0'

	// host
	name, err := os.Hostname()
	if err != nil {
		name = "default"
	}
	c32 := crc32.ChecksumIEEE([]byte(name)) // eg: 4006367001
	bs = strconv.AppendUint(bs, uint64(c32%99), 10)

	// rand 10000 - 99999
	rand.Seed(nt.UnixNano())
	bs = strconv.AppendInt(bs, 10000+rand.Int63n(89999), 10)

	return string(bs)
}
