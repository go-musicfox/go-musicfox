package tag

import (
	"io"
	"os"
)

func ReadFile(path string) (Metadata, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	return Read(file)
}

func Read(input io.ReadSeekCloser) (Metadata, error) {
	version := CheckVersion(input)
	switch version {
	case VersionID3v1:
		return ReadID3v1(input)
	case VersionID3v22:
		return ReadID3v22(input)
	case VersionID3v23:
		return ReadID3v23(input)
	case VersionID3v24:
		return ReadID3v24(input)
	case VersionMP4:
		return ReadMp4(input)
	case VersionFLAC:
		return ReadFLAC(input)
	default:
		return nil, ErrUnsupportedFormat
	}
}

func CheckVersion(input io.ReadSeeker) Version {
	if checkID3v24(input) {
		return VersionID3v24
	}

	if checkID3v23(input) {
		return VersionID3v23
	}

	if checkID3v22(input) {
		return VersionID3v22
	}

	if checkID3v1(input) {
		return VersionID3v1
	}

	if checkMp4(input) {
		return VersionMP4
	}

	if checkFLAC(input) {
		return VersionFLAC
	}

	return VersionUndefined
}
