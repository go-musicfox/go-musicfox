package tag

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"io"
	"net/http"
	"unicode/utf16"
	"unicode/utf8"
)

func seekAndRead(input io.ReadSeeker, offset int64, whence int, read int) ([]byte, error) {
	if input == nil {
		return nil, ErrEmptyFile
	}

	_, err := input.Seek(offset, whence)
	if err != nil {
		return nil, ErrSeekFile
	}

	data := make([]byte, read)
	nReaded, err := input.Read(data)
	if err != nil {
		return nil, err
	}
	if nReaded != read {
		return nil, ErrReadFile
	}

	return data, nil
}

func seekAndReadString(input io.ReadSeeker, offset int64, whence int, read int) (string, error) {
	data, err := seekAndRead(input, offset, whence, read)
	return string(data), err
}

func readBytes(input io.Reader, size int) ([]byte, error) {
	if input == nil {
		return nil, ErrEmptyFile
	}

	data := make([]byte, size)
	nReaded, err := input.Read(data)
	if err != nil {
		return nil, err
	}

	if nReaded != size {
		return nil, ErrReadFile
	}

	return data, nil
}

func GetEncoding(code byte) string {
	if code == 0 || code == 3 {
		return encodingUTF8
	}
	if code == 1 {
		return encodingUTF16
	}
	if code == 2 {
		return encodingUTF16BE
	}
	return ""
}

// TextEncoding -
// Text Encoding for text frame header
// First byte determinate text encoding.
// If ISO-8859-1 is used this byte should be $00, if Unicode is used it should be $01
// Return text encoding. E.g. "utf8", "utf16", etc.
func TextEncoding(b []byte) string {
	if len(b) == 0 {
		return ""
	}

	return GetEncoding(b[0])
}

func DecodeString(b []byte, encoding string) (string, error) {
	switch encoding {
	case encodingUTF8:
		return string(b), nil
	case encodingUTF16:
		value, err := DecodeUTF16(b)
		if err != nil {
			return "", err
		}
		return value, nil
	case encodingUTF16BE:
		return DecodeUTF16BE(b)
	}

	return "", ErrEncodingFormat
}

// Decode UTF-16 Little Endian to UTF-8.
func DecodeUTF16(b []byte) (string, error) {
	if len(b)%2 != 0 {
		return "", ErrDecodeEvenLength
	}

	u16s := make([]uint16, 1)

	ret := &bytes.Buffer{}

	b8buf := make([]byte, 4)

	lb := len(b)
	for i := 0; i < lb; i += 2 {
		u16s[0] = uint16(b[i]) + (uint16(b[i+1]) << 8)
		r := utf16.Decode(u16s)
		n := utf8.EncodeRune(b8buf, r[0])
		ret.Write(b8buf[:n])
	}

	return ret.String(), nil
}

// Decode UTF-16 Big Endian To UTF-8.
func DecodeUTF16BE(b []byte) (string, error) {
	if len(b)%2 != 0 {
		return "", ErrDecodeEvenLength
	}

	u16s := make([]uint16, 1)

	ret := &bytes.Buffer{}

	b8buf := make([]byte, 4)

	lb := len(b)
	for i := 0; i < lb; i += 2 {
		u16s[0] = uint16(b[i+1]) + (uint16(b[i]) << 8)
		r := utf16.Decode(u16s)
		n := utf8.EncodeRune(b8buf, r[0])
		ret.Write(b8buf[:n])
	}

	return ret.String(), nil
}

// ByteToIntSynchsafe -
// Convert byte to int
// In some parts of the tag it is inconvenient to use the
// unsychronisation scheme because the size of unsynchronised data is
// not known in advance, which is particularly problematic with size
// descriptors. The solution in ID3v2 is to use synchsafe integers, in
// which there can never be any false synchs. Synchsafe integers are
// integers that keep its highest bit (bit 7) zeroed, making seven bits
// out of eight available. Thus a 32 bit synchsafe integer can store 28
// bits of information.
func ByteToIntSynchsafe(data []byte) int {
	result := 0
	for _, b := range data {
		result = (result << 7) | int(b)
	}
	return result
}

func IntToByteSynchsafe(data int) []byte {
	// 7F = 0111 1111
	return []byte{
		byte(data>>23) & 0x7F,
		byte(data>>15) & 0x7F,
		byte(data>>7) & 0x7F,
		byte(data) & 0x7F,
	}
}

// Convert byte to int.
func ByteToInt(data []byte) int {
	result := 0
	for _, b := range data {
		result = (result << 8) | int(b)
	}
	return result
}

// Return bit value
// Index starts from 0
// bits order [7,6,5,4,3,2,1,0].
func GetBit(data byte, index byte) byte {
	return 1 & (data >> index)
}

func SetBit(data *byte, bit bool, index byte) {
	if bit {
		*data |= 1 << index
	} else {
		*data &= ^(1 << index)
	}
}

func GetString(b []byte) (string, error) {
	if len(b) < 2 {
		return "", ErrIncorrectTag
	}
	return DecodeString(b[1:], TextEncoding(b))
}

func SetString(value string) []byte {
	// Set UTF-8
	result := []byte{0}
	// Set data
	return append(result, []byte(value)...)
}

// Read format:
// [length, data]
// length in littleIndian.
func readLengthData(input io.Reader, order binary.ByteOrder) ([]byte, error) {
	// length
	var length uint32
	err := binary.Read(input, order, &length)
	if err != nil {
		return nil, err
	}

	// data
	data, err := readBytes(input, int(length))
	if err != nil {
		return nil, err
	}
	return data, nil
}

func writeLengthData(output io.Writer, order binary.ByteOrder, data []byte) error {
	length := uint32(len(data))
	err := binary.Write(output, order, length)
	if err != nil {
		return err
	}

	_, err = output.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func downloadImage(url string) (image.Image, error) {
	// nolint:gosec
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	img, _, err := image.Decode(resp.Body)
	return img, err
}

// nolint:deadcode,unused
func colorModelToBitsPerPixel(model color.Model) int {
	var bpp int
	switch model {
	case color.RGBAModel:
		bpp = 8
	case color.RGBA64Model:
		bpp = 64
	case color.NRGBAModel:
		bpp = 8
	case color.NRGBA64Model:
		bpp = 64
	case color.AlphaModel:
		bpp = 8
	case color.Alpha16Model:
		bpp = 16
	case color.GrayModel:
		bpp = 8
	case color.Gray16Model:
		bpp = 16
	default:
		bpp = 8
	}

	return bpp
}

func SplitBytesWithTextDescription(data []byte, encoding string) [][]byte {
	separator := []byte{0}
	if encoding == encodingUTF16 || encoding == encodingUTF16BE {
		separator = []byte{0, 0}
	}

	result := bytes.SplitN(data, separator, 2)
	if len(result) != 2 {
		return result
	}

	if len(result[1]) == 0 {
		return result
	}

	if result[1][0] == 0 {
		result[0] = append(result[0], result[1][0])
		result[1] = result[1][1:]
	}
	return result
}
