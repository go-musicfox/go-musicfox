package flacpicture

import "errors"

var (
	// ErrorNotPictureMetadataBlock is returned if the metadata provided is not a picture block.
	ErrorNotPictureMetadataBlock = errors.New("Not a picture metadata block")
	// ErrorUnsupportedMIME is returned if the provided image MIME type is unsupported.
	ErrorUnsupportedMIME = errors.New("Unsupported MIME")
)
