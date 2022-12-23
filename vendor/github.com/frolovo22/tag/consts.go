package tag

const (
	// common consts.
	VersionUndefined Version = 0
	VersionID3v1     Version = 1
	VersionID3v22    Version = 2
	VersionID3v23    Version = 3
	VersionID3v24    Version = 4
	VersionMP4       Version = 5
	VersionFLAC      Version = 6

	// id3 consts.
	id3v1SizeHeader       = 128    // ID3v1 constant header size
	id3v1SizeType         = 3      // ID3v1 size of field 'Type'
	id3v1NoGenre          = 255    // no genre value
	id3v1NoTrackNumber    = 1      // no track number
	id3MarkerName         = "TAG"  // marker tag for id3 format
	id3MarkerValue        = "ID3"  // marker tag value for id3 format
	id3v2FrameTXXX        = "TXXX" // TXXX frame name for id3v2 format
	id3v22FrameHeaderSize = 6      // id3v22 frame header size

	// flac consts.
	FLACIdentifier = "fLaC" // flac format identifier

	// util consts.
	encodingUTF8    string = "UTF-8"
	encodingUTF16   string = "UTF-16"
	encodingUTF16BE string = "UTF-16BE"

	// mime types.
	mimeImageJPEG = "image/jpeg"
	mimeImagePNG  = "image/png"
	mimeImageLink = "-->"
)
