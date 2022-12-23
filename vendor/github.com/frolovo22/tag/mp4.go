package tag

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"strconv"
	"time"
)

const Mp4Marker = "ftyp"
const Mp4MoovAtom = "moov"
const Mp4MetaAtom = "meta"
const Mp4MetaUpta = "udta"
const Mp4MetaIlst = "ilst"

const Mp4TagAlbum = "album"
const Mp4TagArtist = "artist"
const Mp4TagAlbumArtist = "album_artist"
const Mp4TagYear = "year"
const Mp4TagTitle = "title"
const Mp4TagGenre = "genre"
const Mp4TagTrack = "track"
const Mp4TagComposer = "composer"
const Mp4TagEncoder = "encoder"
const Mp4TagCopyright = "copyright"
const Mp4TagPicture = "picture"
const Mp4TagGrouping = "grouping"
const Mp4TagKeyword = "keyword"
const Mp4TagLyrics = "lyrics"
const Mp4TagComment = "comment"
const Mp4TagTempo = "tempo"
const Mp4TagCompilation = "compilation"
const Mp4TagDisc = "disk"

var Mp4Types = [...]string{
	"mp41",
	"mp42",
	"isom",
	"iso2",
	"M4A ",
	"M4B ",
}

var atoms = map[string]string{
	"\xa9alb": Mp4TagAlbum,
	"\xa9art": Mp4TagArtist,
	"\xa9ART": Mp4TagArtist,
	"aART":    Mp4TagAlbumArtist,
	"\xa9day": Mp4TagYear,
	"\xa9nam": Mp4TagTitle,
	"\xa9gen": Mp4TagGenre,
	"trkn":    Mp4TagTrack,
	"\xa9wrt": Mp4TagComposer,
	"\xa9too": Mp4TagEncoder,
	"cprt":    Mp4TagCopyright,
	"covr":    Mp4TagPicture,
	"\xa9grp": Mp4TagGrouping,
	"keyw":    Mp4TagKeyword,
	"\xa9lyr": Mp4TagLyrics,
	"\xa9cmt": Mp4TagComment,
	"tmpo":    Mp4TagTempo,
	"cpil":    Mp4TagCompilation,
	"disk":    Mp4TagDisc,
}

type MP4 struct {
	data map[string]interface{}

	dataSource io.ReadSeekCloser
}

func (MP4) GetAllTagNames() []string {
	panic("implement me")
}

func (mp4 *MP4) GetVersion() Version {
	return VersionMP4
}

func (MP4) GetFileData() []byte {
	panic("implement me")
}

func (mp4 *MP4) GetTitle() (string, error) {
	return mp4.getString(Mp4TagTitle)
}

func (mp4 *MP4) GetArtist() (string, error) {
	return mp4.getString(Mp4TagArtist)
}

func (mp4 *MP4) GetAlbum() (string, error) {
	return mp4.getString(Mp4TagAlbum)
}

func (mp4 *MP4) GetYear() (int, error) {
	year, err := mp4.getString(Mp4TagYear)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(year)
}

func (MP4) GetComment() (string, error) {
	panic("implement me")
}

func (mp4 *MP4) GetGenre() (string, error) {
	return mp4.getString(Mp4TagGenre)
}

func (mp4 *MP4) GetAlbumArtist() (string, error) {
	return mp4.getString(Mp4TagAlbumArtist)
}

func (MP4) GetDate() (time.Time, error) {
	panic("implement me")
}

func (MP4) GetArranger() (string, error) {
	panic("implement me")
}

func (MP4) GetAuthor() (string, error) {
	panic("implement me")
}

func (MP4) GetBPM() (int, error) {
	panic("implement me")
}

func (MP4) GetCatalogNumber() (string, error) {
	panic("implement me")
}

func (MP4) GetCompilation() (string, error) {
	panic("implement me")
}

func (mp4 *MP4) GetComposer() (string, error) {
	return mp4.getString(Mp4TagComposer)
}

func (MP4) GetConductor() (string, error) {
	panic("implement me")
}

func (mp4 *MP4) GetCopyright() (string, error) {
	return mp4.getString(Mp4TagCopyright)
}

func (MP4) GetDescription() (string, error) {
	panic("implement me")
}

func (MP4) GetDiscNumber() (int, int, error) {
	panic("implement me")
}

func (mp4 *MP4) GetEncodedBy() (string, error) {
	return mp4.getString(Mp4TagEncoder)
}

func (mp4 *MP4) GetTrackNumber() (int, int, error) {
	track, err := mp4.getInt(Mp4TagTrack)
	if err != nil {
		return 0, 0, err
	}
	total, err2 := mp4.getInt(Mp4TagTrack + "_TOTAL")
	if err2 != nil {
		return 0, 0, err2
	}
	return track, total, nil
}

func (mp4 *MP4) GetPicture() (image.Image, error) {
	pictureBlock, ok := mp4.data[Mp4TagPicture]
	if !ok {
		return nil, ErrTagNotFound
	}
	picture, ok := pictureBlock.(AttachedPicture)
	if !ok {
		return nil, ErrNotPictureBlock
	}

	switch picture.MIME {
	case "image/jpeg":
		return jpeg.Decode(bytes.NewReader(picture.Data))
	case "image/png":
		return png.Decode(bytes.NewReader(picture.Data))
	}

	return nil, ErrIncorrectTag
}

func (MP4) SetTitle(title string) error {
	panic("implement me")
}

func (MP4) SetArtist(artist string) error {
	panic("implement me")
}

func (MP4) SetAlbum(album string) error {
	panic("implement me")
}

func (MP4) SetYear(year int) error {
	panic("implement me")
}

func (MP4) SetComment(comment string) error {
	panic("implement me")
}

func (MP4) SetGenre(genre string) error {
	panic("implement me")
}

func (MP4) SetAlbumArtist(albumArtist string) error {
	panic("implement me")
}

func (MP4) SetDate(date time.Time) error {
	panic("implement me")
}

func (MP4) SetArranger(arranger string) error {
	panic("implement me")
}

func (MP4) SetAuthor(author string) error {
	panic("implement me")
}

func (MP4) SetBPM(bmp int) error {
	panic("implement me")
}

func (MP4) SetCatalogNumber(catalogNumber string) error {
	panic("implement me")
}

func (MP4) SetCompilation(compilation string) error {
	panic("implement me")
}

func (MP4) SetComposer(composer string) error {
	panic("implement me")
}

func (MP4) SetConductor(conductor string) error {
	panic("implement me")
}

func (MP4) SetCopyright(copyright string) error {
	panic("implement me")
}

func (MP4) SetDescription(description string) error {
	panic("implement me")
}

func (MP4) SetDiscNumber(number int, total int) error {
	panic("implement me")
}

func (MP4) SetEncodedBy(encodedBy string) error {
	panic("implement me")
}

func (MP4) SetTrackNumber(number int, total int) error {
	panic("implement me")
}

func (MP4) SetPicture(picture image.Image) error {
	panic("implement me")
}

func (MP4) DeleteAll() error {
	panic("implement me")
}

func (MP4) DeleteTitle() error {
	panic("implement me")
}

func (MP4) DeleteArtist() error {
	panic("implement me")
}

func (MP4) DeleteAlbum() error {
	panic("implement me")
}

func (MP4) DeleteYear() error {
	panic("implement me")
}

func (MP4) DeleteComment() error {
	panic("implement me")
}

func (MP4) DeleteGenre() error {
	panic("implement me")
}

func (MP4) DeleteAlbumArtist() error {
	panic("implement me")
}

func (MP4) DeleteDate() error {
	panic("implement me")
}

func (MP4) DeleteArranger() error {
	panic("implement me")
}

func (MP4) DeleteAuthor() error {
	panic("implement me")
}

func (MP4) DeleteBPM() error {
	panic("implement me")
}

func (MP4) DeleteCatalogNumber() error {
	panic("implement me")
}

func (MP4) DeleteCompilation() error {
	panic("implement me")
}

func (MP4) DeleteComposer() error {
	panic("implement me")
}

func (MP4) DeleteConductor() error {
	panic("implement me")
}

func (MP4) DeleteCopyright() error {
	panic("implement me")
}

func (MP4) DeleteDescription() error {
	panic("implement me")
}

func (MP4) DeleteDiscNumber() error {
	panic("implement me")
}

func (MP4) DeleteEncodedBy() error {
	panic("implement me")
}

func (MP4) DeleteTrackNumber() error {
	panic("implement me")
}

func (MP4) DeletePicture() error {
	panic("implement me")
}

func (MP4) SaveFile(path string) error {
	panic("implement me")
}

func (MP4) Save(input io.WriteSeeker) error {
	panic("implement me")
}

func (mp4 *MP4) getString(tag string) (string, error) {
	val, ok := mp4.data[tag]
	if !ok {
		return "", ErrTagNotFound
	}
	return val.(string), nil
}

func (mp4 *MP4) getInt(tag string) (int, error) {
	val, ok := mp4.data[tag]
	if !ok {
		return 0, ErrTagNotFound
	}
	return val.(int), nil
}

func checkMp4(input io.ReadSeeker) bool {
	if input == nil {
		return false
	}

	data, err := seekAndRead(input, 0, io.SeekStart, 12)
	if err != nil {
		return false
	}
	marker := string(data[4:8])

	if marker == Mp4Marker {
		mp4type := string(data[8:12])
		for _, t := range Mp4Types {
			if mp4type == t {
				return true
			}
		}
	}

	return false
}

func ReadMp4(input io.ReadSeekCloser) (*MP4, error) {
	header := MP4{dataSource: input}
	header.data = map[string]interface{}{}

	// Seek to file start
	startIndex, err := input.Seek(0, io.SeekStart)
	if startIndex != 0 {
		return nil, ErrSeekFile
	}

	if err != nil {
		return nil, err
	}

	for {
		var size uint32 = 0
		err = binary.Read(input, binary.BigEndian, &size)
		if err != nil {
			break
		}

		nameBytes := make([]byte, 4)
		_, err = input.Read(nameBytes)
		if err != nil {
			break
		}
		name := string(nameBytes)

		bytes := make([]byte, size-8)
		_, err = input.Read(bytes)
		if err != nil {
			break
		}

		if name == Mp4MoovAtom {
			parseMoovAtom(bytes, &header)
		}
	}

	return &header, nil
}

func parseMoovAtom(bytes []byte, mp4 *MP4) {
	for {
		size := binary.BigEndian.Uint32(bytes[0:4])
		name := string(bytes[4:8])

		switch name {
		case Mp4MetaAtom:
			bytes = bytes[4:]
			size -= 4
			parseMoovAtom(bytes[8:], mp4)
		case Mp4MetaUpta, Mp4MetaIlst:
			parseMoovAtom(bytes[8:], mp4)
		default:
			atomName, ok := atoms[name]
			if ok {
				parseAtomData(bytes[8:size], atomName, mp4)
			}
		}

		bytes = bytes[size:]

		if len(bytes) == 0 {
			break
		}
	}
}

func parseAtomData(bytes []byte, atomName string, mp4 *MP4) {
	// TODO : different types
	value := string(bytes[16:])
	mp4.data[atomName] = value

	datatype := binary.BigEndian.Uint32(bytes[8:12])
	if datatype == 13 {
		mp4.data[atomName] = AttachedPicture{
			MIME: "image/jpeg",
			Data: bytes[16:],
		}
	}

	if atomName == Mp4TagTrack || atomName == Mp4TagDisc {
		mp4.data[atomName] = int(bytes[19:20][0])
		mp4.data[atomName+"_TOTAL"] = int(bytes[21:22][0])
	}
}

func (mp4 *MP4) Close() error {
	return mp4.dataSource.Close()
}
