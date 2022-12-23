package tag

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"
)

type FLAC struct {
	Blocks []*FlacMetadataBlock

	// Vorbis Comment
	Vendor string
	Tags   map[string]string

	dataStartPos int64
	dataSource   io.ReadSeekCloser
}

func (flac *FLAC) GetAllTagNames() []string {
	result := make([]string, 0, len(flac.Tags))
	for key := range flac.Tags {
		result = append(result, key)
	}
	return result
}

func (flac *FLAC) GetVersion() Version {
	return VersionFLAC
}

func (flac *FLAC) GetFileData() []byte {
	_, _ = flac.dataSource.Seek(flac.dataStartPos, io.SeekStart)
	data, _ := io.ReadAll(flac.dataSource)
	return data
}

func (flac *FLAC) GetTitle() (string, error) {
	return flac.GetVorbisComment("TITLE")
}

func (flac *FLAC) GetArtist() (string, error) {
	return flac.GetVorbisComment("ARTIST")
}

func (flac *FLAC) GetAlbum() (string, error) {
	return flac.GetVorbisComment("ALBUM")
}

func (flac *FLAC) GetYear() (int, error) {
	return flac.GetVorbisCommentInt("YEAR")
}

func (flac *FLAC) GetComment() (string, error) {
	return flac.GetVorbisComment("COMMENT")
}

func (flac *FLAC) GetGenre() (string, error) {
	return flac.GetVorbisComment("GENRE")
}

func (flac *FLAC) GetAlbumArtist() (string, error) {
	return flac.GetVorbisComment("ALBUMARTIST")
}

func (flac *FLAC) GetDate() (time.Time, error) {
	return flac.GetVorbisCommentTime("DATE")
}

func (flac *FLAC) GetArranger() (string, error) {
	return flac.GetVorbisComment("ARRANGER")
}

func (flac *FLAC) GetAuthor() (string, error) {
	return flac.GetVorbisComment("AUTHOR")
}

func (flac *FLAC) GetBPM() (int, error) {
	return flac.GetVorbisCommentInt("BPM")
}

func (flac *FLAC) GetCatalogNumber() (string, error) {
	return flac.GetVorbisComment("CATALOGNUMBER")
}

func (flac *FLAC) GetCompilation() (string, error) {
	return flac.GetVorbisComment("COMPILATION")
}

func (flac *FLAC) GetComposer() (string, error) {
	return flac.GetVorbisComment("COMPOSER")
}

func (flac *FLAC) GetConductor() (string, error) {
	return flac.GetVorbisComment("CONDUCTOR")
}

func (flac *FLAC) GetCopyright() (string, error) {
	return flac.GetVorbisComment("COPYRIGHT")
}

func (flac *FLAC) GetDescription() (string, error) {
	return flac.GetVorbisComment("DESCRIPTION")
}

func (flac *FLAC) GetDiscNumber() (int, int, error) {
	number, err := flac.GetVorbisCommentInt("DISCNUMBER")
	if err != nil {
		return 0, 0, err
	}
	total, err := flac.GetVorbisCommentInt("DISCTOTAL")
	if err != nil {
		return 0, 0, err
	}
	return number, total, nil
}

func (flac *FLAC) GetEncodedBy() (string, error) {
	return flac.GetVorbisComment("ENCODED-BY")
}

func (flac *FLAC) GetTrackNumber() (int, int, error) {
	number, err := flac.GetVorbisCommentInt("TRACKNUMBER")
	if err != nil {
		return 0, 0, err
	}
	total, err := flac.GetVorbisCommentInt("TRACKTOTAL")
	if err != nil {
		return 0, 0, err
	}
	return number, total, nil
}

func (flac *FLAC) GetPicture() (image.Image, error) {
	pictureBlock, err := flac.GetMetadataBlockPicture()
	if err != nil {
		return nil, err
	}
	switch pictureBlock.MIME {
	case mimeImageJPEG:
		return jpeg.Decode(bytes.NewReader(pictureBlock.PictureData))
	case mimeImagePNG:
		return png.Decode(bytes.NewReader(pictureBlock.PictureData))
	case mimeImageLink:
		return downloadImage(string(pictureBlock.PictureData))
	}

	return nil, ErrIncorrectTag
}

func (flac *FLAC) SetTitle(title string) error {
	flac.Tags["TITLE"] = title
	return nil
}

func (flac *FLAC) SetArtist(artist string) error {
	flac.Tags["ARTIST"] = artist
	return nil
}

func (flac *FLAC) SetAlbum(album string) error {
	flac.Tags["ALBUM"] = album
	return nil
}

func (flac *FLAC) SetYear(year int) error {
	flac.Tags["YEAR"] = strconv.Itoa(year)
	return nil
}

func (flac *FLAC) SetComment(comment string) error {
	flac.Tags["COMMENT"] = comment
	return nil
}

func (flac *FLAC) SetGenre(genre string) error {
	flac.Tags["GENRE"] = genre
	return nil
}

func (flac *FLAC) SetAlbumArtist(albumArtist string) error {
	flac.Tags["ALBUMARTIST"] = albumArtist
	return nil
}

func (flac *FLAC) SetDate(date time.Time) error {
	flac.Tags["DATE"] = date.Format("2006-01-02T15:04:05")
	return nil
}

func (flac *FLAC) SetArranger(arranger string) error {
	flac.Tags["ARRANGER"] = arranger
	return nil
}

func (flac *FLAC) SetAuthor(author string) error {
	flac.Tags["AUTHOR"] = author
	return nil
}

func (flac *FLAC) SetBPM(bmp int) error {
	flac.Tags["BMP"] = strconv.Itoa(bmp)
	return nil
}

func (flac *FLAC) SetCatalogNumber(catalogNumber string) error {
	flac.Tags["CATALOGNUMBER"] = catalogNumber
	return nil
}

func (flac *FLAC) SetCompilation(compilation string) error {
	flac.Tags["COMPILATION"] = compilation
	return nil
}

func (flac *FLAC) SetComposer(composer string) error {
	flac.Tags["COMPOSER"] = composer
	return nil
}

func (flac *FLAC) SetConductor(conductor string) error {
	flac.Tags["CONDUCTOR"] = conductor
	return nil
}

func (flac *FLAC) SetCopyright(copyright string) error {
	flac.Tags["COPYRIGHT"] = copyright
	return nil
}

func (flac *FLAC) SetDescription(description string) error {
	flac.Tags["DESCRIPTION"] = description
	return nil
}

func (flac *FLAC) SetDiscNumber(number int, total int) error {
	flac.Tags["DISCNUMBER"] = strconv.Itoa(number)
	flac.Tags["DISCTOTAL"] = strconv.Itoa(total)
	return nil
}

func (flac *FLAC) SetEncodedBy(encodedBy string) error {
	flac.Tags["ENCODED-BY"] = encodedBy
	return nil
}

func (flac *FLAC) SetTrackNumber(number int, total int) error {
	flac.Tags["TRACKNUMBER"] = strconv.Itoa(number)
	flac.Tags["TRACKTOTAL"] = strconv.Itoa(total)
	return nil
}

func (flac *FLAC) SetPicture(picture image.Image) error {
	// TODO
	// bitsPerPixel := colorModelToBitsPerPixel(picture.ColorModel())
	// pictureBlock := FlacMetadataBlockPicture{
	//	Type:           3, // Other type
	//	MIME:           "image/png",
	//	Description:    "",
	//	Width:          int32(picture.Bounds().Size().X),
	//	Height:         int32(picture.Bounds().Size().Y),
	//	BitsPerPixel:   int32(bitsPerPixel),
	//	NumberOfColors: 0,
	//	PictureData:    nil,
	//}
	var size int
	var data []byte

	for _, val := range flac.Blocks {
		if val.Type == FlacPicture {
			val.Size = size
			val.Data = data
		}
	}
	return nil
}

func (flac *FLAC) DeleteAll() error {
	flac.Tags = map[string]string{}
	return nil
}

func (flac *FLAC) DeleteTitle() error {
	delete(flac.Tags, "TITLE")
	return nil
}

func (flac *FLAC) DeleteArtist() error {
	delete(flac.Tags, "ARTIST")
	return nil
}

func (flac *FLAC) DeleteAlbum() error {
	delete(flac.Tags, "ALBUM")
	return nil
}

func (flac *FLAC) DeleteYear() error {
	delete(flac.Tags, "YEAR")
	return nil
}

func (flac *FLAC) DeleteComment() error {
	delete(flac.Tags, "COMMENT")
	return nil
}

func (flac *FLAC) DeleteGenre() error {
	delete(flac.Tags, "GENRE")
	return nil
}

func (flac *FLAC) DeleteAlbumArtist() error {
	delete(flac.Tags, "ALBUMARTIST")
	return nil
}

func (flac *FLAC) DeleteDate() error {
	delete(flac.Tags, "DATE")
	return nil
}

func (flac *FLAC) DeleteArranger() error {
	delete(flac.Tags, "ARRANGER")
	return nil
}

func (flac *FLAC) DeleteAuthor() error {
	delete(flac.Tags, "AUTHOR")
	return nil
}

func (flac *FLAC) DeleteBPM() error {
	delete(flac.Tags, "BPM")
	return nil
}

func (flac *FLAC) DeleteCatalogNumber() error {
	delete(flac.Tags, "CATALOGNUMBER")
	return nil
}

func (flac *FLAC) DeleteCompilation() error {
	delete(flac.Tags, "COMPILATION")
	return nil
}

func (flac *FLAC) DeleteComposer() error {
	delete(flac.Tags, "COMPOSER")
	return nil
}

func (flac *FLAC) DeleteConductor() error {
	delete(flac.Tags, "CONDUCTOR")
	return nil
}

func (flac *FLAC) DeleteCopyright() error {
	delete(flac.Tags, "COPYRIGHT")
	return nil
}

func (flac *FLAC) DeleteDescription() error {
	delete(flac.Tags, "DESCRIPTION")
	return nil
}

func (flac *FLAC) DeleteDiscNumber() error {
	delete(flac.Tags, "DISCNUMBER")
	delete(flac.Tags, "DISCTOTAL")
	return nil
}

func (flac *FLAC) DeleteEncodedBy() error {
	delete(flac.Tags, "ENCODED-BY")
	return nil
}

func (flac *FLAC) DeleteTrackNumber() error {
	delete(flac.Tags, "TRACKNUMBER")
	delete(flac.Tags, "TRACKTOTAL")
	return nil
}

func (flac *FLAC) DeletePicture() error {
	index := -1
	for i, val := range flac.Blocks {
		if val.Type == FlacPicture {
			index = i
		}
	}
	if index != -1 {
		flac.Blocks = append(flac.Blocks[:index], flac.Blocks[index+1:]...)
	}
	return nil
}

func (flac *FLAC) SaveFile(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return flac.Save(file)
}

func (flac *FLAC) Save(input io.WriteSeeker) error {
	// Sync metadata to its parent block
	// Write out file header
	// Then all metablocks
	// Then all audio blocks
	// Update comments

	if _, err := input.Write([]byte(FLACIdentifier)); err != nil {
		return err
	}
	metadataWritten := false
	for i, meta := range flac.Blocks {
		if meta.Type == FlacVorbisComment {
			flac.Blocks[i].Data = serializeVorbisComments(flac.Tags, flac.Vendor)
			flac.Blocks[i].Size = len(flac.Blocks[i].Data)
			metadataWritten = true
			break
		}
		//
	}
	if !metadataWritten {
		// Append a metadata block to the list
		data := serializeVorbisComments(flac.Tags, flac.Vendor)
		block := &FlacMetadataBlock{
			IsLast: false,
			Type:   FlacVorbisComment,
			Size:   len(data),
			Data:   data,
		}
		flac.Blocks = append(flac.Blocks, block)
	}

	for i, meta := range flac.Blocks {
		last := i == len(flac.Blocks)-1
		if err := meta.Write(input, last); err != nil {
			return err
		}
	}

	if _, err := flac.dataSource.Seek(flac.dataStartPos, io.SeekStart); err != nil {
		return err
	}
	if _, err := io.Copy(input, flac.dataSource); err != nil {
		return err
	}
	return nil
}

func (flac *FLAC) Close() error {
	return flac.dataSource.Close()
}
func checkFLAC(input io.ReadSeeker) bool {
	data, err := seekAndRead(input, 0, io.SeekStart, 4)
	if err != nil {
		return false
	}

	if string(data) != FLACIdentifier {
		return false
	}

	return true
}

type FlacMetadataBlockType byte

/*
BLOCK_TYPE:

	0 : STREAMINFO
	1 : PADDING
	2 : APPLICATION
	3 : SEEKTABLE
	4 : VORBIS_COMMENT
	5 : CUESHEET
	6 : PICTURE
	7-126 : reserved
	127 : invalid, to avoid confusion with a frame sync code
*/
const (
	FlacStreamInfo    FlacMetadataBlockType = 0
	FlacPadding       FlacMetadataBlockType = 1
	FlacApplication   FlacMetadataBlockType = 2
	FlacSeekTable     FlacMetadataBlockType = 3
	FlacVorbisComment FlacMetadataBlockType = 4
	FlacCueSheet      FlacMetadataBlockType = 5
	FlacPicture       FlacMetadataBlockType = 6
)

type FlacMetadataBlock struct {
	// IsLast - Last-metadata-block flag:
	// '1' if this block is the last metadata block before the audio blocks,
	// '0' otherwise.
	IsLast bool
	Type   FlacMetadataBlockType
	Size   int
	Data   []byte
}

func ReadFLAC(input io.ReadSeekCloser) (*FLAC, error) {
	flac := FLAC{
		Tags:       map[string]string{},
		dataSource: input,
	}

	// FLAC identifier
	data, err := seekAndRead(input, 0, io.SeekStart, 4)
	if err != nil {
		return nil, err
	}

	if string(data) != "fLaC" {
		return nil, ErrFileMarker
	}

	// read blocks
	for {
		var block *FlacMetadataBlock
		block, err = readMetadataBlock(input)
		if err != nil {
			return nil, err
		}

		if block.Type == FlacVorbisComment {
			var comments []VorbisComment
			var vendor string

			comments, vendor, err = readVorbisComments(bytes.NewReader(block.Data))
			if err != nil {
				return nil, err
			}

			flac.Vendor = vendor
			for i := range comments {
				// case insensitive
				field := strings.ToUpper(comments[i].Name)
				flac.Tags[field] = comments[i].Value
			}
		} else {
			flac.Blocks = append(flac.Blocks, block)
		}

		// last block before audio frame
		if block.IsLast {
			break
		}
	}

	flac.dataStartPos, err = input.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, err
	}

	return &flac, nil
}

func readMetadataBlock(input io.Reader) (*FlacMetadataBlock, error) {
	header := FlacMetadataBlock{}

	// 4 - header size
	headerBytes, err := readBytes(input, 4)
	if err != nil {
		return nil, err
	}

	// first bit
	isLastBit := GetBit(headerBytes[0], 7)
	if isLastBit == 1 {
		header.IsLast = true
	}

	// Only 0-6 types
	// 1-7 bits
	blockType := headerBytes[0] & 0x7F
	if blockType > 6 {
		return nil, ErrReadFile
	}
	header.Type = FlacMetadataBlockType(blockType)

	// 3 bytes size
	header.Size = ByteToInt(headerBytes[1:])

	// block data
	header.Data, err = readBytes(input, header.Size)
	if err != nil {
		return nil, err
	}
	return &header, nil
}

func (block *FlacMetadataBlock) Write(w io.Writer, isLast bool) error {
	block.Size = len(block.Data)
	block.IsLast = isLast
	blockHeader := []byte{byte(block.Type) & 0x7F, 0, 0, 0}

	if isLast {
		blockHeader[0] |= 1 << 7
	}
	// 3 bytes encodes the size in big endian
	blockHeader[1] = byte((block.Size) >> 16)
	blockHeader[2] = byte((block.Size) >> 8 & 0xFF)
	blockHeader[3] = byte((block.Size) >> 0 & 0xFF)
	if _, err := w.Write(blockHeader); err != nil {
		return err
	}
	if _, err := w.Write(block.Data); err != nil {
		return err
	}
	return nil
}

type VorbisComment struct {
	Name  string
	Value string
}

func (flac *FLAC) GetVorbisComment(key string) (string, error) {
	val, ok := flac.Tags[key]
	if !ok {
		return "", ErrTagNotFound
	}
	return val, nil
}

// The comment header is decoded as follows:
//
//  1. [vendor_length] = read an unsigned integer of 32 bits
//
//  2. [vendor_string] = read a UTF-8 vector as [vendor_length] octets
//
//  3. [user_comment_list_length] = read an unsigned integer of 32 bits
//
//  4. iterate [user_comment_list_length] times {
//
//  5. [length] = read an unsigned integer of 32 bits
//
//  6. this iteration's user comment = read a UTF-8 vector as [length] octets
//
//     }
//
//  7. [framing_bit] = read a single bit as boolean
func readVorbisComments(input io.Reader) ([]VorbisComment, string, error) {
	result := []VorbisComment{}

	// vendor
	vendorByte, err := readLengthData(input, binary.LittleEndian)
	if err != nil {
		return nil, "", err
	}

	// user_comment_list_length
	var length uint32
	err = binary.Read(input, binary.LittleEndian, &length)
	if err != nil {
		return nil, "", err
	}

	// iterate
	for i := 0; i < int(length); i++ {
		data, err := readLengthData(input, binary.LittleEndian)
		if err != nil {
			return nil, "", err
		}

		// Parse data
		vorbis := strings.SplitN(string(data), "=", 2)
		if len(vorbis) != 2 {
			return nil, "", ErrIncorrectTag
		}

		comment := VorbisComment{
			Name:  vorbis[0],
			Value: vorbis[1],
		}
		result = append(result, comment)
	}
	return result, string(vendorByte), nil
}

func serializeVorbisComments(comments map[string]string, vendorHeader string) []byte {
	// Serialize out the vorbis comments as a metadata block payload
	// Spawn out all the tag blobs first
	output := bytes.NewBuffer([]byte{})
	for key, value := range comments {
		line := key + "=" + value
		if err := writeLengthData(output, binary.LittleEndian, []byte(line)); err != nil {
			return []byte{}
		}
	}

	// Now that we have the payload content figured out, we can reconstruct its headers
	dataPayload, err := ioutil.ReadAll(output)
	if err != nil {
		return []byte{}
	}

	output.Grow(len(vendorHeader) + 4)
	output.Reset() // clear our scratchpad
	if err = writeLengthData(output, binary.LittleEndian, []byte(vendorHeader)); err != nil {
		return []byte{}
	}

	userCommentLength := uint32(len(comments))
	if err = binary.Write(output, binary.LittleEndian, userCommentLength); err != nil {
		return []byte{}
	}
	output.Write(dataPayload)

	payload, err := ioutil.ReadAll(output)
	if err != nil {
		return []byte{}
	}

	return payload
}
func (flac *FLAC) GetVorbisCommentInt(key string) (int, error) {
	comment, err := flac.GetVorbisComment(key)
	if err != nil {
		return 0, err
	}

	return strconv.Atoi(comment)
}

func (flac *FLAC) GetVorbisCommentTime(key string) (time.Time, error) {
	comment, err := flac.GetVorbisComment(key)
	if err != nil {
		return time.Now(), err
	}
	result, err := time.Parse("2006-01-02T15:04:05", comment)
	if err != nil {
		return time.Now(), err
	}
	return result, nil
}

type FlacMetadataBlockPicture struct {
	Type           int32
	MIME           string
	Description    string
	Width          int32
	Height         int32
	BitsPerPixel   int32
	NumberOfColors int32
	PictureData    []byte
}

func (flac *FLAC) GetMetadataBlockPicture() (*FlacMetadataBlockPicture, error) {
	for _, block := range flac.Blocks {
		if block.Type == FlacPicture {
			return readFlacPicture(bytes.NewReader(block.Data))
		}
	}

	return nil, ErrTagNotFound
}

func readFlacPicture(input io.Reader) (*FlacMetadataBlockPicture, error) {
	var picture FlacMetadataBlockPicture

	// Picture type
	err := binary.Read(input, binary.BigEndian, &picture.Type)
	if err != nil {
		return nil, err
	}

	// MIME
	MIMEBytes, err := readLengthData(input, binary.BigEndian)
	if err != nil {
		return nil, err
	}
	picture.MIME = string(MIMEBytes)

	// Description
	DescriptionBytes, err := readLengthData(input, binary.BigEndian)
	if err != nil {
		return nil, err
	}
	picture.Description = string(DescriptionBytes)

	// Width
	err = binary.Read(input, binary.BigEndian, &picture.Width)
	if err != nil {
		return nil, err
	}

	// Height
	err = binary.Read(input, binary.BigEndian, &picture.Height)
	if err != nil {
		return nil, err
	}

	// Bits per pixel
	err = binary.Read(input, binary.BigEndian, &picture.BitsPerPixel)
	if err != nil {
		return nil, err
	}

	// Number of colors
	err = binary.Read(input, binary.BigEndian, &picture.NumberOfColors)
	if err != nil {
		return nil, err
	}

	// Picture data
	picture.PictureData, err = readLengthData(input, binary.BigEndian)
	if err != nil {
		return nil, err
	}

	return &picture, nil
}
