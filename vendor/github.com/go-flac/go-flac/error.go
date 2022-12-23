package flac

import "errors"

var (
	// ErrorNoFLACHeader indicates that "fLaC" marker not found at the beginning of the file
	ErrorNoFLACHeader = errors.New("fLaC head incorrect")
	// ErrorNoStreamInfo indicates that StreamInfo Metablock not present or is not the first Metablock
	ErrorNoStreamInfo = errors.New("stream info not present")
	// ErrorStreamInfoEarlyEOF indicates that an unexpected EOF is hit while reading StreamInfo Metablock
	ErrorStreamInfoEarlyEOF = errors.New("unexpected end of stream while reading stream info")
	// ErrorNoSyncCode indicates that the frames are malformed as the sync code is not present after the last Metablock
	ErrorNoSyncCode = errors.New("frames do not begin with sync code")
)
