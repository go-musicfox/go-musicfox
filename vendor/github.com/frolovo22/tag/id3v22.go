package tag

import (
	"bytes"
	"errors"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type ID3v22Frame struct {
	Key   string
	Value []byte
}

type ID3v22 struct {
	Marker string // Always
	Length int
	Frames []ID3v22Frame

	dataSource io.ReadSeekCloser
}

func (id3v2 *ID3v22) GetAllTagNames() []string {
	panic("implement me")
}

func (id3v2 *ID3v22) GetVersion() Version {
	return VersionID3v22
}

func (id3v2 *ID3v22) GetFileData() []byte {
	panic("implement me")
}

func (id3v2 *ID3v22) GetTitle() (string, error) {
	return id3v2.GetString("TT2")
}

func (id3v2 *ID3v22) GetArtist() (string, error) {
	return id3v2.GetString("TP1")
}

func (id3v2 *ID3v22) GetAlbum() (string, error) {
	return id3v2.GetString("TAL")
}

func (id3v2 *ID3v22) GetYear() (int, error) {
	year, err := id3v2.GetString("TYE")
	if err != nil {
		return 0, nil
	}
	return strconv.Atoi(year)
}

func (id3v2 *ID3v22) GetComment() (string, error) {
	panic("implement me")
}

func (id3v2 *ID3v22) GetGenre() (string, error) {
	genre, err := id3v2.GetString("TCO")
	if err != nil {
		return "", err
	}
	re := regexp.MustCompile(`\([0-9]+\)`)
	match := re.FindAllString(genre, -1)
	if len(match) > 0 {
		code, err := strconv.Atoi(match[0][1 : len(match[0])-1])
		if err != nil {
			return "", err
		}
		return genres[Genre(code)], nil
	}
	return "", nil
}

func (id3v2 *ID3v22) GetAlbumArtist() (string, error) {
	panic("implement me")
}

func (id3v2 *ID3v22) GetDate() (time.Time, error) {
	panic("implement me")
}

func (id3v2 *ID3v22) GetArranger() (string, error) {
	panic("implement me")
}

func (id3v2 *ID3v22) GetAuthor() (string, error) {
	panic("implement me")
}

func (id3v2 *ID3v22) GetBPM() (int, error) {
	panic("implement me")
}

func (id3v2 *ID3v22) GetCatalogNumber() (string, error) {
	panic("implement me")
}

func (id3v2 *ID3v22) GetCompilation() (string, error) {
	panic("implement me")
}

func (id3v2 *ID3v22) GetComposer() (string, error) {
	panic("implement me")
}

func (id3v2 *ID3v22) GetConductor() (string, error) {
	panic("implement me")
}

func (id3v2 *ID3v22) GetCopyright() (string, error) {
	panic("implement me")
}

func (id3v2 *ID3v22) GetDescription() (string, error) {
	panic("implement me")
}

func (id3v2 *ID3v22) GetDiscNumber() (int, int, error) {
	panic("implement me")
}

func (id3v2 *ID3v22) GetEncodedBy() (string, error) {
	return id3v2.GetString("TEN")
}

func (id3v2 *ID3v22) GetTrackNumber() (int, int, error) {
	track, err := id3v2.GetString("TRK")
	if err != nil {
		return 0, 0, err
	}
	parts := strings.Split(track, "/")
	if len(parts) == 1 {
		number, err := strconv.Atoi(parts[0])
		if err != nil {
			return 0, 0, err
		}
		return number, number, nil
	} else if len(parts) == 2 {
		number1, err := strconv.Atoi(parts[0])
		if err != nil {
			return 0, 0, err
		}
		number2, err := strconv.Atoi(parts[1])
		if err != nil {
			return 0, 0, err
		}
		return number1, number2, nil
	}

	return 0, 0, ErrIncorrectTag
}

func (id3v2 *ID3v22) GetPicture() (image.Image, error) {
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

func (id3v2 *ID3v22) GetAttachedPicture() (*AttachedPicture, error) {
	var picture AttachedPicture

	bytes, err := id3v2.GetBytes("PIC")
	if err != nil {
		return nil, err
	}

	textEncoding := bytes[0]
	mimeText := string(bytes[1:4])
	if mimeText == "JPG" {
		picture.MIME = mimeImageJPEG
	} else if mimeText == "PNG" {
		picture.MIME = mimeImagePNG
	}

	picture.PictureType = bytes[4]

	values := SplitBytesWithTextDescription(bytes[5:], GetEncoding(textEncoding))
	if len(values) != 2 {
		return nil, ErrIncorrectTag
	}

	desc, err := DecodeString(values[0], GetEncoding(textEncoding))
	if err != nil {
		return nil, err
	}

	picture.Description = desc
	picture.Data = values[1]

	return &picture, nil
}

func (id3v2 *ID3v22) SetTitle(title string) error {
	panic("implement me")
}

func (id3v2 *ID3v22) SetArtist(artist string) error {
	panic("implement me")
}

func (id3v2 *ID3v22) SetAlbum(album string) error {
	panic("implement me")
}

func (id3v2 *ID3v22) SetYear(year int) error {
	panic("implement me")
}

func (id3v2 *ID3v22) SetComment(comment string) error {
	panic("implement me")
}

func (id3v2 *ID3v22) SetGenre(genre string) error {
	panic("implement me")
}

func (id3v2 *ID3v22) SetAlbumArtist(albumArtist string) error {
	panic("implement me")
}

func (id3v2 *ID3v22) SetDate(date time.Time) error {
	panic("implement me")
}

func (id3v2 *ID3v22) SetArranger(arranger string) error {
	panic("implement me")
}

func (id3v2 *ID3v22) SetAuthor(author string) error {
	panic("implement me")
}

func (id3v2 *ID3v22) SetBPM(bmp int) error {
	panic("implement me")
}

func (id3v2 *ID3v22) SetCatalogNumber(catalogNumber string) error {
	panic("implement me")
}

func (id3v2 *ID3v22) SetCompilation(compilation string) error {
	panic("implement me")
}

func (id3v2 *ID3v22) SetComposer(composer string) error {
	panic("implement me")
}

func (id3v2 *ID3v22) SetConductor(conductor string) error {
	panic("implement me")
}

func (id3v2 *ID3v22) SetCopyright(copyright string) error {
	panic("implement me")
}

func (id3v2 *ID3v22) SetDescription(description string) error {
	panic("implement me")
}

func (id3v2 *ID3v22) SetDiscNumber(number int, total int) error {
	panic("implement me")
}

func (id3v2 *ID3v22) SetEncodedBy(encodedBy string) error {
	panic("implement me")
}

func (id3v2 *ID3v22) SetTrackNumber(number int, total int) error {
	panic("implement me")
}

func (id3v2 *ID3v22) SetPicture(picture image.Image) error {
	panic("implement me")
}

func (id3v2 *ID3v22) DeleteAll() error {
	panic("implement me")
}

func (id3v2 *ID3v22) DeleteTitle() error {
	panic("implement me")
}

func (id3v2 *ID3v22) DeleteArtist() error {
	panic("implement me")
}

func (id3v2 *ID3v22) DeleteAlbum() error {
	panic("implement me")
}

func (id3v2 *ID3v22) DeleteYear() error {
	panic("implement me")
}

func (id3v2 *ID3v22) DeleteComment() error {
	panic("implement me")
}

func (id3v2 *ID3v22) DeleteGenre() error {
	panic("implement me")
}

func (id3v2 *ID3v22) DeleteAlbumArtist() error {
	panic("implement me")
}

func (id3v2 *ID3v22) DeleteDate() error {
	panic("implement me")
}

func (id3v2 *ID3v22) DeleteArranger() error {
	panic("implement me")
}

func (id3v2 *ID3v22) DeleteAuthor() error {
	panic("implement me")
}

func (id3v2 *ID3v22) DeleteBPM() error {
	panic("implement me")
}

func (id3v2 *ID3v22) DeleteCatalogNumber() error {
	panic("implement me")
}

func (id3v2 *ID3v22) DeleteCompilation() error {
	panic("implement me")
}

func (id3v2 *ID3v22) DeleteComposer() error {
	panic("implement me")
}

func (id3v2 *ID3v22) DeleteConductor() error {
	panic("implement me")
}

func (id3v2 *ID3v22) DeleteCopyright() error {
	panic("implement me")
}

func (id3v2 *ID3v22) DeleteDescription() error {
	panic("implement me")
}

func (id3v2 *ID3v22) DeleteDiscNumber() error {
	panic("implement me")
}

func (id3v2 *ID3v22) DeleteEncodedBy() error {
	panic("implement me")
}

func (id3v2 *ID3v22) DeleteTrackNumber() error {
	panic("implement me")
}

func (id3v2 *ID3v22) DeletePicture() error {
	panic("implement me")
}

func (id3v2 *ID3v22) SaveFile(path string) error {
	panic("implement me")
}

func (id3v2 *ID3v22) Save(input io.WriteSeeker) error {
	panic("implement me")
}

// nolint:gocyclo
func ReadID3v22(input io.ReadSeekCloser) (*ID3v22, error) {
	header := ID3v22{
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
	if versionByte != 2 {
		return nil, ErrUnsupportedFormat
	}

	// Length
	length := ByteToIntSynchsafe(headerByte[6:10])
	header.Length = length

	curRead := 0
	for curRead < length {
		bytesExtendedHeader := make([]byte, id3v22FrameHeaderSize)
		nReaded, err = input.Read(bytesExtendedHeader)
		if err != nil {
			return nil, err
		}
		if nReaded != 6 {
			return nil, errors.New("error extended header length")
		}
		// Frame identifier
		key := string(bytesExtendedHeader[0:3])

		// Frame data size
		size := ByteToInt(bytesExtendedHeader[3:id3v22FrameHeaderSize])

		bytesExtendedValue := make([]byte, size)
		nReaded, err = input.Read(bytesExtendedValue)
		if err != nil {
			return nil, err
		}
		if nReaded != size {
			return nil, errors.New("error extended value length")
		}

		if key[0:1] == "T" {
			pos := -1
			for i, v := range bytesExtendedValue {
				if v == 0 && i > 0 {
					pos = i
				}
			}
			if pos != -1 {
				bytesExtendedValue = bytesExtendedValue[0:pos]
			}
		}

		header.Frames = append(header.Frames, ID3v22Frame{
			key,
			bytesExtendedValue,
		})

		curRead += id3v22FrameHeaderSize + size
	}
	return &header, nil
}

func checkID3v22(input io.ReadSeeker) bool {
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
	return versionByte == 2
}

func (id3v2 *ID3v22) GetString(name string) (string, error) {
	for i := range id3v2.Frames {
		if id3v2.Frames[i].Key == name {
			return GetString(id3v2.Frames[i].Value)
		}
	}
	return "", ErrTagNotFound
}

func (id3v2 *ID3v22) GetBytes(name string) ([]byte, error) {
	for i := range id3v2.Frames {
		if id3v2.Frames[i].Key == name {
			return id3v2.Frames[i].Value, nil
		}
	}
	return nil, ErrTagNotFound
}

func (id3v2 *ID3v22) Close() error {
	return id3v2.dataSource.Close()
}
