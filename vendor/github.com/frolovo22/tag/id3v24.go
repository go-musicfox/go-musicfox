package tag

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

type id3v24Flags byte

func (flags id3v24Flags) String() string {
	return strconv.Itoa(int(flags))
}

func (flags id3v24Flags) IsUnsynchronisation() bool {
	return GetBit(byte(flags), 7) == 1
}

func (flags id3v24Flags) SetUnsynchronisation(data bool) {
	SetBit((*byte)(&flags), data, 7)
}

func (flags id3v24Flags) HasExtendedHeader() bool {
	return GetBit(byte(flags), 6) == 1
}

func (flags id3v24Flags) SetExtendedHeader(data bool) {
	SetBit((*byte)(&flags), data, 7)
}

func (flags id3v24Flags) IsExperimentalIndicator() bool {
	return GetBit(byte(flags), 5) == 1
}

func (flags id3v24Flags) SetExperimentalIndicator(data bool) {
	SetBit((*byte)(&flags), data, 7)
}

type ID3v24Frame struct {
	Key   string
	Value []byte
}

type ID3v24 struct {
	Marker     string // Always 'ID3'
	Version    Version
	SubVersion int
	Flags      id3v24Flags
	Length     int
	Frames     []ID3v24Frame

	dataStartPos int64
	dataSource   io.ReadSeekCloser
}

type AttachedPicture struct {
	MIME        string
	PictureType byte
	Description string
	Data        []byte
}

func (id3v2 *ID3v24) GetAllTagNames() []string {
	var result []string
	for i := range id3v2.Frames {
		result = append(result, id3v2.Frames[i].Key)
	}
	return result
}

func (id3v2 *ID3v24) GetVersion() Version {
	return id3v2.Version
}

func (id3v2 *ID3v24) GetFileData() []byte {
	_, _ = id3v2.dataSource.Seek(id3v2.dataStartPos, io.SeekStart)
	data, _ := io.ReadAll(id3v2.dataSource)
	return data
}

func (id3v2 *ID3v24) GetTitle() (string, error) {
	return id3v2.GetString("TIT2")
}

func (id3v2 *ID3v24) GetArtist() (string, error) {
	return id3v2.GetString("TPE1")
}

func (id3v2 *ID3v24) GetAlbum() (string, error) {
	return id3v2.GetString("TALB")
}

func (id3v2 *ID3v24) GetYear() (int, error) {
	date, err := id3v2.GetTimestamp("TDOR")
	return date.Year(), err
}

func (id3v2 *ID3v24) GetComment() (string, error) {
	// id3v2
	// Comment struct must be greater than 4
	// [lang \x00 text] - comment format
	// lang - 3 symbols
	// \x00 - const, delimeter
	// text - all after
	commentStr, err := id3v2.GetString("COMM")
	if err != nil {
		return "", err
	}

	if len(commentStr) < 4 {
		return "", ErrIncorrectLength
	}

	return commentStr[4:], nil
}

func (id3v2 *ID3v24) GetGenre() (string, error) {
	return id3v2.GetString("TCON")
}

func (id3v2 *ID3v24) GetAlbumArtist() (string, error) {
	return id3v2.GetString("TPE2")
}

func (id3v2 *ID3v24) GetDate() (time.Time, error) {
	return id3v2.GetTimestamp("TDRC")
}

func (id3v2 *ID3v24) GetArranger() (string, error) {
	return id3v2.GetString("TIPL")
}

func (id3v2 *ID3v24) GetAuthor() (string, error) {
	return id3v2.GetString("TOLY")
}

func (id3v2 *ID3v24) GetBPM() (int, error) {
	return id3v2.GetInt("TBPM")
}

func (id3v2 *ID3v24) GetCatalogNumber() (string, error) {
	return id3v2.GetStringTXXX("CATALOGNUMBER")
}

func (id3v2 *ID3v24) GetCompilation() (string, error) {
	return id3v2.GetString("TCMP")
}

func (id3v2 *ID3v24) GetComposer() (string, error) {
	return id3v2.GetString("TCOM")
}

func (id3v2 *ID3v24) GetConductor() (string, error) {
	return id3v2.GetString("TPE3")
}

func (id3v2 *ID3v24) GetCopyright() (string, error) {
	return id3v2.GetString("TCOP")
}

func (id3v2 *ID3v24) GetDescription() (string, error) {
	return id3v2.GetString("TIT3")
}

func (id3v2 *ID3v24) GetDiscNumber() (int, int, error) {
	dickNumber, err := id3v2.GetString("TPOS")
	if err != nil {
		return 0, 0, err
	}
	numbers := strings.Split(dickNumber, "/")
	if len(numbers) != 2 {
		return 0, 0, ErrIncorrectLength
	}
	number, err := strconv.Atoi(numbers[0])
	if err != nil {
		return 0, 0, err
	}
	total, err := strconv.Atoi(numbers[1])
	if err != nil {
		return 0, 0, err
	}
	return number, total, nil
}

func (id3v2 *ID3v24) GetEncodedBy() (string, error) {
	return id3v2.GetString("TENC")
}

func (id3v2 *ID3v24) GetTrackNumber() (int, int, error) {
	track, err := id3v2.GetInt("TRCK")
	return track, track, err
}

func (id3v2 *ID3v24) GetPicture() (image.Image, error) {
	pic, err := id3v2.GetAttachedPicture()
	if err != nil {
		return nil, err
	}
	switch pic.MIME {
	case mimeImageJPEG:
		return jpeg.Decode(bytes.NewReader(pic.Data))
	case mimeImagePNG:
		return png.Decode(bytes.NewReader(pic.Data))
	case mimeImageLink:
		return downloadImage(string(pic.Data))
	default:
		return nil, ErrIncorrectTag
	}
}

func (id3v2 *ID3v24) SetTitle(title string) error {
	return id3v2.SetString("TIT2", title)
}

func (id3v2 *ID3v24) SetArtist(artist string) error {
	return id3v2.SetString("TPE1", artist)
}

func (id3v2 *ID3v24) SetAlbum(album string) error {
	return id3v2.SetString("TALB", album)
}

func (id3v2 *ID3v24) SetYear(year int) error {
	curDate, err := id3v2.GetTimestamp("TDOR")
	if err != nil {
		// set only year
		return id3v2.SetTimestamp(
			"TDOR",
			time.Date(year, 0, 0, 0, 0, 0, 0, time.Local),
		)
	}
	return id3v2.SetTimestamp(
		"TDOR",
		time.Date(
			year,
			curDate.Month(),
			curDate.Day(),
			curDate.Hour(),
			curDate.Minute(),
			curDate.Second(),
			curDate.Nanosecond(),
			curDate.Location(),
		),
	)
}

func (id3v2 *ID3v24) SetComment(comment string) error {
	return id3v2.SetString("COMM", comment)
}

func (id3v2 *ID3v24) SetGenre(genre string) error {
	return id3v2.SetString("TCON", genre)
}

func (id3v2 *ID3v24) SetAlbumArtist(albumArtist string) error {
	return id3v2.SetString("TPE2", albumArtist)
}

func (id3v2 *ID3v24) SetDate(date time.Time) error {
	return id3v2.SetTimestamp("TDRC", date)
}

func (id3v2 *ID3v24) SetArranger(arranger string) error {
	return id3v2.SetString("IPLS", arranger)
}

func (id3v2 *ID3v24) SetAuthor(author string) error {
	return id3v2.SetString("TOLY", author)
}

func (id3v2 *ID3v24) SetBPM(bmp int) error {
	return id3v2.SetInt("TBMP", bmp)
}

func (id3v2 *ID3v24) SetCatalogNumber(catalogNumber string) error {
	return id3v2.SetString("TXXX", catalogNumber)
}

func (id3v2 *ID3v24) SetCompilation(compilation string) error {
	return id3v2.SetString("TCMP", compilation)
}

func (id3v2 *ID3v24) SetComposer(composer string) error {
	return id3v2.SetString("TCOM", composer)
}

func (id3v2 *ID3v24) SetConductor(conductor string) error {
	return id3v2.SetString("TPE3", conductor)
}

func (id3v2 *ID3v24) SetCopyright(copyright string) error {
	return id3v2.SetString("TCOP", copyright)
}

func (id3v2 *ID3v24) SetDescription(description string) error {
	return id3v2.SetString("TIT3", description)
}

func (id3v2 *ID3v24) SetDiscNumber(number int, total int) error {
	return id3v2.SetString("TPOS", fmt.Sprintf("%d/%d", number, total))
}

func (id3v2 *ID3v24) SetEncodedBy(encodedBy string) error {
	return id3v2.SetString("TENC", encodedBy)
}

func (id3v2 *ID3v24) SetTrackNumber(number int, total int) error {
	// only number
	return id3v2.SetInt("TRCK", number)
}

func (id3v2 *ID3v24) SetPicture(picture image.Image) error {
	// Only PNG
	buf := new(bytes.Buffer)
	err := png.Encode(buf, picture)
	if err != nil {
		return err
	}

	attacheched, err := id3v2.GetAttachedPicture()
	if err != nil {
		// Set default params
		newPicture := AttachedPicture{
			MIME:        "image/png",
			PictureType: 2, // Other file info
			Description: "",
			Data:        buf.Bytes(),
		}
		return id3v2.SetAttachedPicture(&newPicture)
	}
	// save metainfo
	attacheched.MIME = mimeImagePNG
	attacheched.Data = buf.Bytes()

	return id3v2.SetAttachedPicture(attacheched)
}

func (id3v2 *ID3v24) DeleteAll() error {
	id3v2.Frames = []ID3v24Frame{}
	return nil
}

func (id3v2 *ID3v24) DeleteTitle() error {
	return id3v2.DeleteTag("TIT2")
}

func (id3v2 *ID3v24) DeleteArtist() error {
	return id3v2.DeleteTag("TPE1")
}

func (id3v2 *ID3v24) DeleteAlbum() error {
	return id3v2.DeleteTag("TALB")
}

func (id3v2 *ID3v24) DeleteYear() error {
	return id3v2.DeleteTag("TDOR")
}

func (id3v2 *ID3v24) DeleteComment() error {
	return id3v2.DeleteTag("COMM")
}

func (id3v2 *ID3v24) DeleteGenre() error {
	return id3v2.DeleteTag("TCON")
}

func (id3v2 *ID3v24) DeleteAlbumArtist() error {
	return id3v2.DeleteTag("TPE2")
}

func (id3v2 *ID3v24) DeleteDate() error {
	return id3v2.DeleteTag("TDRC")
}

func (id3v2 *ID3v24) DeleteArranger() error {
	return id3v2.DeleteTag("IPLS")
}

func (id3v2 *ID3v24) DeleteAuthor() error {
	return id3v2.DeleteTag("TOLY")
}

func (id3v2 *ID3v24) DeleteBPM() error {
	return id3v2.DeleteTag("TBMP")
}

func (id3v2 *ID3v24) DeleteCatalogNumber() error {
	return id3v2.DeleteTagTXXX("CATALOGNUMBER")
}

func (id3v2 *ID3v24) DeleteCompilation() error {
	return id3v2.DeleteTag("TCMP")
}

func (id3v2 *ID3v24) DeleteComposer() error {
	return id3v2.DeleteTag("TCOM")
}

func (id3v2 *ID3v24) DeleteConductor() error {
	return id3v2.DeleteTag("TPE3")
}

func (id3v2 *ID3v24) DeleteCopyright() error {
	return id3v2.DeleteTag("TCOP")
}

func (id3v2 *ID3v24) DeleteDescription() error {
	return id3v2.DeleteTag("TIT3")
}

func (id3v2 *ID3v24) DeleteDiscNumber() error {
	return id3v2.DeleteTag("TPOS")
}

func (id3v2 *ID3v24) DeleteEncodedBy() error {
	return id3v2.DeleteTag("TENC")
}

func (id3v2 *ID3v24) DeleteTrackNumber() error {
	return id3v2.DeleteTag("TRCK")
}

func (id3v2 *ID3v24) DeletePicture() error {
	return id3v2.DeleteTag("APIC")
}

func (id3v2 *ID3v24) SaveFile(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return id3v2.Save(file)
}

func (id3v2 *ID3v24) Save(input io.WriteSeeker) error {
	// write header
	err := id3v2.writeHeaderID3v24(input)
	if err != nil {
		return err
	}

	// write tags
	err = id3v2.writeFramesID3v24(input)
	if err != nil {
		return err
	}

	// write data
	// write data
	if _, err = id3v2.dataSource.Seek(id3v2.dataStartPos, io.SeekStart); err != nil {
		return err
	}
	if _, err = io.Copy(input, id3v2.dataSource); err != nil {
		return err
	}
	return nil
}

func (id3v2 *ID3v24) writeHeaderID3v24(writer io.Writer) error {
	headerByte := make([]byte, 10)

	// ID3
	copy(headerByte[0:3], id3MarkerValue)

	// Version, Subversion, Flags
	copy(headerByte[3:6], []byte{4, 0, 0})

	// Length
	length := id3v2.getFramesLength()
	lengthByte := IntToByteSynchsafe(length)
	copy(headerByte[6:10], lengthByte)

	nWriten, err := writer.Write(headerByte)
	if err != nil {
		return err
	}
	if nWriten != 10 {
		return ErrWriting
	}
	return nil
}

func (id3v2 *ID3v24) writeFramesID3v24(writer io.Writer) error {
	for i := range id3v2.Frames {
		header := make([]byte, 10)

		// Frame id
		copy(header, id3v2.Frames[i].Key)

		// Frame size
		length := len(id3v2.Frames[i].Value)
		header[4] = byte(length >> 24)
		header[5] = byte(length >> 16)
		header[6] = byte(length >> 8)
		header[7] = byte(length)

		// write header
		_, err := writer.Write(header)
		if err != nil {
			return err
		}

		// write data
		_, err = writer.Write(id3v2.Frames[i].Value)
		if err != nil {
			return err
		}
	}

	return nil
}

func (id3v2 *ID3v24) getFramesLength() int {
	result := 0
	for i := range id3v2.Frames {
		// 10 - size of tag header
		result += 10 + len(id3v2.Frames[i].Value)
	}
	return result
}

func (id3v2 *ID3v24) String() string {
	result := "Marker: " + id3v2.Marker + "\n" +
		"Version: " + id3v2.Version.String() + "\n" +
		"Subversion: " + strconv.Itoa(id3v2.SubVersion) + "\n" +
		"Flags: " + id3v2.Flags.String() + "\n" +
		"Length: " + strconv.Itoa(id3v2.Length) + "\n"

	for i := range id3v2.Frames {
		result += id3v2.Frames[i].Key + ": " + string(id3v2.Frames[i].Value) + "\n"
	}

	return result
}

func checkID3v24(input io.ReadSeeker) bool {
	if input == nil {
		return false
	}

	// read marker (3 bytes) and version (1 byte) for ID3v2
	data, err := seekAndRead(input, 0, io.SeekStart, 4)
	if err != nil {
		return false
	}
	marker := string(data[0:3])

	// id3v2
	if marker != id3MarkerValue {
		return false
	}

	versionByte := data[3]
	return versionByte == 4
}

func ReadID3v24(input io.ReadSeekCloser) (*ID3v24, error) {
	header := ID3v24{
		dataSource: input,
	}
	if input == nil {
		return nil, ErrEmptyFile
	}

	// Header size
	headerByte, err := seekAndRead(input, 0, io.SeekStart, 10)
	if err != nil {
		return nil, err
	}

	// Marker
	marker := string(headerByte[0:3])
	if marker != "ID3" {
		return nil, errors.New("error file marker")
	}
	header.Marker = marker

	// Version
	versionByte := headerByte[3]
	if versionByte != 4 {
		return nil, ErrUnsupportedFormat
	}
	header.Version = VersionID3v24

	// Sub version
	subVersionByte := headerByte[4]
	header.SubVersion = int(subVersionByte)

	// Flags
	header.Flags = id3v24Flags(headerByte[5])

	// Length
	length := ByteToIntSynchsafe(headerByte[6:10])
	header.Length = length

	// Extended headers
	header.Frames = []ID3v24Frame{}
	curRead := 0
	for curRead < length {
		var bytesExtendedHeader []byte
		bytesExtendedHeader, err = readBytes(input, 10)
		if err != nil {
			return nil, err
		}

		// Frame identifier
		key := string(bytesExtendedHeader[0:4])

		/*if bytesExtendedHeader[0] == 0 &&
		bytesExtendedHeader[1] == 0 &&
		bytesExtendedHeader[2] == 0 &&
		bytesExtendedHeader[3] == 0 {
			break
		}*/

		// Frame data size
		size := ByteToInt(bytesExtendedHeader[4:8])

		var bytesExtendedValue []byte
		bytesExtendedValue, err = readBytes(input, size)
		if err != nil {
			return nil, err
		}

		header.Frames = append(header.Frames, ID3v24Frame{
			key,
			bytesExtendedValue,
		})

		curRead += 10 + size
	}

	// TODO
	if curRead != length {
		return nil, errors.New("error extended frames")
	}

	// file data
	header.dataStartPos, err = input.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, err
	}

	return &header, nil
}

func (id3v2 *ID3v24) GetString(name string) (string, error) {
	for i := range id3v2.Frames {
		if id3v2.Frames[i].Key == name {
			return GetString(id3v2.Frames[i].Value)
		}
	}
	return "", ErrTagNotFound
}

func (id3v2 *ID3v24) SetString(name string, value string) error {
	frame := ID3v24Frame{
		Key:   name,
		Value: SetString(value),
	}

	for i := range id3v2.Frames {
		if id3v2.Frames[i].Key == name {
			id3v2.Frames[i].Value = frame.Value
			return nil
		}
	}

	id3v2.Frames = append(id3v2.Frames, frame)
	return nil
}

func (id3v2 *ID3v24) GetTimestamp(name string) (time.Time, error) {
	str, err := id3v2.GetString(name)
	if err != nil {
		return time.Now(), err
	}
	result, err := time.Parse("2006-01-02T15:04:05", str)
	if err != nil {
		return time.Now(), err
	}
	return result, nil
}

func (id3v2 *ID3v24) SetTimestamp(name string, value time.Time) error {
	str := value.Format("2006-01-02T15:04:05")
	return id3v2.SetString(name, str)
}

func (id3v2 *ID3v24) GetInt(name string) (int, error) {
	intStr, err := id3v2.GetString(name)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(intStr)
}

func (id3v2 *ID3v24) SetInt(name string, value int) error {
	return id3v2.SetString(name, strconv.Itoa(value))
}

func (id3v2 *ID3v24) GetAttachedPicture() (*AttachedPicture, error) {
	var picture AttachedPicture

	picStr, err := id3v2.GetString("APIC")
	if err != nil {
		return nil, err
	}
	values := strings.SplitN(picStr, "\x00", 3)
	if len(values) != 3 {
		return nil, ErrIncorrectTag
	}

	// MIME
	picture.MIME = values[0]

	// Type
	if len(values[1]) == 0 {
		return nil, ErrIncorrectTag
	}
	picture.PictureType = values[1][0]

	// Description
	picture.Description = values[1][1:]

	// Image data
	picture.Data = []byte(values[2])

	return &picture, nil
}

// nolint:gocritic
func (id3v2 *ID3v24) SetAttachedPicture(picture *AttachedPicture) error {
	// set UTF-8
	result := []byte{0}

	// MIME type
	result = append(result, []byte(picture.MIME)...)
	result = append(result, 0x00)

	// Picture type
	result = append(result, picture.PictureType)

	// Picture description
	result = append(result, []byte(picture.Description)...)
	result = append(result, 0x00)

	// Picture data
	result = append(result, picture.Data...)

	return id3v2.SetString("APIC", string(result))
}

func (id3v2 *ID3v24) DeleteTag(name string) error {
	index := -1
	for i := range id3v2.Frames {
		if id3v2.Frames[i].Key == name {
			index = i
			break
		}
	}
	// already deleted
	if index == -1 {
		return nil
	}
	id3v2.Frames = append(id3v2.Frames[:index], id3v2.Frames[index+1:]...)
	return nil
}

func (id3v2 *ID3v24) DeleteTagTXXX(name string) error {
	index := -1
	for i := range id3v2.Frames {
		if id3v2.Frames[i].Key == id3v2FrameTXXX {
			str, err := GetString(id3v2.Frames[i].Value)
			if err != nil {
				return err
			}
			info := strings.SplitN(str, "\x00", 2)
			if len(info) != 2 {
				return ErrIncorrectTag
			}
			if info[0] == name {
				index = i
				break
			}
		}
	}
	// already deleted
	if index == -1 {
		return nil
	}

	id3v2.Frames = append(id3v2.Frames[:index], id3v2.Frames[index+1:]...)
	return nil
}

// GetStringTXXX - get user frame
// Header for 'User defined text information frame'
// Text encoding     $xx
// Description       <text string according to encoding> $00 (00)
// Value             <text string according to encoding>.
func (id3v2 *ID3v24) GetStringTXXX(name string) (string, error) {
	for i := range id3v2.Frames {
		if id3v2.Frames[i].Key == id3v2FrameTXXX {
			str, err := GetString(id3v2.Frames[i].Value)
			if err != nil {
				return "", err
			}
			info := strings.SplitN(str, "\x00", 2)
			if len(info) != 2 {
				return "", ErrIncorrectTag
			}
			if info[0] == name {
				return info[1], nil
			}
		}
	}
	return "", ErrTagNotFound
}

func (id3v2 *ID3v24) SetStringTXXX(name string, value string) error {
	result := ID3v24Frame{
		Key:   id3v2FrameTXXX,
		Value: SetString(name + "\x00" + value),
	}

	// find tag
	for i := range id3v2.Frames {
		if id3v2.Frames[i].Key == id3v2FrameTXXX {
			str, err := GetString(id3v2.Frames[i].Value)
			if err != nil {
				continue
			}

			info := strings.SplitN(str, "\x00", 2)
			if len(info) != 2 {
				continue
			}

			if info[0] == name {
				id3v2.Frames[i] = result
				return nil
			}
		}
	}

	id3v2.Frames = append(id3v2.Frames, result)
	return nil
}

func (id3v2 *ID3v24) GetIntTXXX(name string) (int, error) {
	str, err := id3v2.GetStringTXXX(name)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(str)
}

func (id3v2 *ID3v24) Close() error {
	return id3v2.dataSource.Close()
}
