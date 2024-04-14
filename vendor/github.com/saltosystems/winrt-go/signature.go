package winrt

import (
	"crypto/sha1" // #nosec this is not used for security purposes
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/go-ole/go-ole"
)

// Primitive types signatures
const (
	SignatureUInt8   = "u1"
	SignatureUInt16  = "u2"
	SignatureUInt32  = "u4"
	SignatureUInt64  = "u8"
	SignatureInt8    = "i1"
	SignatureInt16   = "i2"
	SignatureInt32   = "i4"
	SignatureInt64   = "i8"
	SignatureFloat32 = "f4"
	SignatureFloat64 = "f8"
	SignatureBool    = "b1"
	SignatureChar    = "c2"
	SignatureString  = "string"
	SignatureGUID    = "g16"
)

// ParameterizedInstanceGUID creates a `GUID` for a "generic" WinRT delegate or interface. This was ported from the RUST implementation
// of WinRT, checkout for the source code:
// https://github.com/microsoft/windows-rs/blob/68576f37df4c02f09bc6e4dd1ed8ed8844c6eb9c/crates/libs/windows/src/core/guid.rs#L44
//
// Checkout the following link for documentation on how the signatures are generated:
// https://docs.microsoft.com/en-us/uwp/winrt-cref/winrt-type-system#guid-generation-for-parameterized-types
func ParameterizedInstanceGUID(baseGUID string, signatures ...string) string {
	res := fmt.Sprintf("pinterface({%s};%s)", baseGUID, strings.Join(signatures, ";"))
	return guidFromSignature(res)
}

func guidFromSignature(signature string) string {
	// base wrt_pinterface_namespace => 11f47ad5-7b73-42c0-abae-878b1e16adee
	data := []byte{0x11, 0xf4, 0x7a, 0xd5, 0x7b, 0x73, 0x42, 0xc0, 0xab, 0xae, 0x87, 0x8b, 0x1e, 0x16, 0xad, 0xee}

	data = append(data, []byte(signature)...)

	hash := sha1.New() // #nosec this is not used for security purposes
	if _, err := hash.Write(data); err != nil {
		return "_ERROR_"
	}

	bytes := hash.Sum(nil)
	first := binary.BigEndian.Uint32(bytes[0:4])
	second := binary.BigEndian.Uint16(bytes[4:6])
	third := binary.BigEndian.Uint16(bytes[6:8])
	third = (third & 0x0fff) | (5 << 12)
	fourth := (bytes[8] & 0x3f) | 0x80

	guid := guidFromValues(first, second, third, [8]byte{fourth, bytes[9], bytes[10], bytes[11], bytes[12], bytes[13], bytes[14], bytes[15]})
	return guid.String()
}

func guidFromValues(first uint32, second, third uint16, rest [8]byte) ole.GUID {
	return ole.GUID{
		Data1: first,
		Data2: second,
		Data3: third,
		Data4: rest,
	}
}
