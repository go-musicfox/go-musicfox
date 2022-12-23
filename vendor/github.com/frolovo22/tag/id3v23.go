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

type id3v23Flags byte

func (flags id3v23Flags) String() string {
	return strconv.Itoa(int(flags))
}

func (flags id3v23Flags) IsUnsynchronisation() bool {
	return GetBit(byte(flags), 7) == 1
}

func (flags id3v23Flags) SetUnsynchronisation(data bool) {
	SetBit((*byte)(&flags), data, 7)
}

func (flags id3v23Flags) HasExtendedHeader() bool {
	return GetBit(byte(flags), 6) == 1
}

func (flags id3v23Flags) SetExtendedHeader(data bool) {
	SetBit((*byte)(&flags), data, 7)
}

func (flags id3v23Flags) IsExperimentalIndicator() bool {
	return GetBit(byte(flags), 5) == 1
}

func (flags id3v23Flags) SetExperimentalIndicator(data bool) {
	SetBit((*byte)(&flags), data, 7)
}

type ID3v23Frame struct {
	Key   string
	Value []byte
}

type ID3v23 struct {
	Marker     string // Always 'ID3'
	Version    Version
	SubVersion int
	Flags      id3v23Flags
	Length     int
	Frames     []ID3v23Frame

	dataStartPos int64
	dataSource   io.ReadSeekCloser
}

func (id3v2 *ID3v23) GetAllTagNames() []string {
	var result []string
	for i := range id3v2.Frames {
		result = append(result, id3v2.Frames[i].Key)
	}
	return result
}

func (id3v2 *ID3v23) GetVersion() Version {
	return id3v2.Version
}

func (id3v2 *ID3v23) GetFileData() []byte {
	_, _ = id3v2.dataSource.Seek(id3v2.dataStartPos, io.SeekStart)
	data, _ := io.ReadAll(id3v2.dataSource)
	return data
}

func (id3v2 *ID3v23) GetTitle() (string, error) {
	return id3v2.GetString("TIT2")
}

func (id3v2 *ID3v23) GetArtist() (string, error) {
	return id3v2.GetString("TPE1")
}

func (id3v2 *ID3v23) GetAlbum() (string, error) {
	return id3v2.GetString("TALB")
}

func (id3v2 *ID3v23) GetYear() (int, error) {
	date, err := id3v2.GetTimestamp("TYEAR")
	return date.Year(), err
}

func (id3v2 *ID3v23) GetComment() (string, error) {
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

func (id3v2 *ID3v23) GetGenre() (string, error) {
	return id3v2.GetString("TCON")
}

func (id3v2 *ID3v23) GetAlbumArtist() (string, error) {
	return id3v2.GetString("TPE2")
}

func (id3v2 *ID3v23) GetDate() (time.Time, error) {
	return id3v2.GetTimestamp("TYEAR")
}

func (id3v2 *ID3v23) GetArranger() (string, error) {
	return id3v2.GetString("IPLS")
}

func (id3v2 *ID3v23) GetAuthor() (string, error) {
	return id3v2.GetString("TOLY")
}

func (id3v2 *ID3v23) GetBPM() (int, error) {
	return id3v2.GetInt("TBPM")
}

func (id3v2 *ID3v23) GetCatalogNumber() (string, error) {
	return id3v2.GetStringTXXX("CATALOGNUMBER")
}

func (id3v2 *ID3v23) GetCompilation() (string, error) {
	return id3v2.GetString("TCMP")
}

func (id3v2 *ID3v23) GetComposer() (string, error) {
	return id3v2.GetString("TCOM")
}

func (id3v2 *ID3v23) GetConductor() (string, error) {
	return id3v2.GetString("TPE3")
}

func (id3v2 *ID3v23) GetCopyright() (string, error) {
	return id3v2.GetString("TCOP")
}

func (id3v2 *ID3v23) GetDescription() (string, error) {
	return id3v2.GetString("TIT3")
}

func (id3v2 *ID3v23) GetDiscNumber() (int, int, error) {
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

func (id3v2 *ID3v23) GetEncodedBy() (string, error) {
	return id3v2.GetString("TENC")
}

func (id3v2 *ID3v23) GetTrackNumber() (int, int, error) {
	track, err := id3v2.GetInt("TRCK")
	return track, track, err
}

func (id3v2 *ID3v23) GetPicture() (image.Image, error) {
	pic, err := id3v2.GetAttachedPicture()
	if err != nil {
		return nil, err
	}
	switch pic.MIME {
	case mimeImageJPEG:
		return jpeg.Decode(bytes.NewReader(pic.Data))
	case mimeImagePNG:
		return png.Decode(bytes.NewReader(pic.Data))
	default:
		return nil, ErrIncorrectTag
	}
}

func (id3v2 *ID3v23) SetTitle(title string) error {
	return id3v2.SetString("TIT2", title)
}

func (id3v2 *ID3v23) SetArtist(artist string) error {
	return id3v2.SetString("TPE1", artist)
}

func (id3v2 *ID3v23) SetAlbum(album string) error {
	return id3v2.SetString("TALB", album)
}

func (id3v2 *ID3v23) SetYear(year int) error {
	curDate, err := id3v2.GetTimestamp("TYER")
	if err != nil {
		// set only year
		return id3v2.SetTimestamp(
			"TYER",
			time.Date(year, 0, 0, 0, 0, 0, 0, time.Local),
		)
	}

	return id3v2.SetTimestamp(
		"TYER",
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

func (id3v2 *ID3v23) SetComment(comment string) error {
	return id3v2.SetString("COMM", comment)
}

func (id3v2 *ID3v23) SetGenre(genre string) error {
	return id3v2.SetString("TCON", genre)
}

func (id3v2 *ID3v23) SetAlbumArtist(albumArtist string) error {
	return id3v2.SetString("TPE2", albumArtist)
}

func (id3v2 *ID3v23) SetDate(date time.Time) error {
	return id3v2.SetTimestamp("TYER", date)
}

func (id3v2 *ID3v23) SetArranger(arranger string) error {
	return id3v2.SetString("IPLS", arranger)
}

func (id3v2 *ID3v23) SetAuthor(author string) error {
	return id3v2.SetString("TOLY", author)
}

func (id3v2 *ID3v23) SetBPM(bmp int) error {
	return id3v2.SetInt("TBMP", bmp)
}

func (id3v2 *ID3v23) SetCatalogNumber(catalogNumber string) error {
	return id3v2.SetString("TXXX", catalogNumber)
}

func (id3v2 *ID3v23) SetCompilation(compilation string) error {
	return id3v2.SetString("TCMP", compilation)
}

func (id3v2 *ID3v23) SetComposer(composer string) error {
	return id3v2.SetString("TCOM", composer)
}

func (id3v2 *ID3v23) SetConductor(conductor string) error {
	return id3v2.SetString("TPE3", conductor)
}

func (id3v2 *ID3v23) SetCopyright(copyright string) error {
	return id3v2.SetString("TCOP", copyright)
}

func (id3v2 *ID3v23) SetDescription(description string) error {
	return id3v2.SetString("TIT3", description)
}

func (id3v2 *ID3v23) SetDiscNumber(number int, total int) error {
	return id3v2.SetString("TPOS", fmt.Sprintf("%d/%d", number, total))
}

func (id3v2 *ID3v23) SetEncodedBy(encodedBy string) error {
	return id3v2.SetString("TENC", encodedBy)
}

func (id3v2 *ID3v23) SetTrackNumber(number int, total int) error {
	// only number
	return id3v2.SetInt("TRCK", number)
}

func (id3v2 *ID3v23) SetPicture(picture image.Image) error {
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

func (id3v2 *ID3v23) DeleteAll() error {
	id3v2.Frames = []ID3v23Frame{}
	return nil
}

func (id3v2 *ID3v23) DeleteTitle() error {
	return id3v2.DeleteTag("TIT2")
}

func (id3v2 *ID3v23) DeleteArtist() error {
	return id3v2.DeleteTag("TPE1")
}

func (id3v2 *ID3v23) DeleteAlbum() error {
	return id3v2.DeleteTag("TALB")
}

func (id3v2 *ID3v23) DeleteYear() error {
	return id3v2.DeleteTag("TYER")
}

func (id3v2 *ID3v23) DeleteComment() error {
	return id3v2.DeleteTag("COMM")
}

func (id3v2 *ID3v23) DeleteGenre() error {
	return id3v2.DeleteTag("TCON")
}

func (id3v2 *ID3v23) DeleteAlbumArtist() error {
	return id3v2.DeleteTag("TPE2")
}

func (id3v2 *ID3v23) DeleteDate() error {
	return id3v2.DeleteTag("TYER")
}

func (id3v2 *ID3v23) DeleteArranger() error {
	return id3v2.DeleteTag("IPLS")
}

func (id3v2 *ID3v23) DeleteAuthor() error {
	return id3v2.DeleteTag("TOLY")
}

func (id3v2 *ID3v23) DeleteBPM() error {
	return id3v2.DeleteTag("TBMP")
}

func (id3v2 *ID3v23) DeleteCatalogNumber() error {
	return id3v2.DeleteTagTXXX("CATALOGNUMBER")
}

func (id3v2 *ID3v23) DeleteCompilation() error {
	return id3v2.DeleteTag("TCMP")
}

func (id3v2 *ID3v23) DeleteComposer() error {
	return id3v2.DeleteTag("TCOM")
}

func (id3v2 *ID3v23) DeleteConductor() error {
	return id3v2.DeleteTag("TPE3")
}

func (id3v2 *ID3v23) DeleteCopyright() error {
	return id3v2.DeleteTag("TCOP")
}

func (id3v2 *ID3v23) DeleteDescription() error {
	return id3v2.DeleteTag("TIT3")
}

func (id3v2 *ID3v23) DeleteDiscNumber() error {
	return id3v2.DeleteTag("TPOS")
}

func (id3v2 *ID3v23) DeleteEncodedBy() error {
	return id3v2.DeleteTag("TENC")
}

func (id3v2 *ID3v23) DeleteTrackNumber() error {
	return id3v2.DeleteTag("TRCK")
}

func (id3v2 *ID3v23) DeletePicture() error {
	return id3v2.DeleteTag("APIC")
}

func (id3v2 *ID3v23) SaveFile(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return id3v2.Save(file)
}

func (id3v2 *ID3v23) Save(input io.WriteSeeker) error {
	// write header
	err := id3v2.writeHeaderID3v23(input)
	if err != nil {
		return err
	}

	// write tags
	err = id3v2.writeFramesID3v23(input)
	if err != nil {
		return err
	}

	// write data
	if _, err = id3v2.dataSource.Seek(id3v2.dataStartPos, io.SeekStart); err != nil {
		return err
	}
	if _, err = io.Copy(input, id3v2.dataSource); err != nil {
		return err
	}
	return nil
}

func (id3v2 *ID3v23) writeHeaderID3v23(writer io.Writer) error {
	headerByte := make([]byte, 10)

	// ID3
	copy(headerByte[0:3], id3MarkerValue)

	// Version, Subversion, Flags
	copy(headerByte[3:6], []byte{3, 0, 0})

	// Length
	length := id3v2.getFramesLength()
	lengthByte := IntToByteSynchsafe(length)
	copy(headerByte[6:10], lengthByte)

	nWritten, err := writer.Write(headerByte)
	if err != nil {
		return err
	}
	if nWritten != 10 {
		return ErrWriting
	}
	return nil
}

func (id3v2 *ID3v23) writeFramesID3v23(writer io.Writer) error {
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

func (id3v2 *ID3v23) getFramesLength() int {
	result := 0
	for i := range id3v2.Frames {
		// 10 - size of tag header
		result += 10 + len(id3v2.Frames[i].Value)
	}
	return result
}

func (id3v2 *ID3v23) String() string {
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

func checkID3v23(input io.ReadSeeker) bool {
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
	return versionByte == 3
}

// nolint:funlen
func ReadID3v23(input io.ReadSeekCloser) (*ID3v23, error) {
	header := ID3v23{
		dataSource: input,
	}
	if input == nil {
		return nil, ErrEmptyFile
	}

	// Seek to file start
	startIndex, err := input.Seek(0, io.SeekStart)
	if startIndex != 0 {
		return nil, ErrSeekFile
	}

	if err != nil {
		return nil, err
	}

	// Header size
	headerByte := make([]byte, 10)
	nReaded, err := input.Read(headerByte)
	if err != nil {
		return nil, err
	}
	if nReaded != 10 {
		return nil, errors.New("error header length")
	}

	// Marker
	marker := string(headerByte[0:3])
	if marker != id3MarkerValue {
		return nil, errors.New("error file marker")
	}
	header.Marker = marker

	// Version
	versionByte := headerByte[3]
	if versionByte != 3 {
		return nil, ErrUnsupportedFormat
	}

	// Sub version
	subVersionByte := headerByte[4]
	header.SubVersion = int(subVersionByte)

	// Flags
	header.Flags = id3v23Flags(headerByte[5])

	// Length
	length := ByteToIntSynchsafe(headerByte[6:10])
	header.Length = length

	// Extended headers
	header.Frames = []ID3v23Frame{}
	curRead := 0
	for curRead < length {
		bytesExtendedHeader := make([]byte, 10)
		nReaded, err = input.Read(bytesExtendedHeader)
		if err != nil {
			return nil, err
		}
		if nReaded != 10 {
			return nil, errors.New("error extended header length")
		}
		// Frame identifier
		key := string(bytesExtendedHeader[0:4])

		// Frame data size
		size := ByteToInt(bytesExtendedHeader[4:8])

		bytesExtendedValue := make([]byte, size)
		nReaded, err = input.Read(bytesExtendedValue)
		if err != nil {
			return nil, err
		}
		if nReaded != size {
			return nil, errors.New("error extended value length")
		}

		header.Frames = append(header.Frames, ID3v23Frame{
			key,
			bytesExtendedValue,
		})

		curRead += 10 + size
	}

	// TODO
	if curRead != length {
		return nil, errors.New("error extended frames")
	}

	header.dataStartPos, err = input.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, err
	}

	return &header, nil
}

func (id3v2 *ID3v23) GetString(name string) (string, error) {
	for i := range id3v2.Frames {
		if id3v2.Frames[i].Key == name {
			return GetString(id3v2.Frames[i].Value)
		}
	}
	return "", ErrTagNotFound
}

func (id3v2 *ID3v23) SetString(name string, value string) error {
	frame := ID3v23Frame{
		Key:   name,
		Value: SetString(value),
	}

	// if found set new frame value
	for i := range id3v2.Frames {
		if id3v2.Frames[i].Key == name {
			id3v2.Frames[i].Value = frame.Value
			return nil
		}
	}

	// if not found add new frame to frames
	id3v2.Frames = append(id3v2.Frames, frame)
	return nil
}

func (id3v2 *ID3v23) GetTimestamp(name string) (time.Time, error) {
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

func (id3v2 *ID3v23) SetTimestamp(name string, value time.Time) error {
	str := value.Format("2006-01-02T15:04:05")
	return id3v2.SetString(name, str)
}

func (id3v2 *ID3v23) GetInt(name string) (int, error) {
	intStr, err := id3v2.GetString(name)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(intStr)
}

func (id3v2 *ID3v23) SetInt(name string, value int) error {
	return id3v2.SetString(name, strconv.Itoa(value))
}

func (id3v2 *ID3v23) GetAttachedPicture() (*AttachedPicture, error) {
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
func (id3v2 *ID3v23) SetAttachedPicture(picture *AttachedPicture) error {
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

func (id3v2 *ID3v23) DeleteTag(name string) error {
	index := -1
	for i := range id3v2.Frames {
		if id3v2.Frames[i].Key == name {
			index = i
			break
		}
	}

	// already deleted or not exist
	if index == -1 {
		return nil
	}

	id3v2.Frames = append(id3v2.Frames[:index], id3v2.Frames[index+1:]...)
	return nil
}

func (id3v2 *ID3v23) DeleteTagTXXX(name string) error {
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

	// already deleted or not exist
	if index == -1 {
		return nil
	}

	id3v2.Frames = append(id3v2.Frames[:index], id3v2.Frames[index+1:]...)
	return nil
}

// GetStringTXXX - read TXXX frame
// Header for 'User defined text information frame'
// Text encoding     $xx
// Description       <text string according to encoding> $00 (00)
// Value             <text string according to encoding>.
func (id3v2 *ID3v23) GetStringTXXX(name string) (string, error) {
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

func (id3v2 *ID3v23) SetStringTXXX(name string, value string) error {
	result := ID3v23Frame{
		Key:   "TXXX",
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

func (id3v2 *ID3v23) GetIntTXXX(name string) (int, error) {
	str, err := id3v2.GetStringTXXX(name)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(str)
}

func (id3v2 *ID3v23) Close() error {
	return id3v2.dataSource.Close()
}
