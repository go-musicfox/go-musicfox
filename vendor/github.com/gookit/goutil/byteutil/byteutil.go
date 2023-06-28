package byteutil

import (
	"bytes"
	"fmt"
	"math/rand"
	"strconv"
	"time"
	"unsafe"
)

// Random bytes generate
func Random(length int) ([]byte, error) {
	b := make([]byte, length)
	// Note that err == nil only if we read len(b) bytes.
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	return b, nil
}

// FirstLine from command output
func FirstLine(bs []byte) []byte {
	if i := bytes.IndexByte(bs, '\n'); i >= 0 {
		return bs[0:i]
	}
	return bs
}

// StrOrErr convert to string, return empty string on error.
func StrOrErr(bs []byte, err error) (string, error) {
	if err != nil {
		return "", err
	}
	return string(bs), err
}

// SafeString convert to string, return empty string on error.
func SafeString(bs []byte, err error) string {
	if err != nil {
		return ""
	}
	return string(bs)
}

// String unsafe convert bytes to string
func String(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// ToString convert bytes to string
func ToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// AppendAny append any value to byte slice
func AppendAny(dst []byte, v any) []byte {
	if v == nil {
		return append(dst, "<nil>"...)
	}

	switch val := v.(type) {
	case []byte:
		dst = append(dst, val...)
	case string:
		dst = append(dst, val...)
	case int:
		dst = strconv.AppendInt(dst, int64(val), 10)
	case int8:
		dst = strconv.AppendInt(dst, int64(val), 10)
	case int16:
		dst = strconv.AppendInt(dst, int64(val), 10)
	case int32:
		dst = strconv.AppendInt(dst, int64(val), 10)
	case int64:
		dst = strconv.AppendInt(dst, val, 10)
	case uint:
		dst = strconv.AppendUint(dst, uint64(val), 10)
	case uint8:
		dst = strconv.AppendUint(dst, uint64(val), 10)
	case uint16:
		dst = strconv.AppendUint(dst, uint64(val), 10)
	case uint32:
		dst = strconv.AppendUint(dst, uint64(val), 10)
	case uint64:
		dst = strconv.AppendUint(dst, val, 10)
	case float32:
		dst = strconv.AppendFloat(dst, float64(val), 'f', -1, 32)
	case float64:
		dst = strconv.AppendFloat(dst, val, 'f', -1, 64)
	case bool:
		dst = strconv.AppendBool(dst, val)
	case time.Time:
		dst = val.AppendFormat(dst, time.RFC3339)
	case time.Duration:
		dst = strconv.AppendInt(dst, int64(val), 10)
	case error:
		dst = append(dst, val.Error()...)
	case fmt.Stringer:
		dst = append(dst, val.String()...)
	default:
		dst = append(dst, fmt.Sprint(v)...)
	}
	return dst
}

// Cut bytes. like the strings.Cut()
func Cut(bs []byte, sep byte) (before, after []byte, found bool) {
	if i := bytes.IndexByte(bs, sep); i >= 0 {
		return bs[:i], bs[i+1:], true
	}

	before = bs
	return
}
