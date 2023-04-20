// Copyright 2015-2016 Cocoon Labs Ltd.
//
// See LICENSE file for terms and conditions.

// Package libflac provides Go bindings to the libFLAC codec library.
package libflac

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"sync"
	"unsafe"
)

/*
#cgo pkg-config: flac
#include <stdlib.h>

#include "FLAC/stream_decoder.h"
#include "FLAC/stream_encoder.h"

extern void
decoderErrorCallback_cgo(const FLAC__StreamDecoder *,
		         FLAC__StreamDecoderErrorStatus,
		         void *);


extern void
decoderMetadataCallback_cgo(const FLAC__StreamDecoder *,
			    const FLAC__StreamMetadata *,
			    void *);

extern FLAC__StreamDecoderWriteStatus
decoderWriteCallback_cgo(const FLAC__StreamDecoder *,
		         const FLAC__Frame *,
		         const FLAC__int32 **,
		         void *);

FLAC__StreamDecoderReadStatus
decoderReadCallback_cgo(const FLAC__StreamDecoder *,
		        const FLAC__byte *,
			size_t *,
		        void *);

FLAC__StreamDecoderSeekStatus
decoderSeekCallback_cgo(const FLAC__StreamDecoder *,
			FLAC__uint64,
		        void *);

FLAC__StreamDecoderTellStatus
decoderTellCallback_cgo(const FLAC__StreamDecoder *,
			FLAC__uint64*,
		        void *);

FLAC__StreamDecoderLengthStatus
decoderLengthCallback_cgo(const FLAC__StreamDecoder *,
			FLAC__uint64*,
		        void *);

FLAC__bool
decoderEofCallback_cgo(const FLAC__StreamDecoder *,
                void *);

FLAC__StreamEncoderWriteStatus
encoderWriteCallback_cgo(const FLAC__StreamEncoder *,
			 const FLAC__byte *,
			 size_t, unsigned,
			 unsigned,
		         void *);

FLAC__StreamEncoderSeekStatus
encoderSeekCallback_cgo(const FLAC__StreamEncoder *,
			FLAC__uint64,
		        void *);

FLAC__StreamEncoderTellStatus
encoderTellCallback_cgo(const FLAC__StreamEncoder *,
			FLAC__uint64 *,
		        void *);

extern const char *
get_decoder_error_str(FLAC__StreamDecoderErrorStatus status);

extern int
get_decoder_channels(FLAC__StreamMetadata *metadata);

extern int
get_decoder_depth(FLAC__StreamMetadata *metadata);

extern int
get_decoder_rate(FLAC__StreamMetadata *metadata);

extern void
get_audio_samples(int32_t *output, const FLAC__int32 **input,
                  unsigned int blocksize, unsigned int channels);

*/
import "C"

// Go 1.6 does not allow us to pass go pointers into C that will be stored and
// used in callbacks and suggests we use a value lookup for pointer callbacks.
// https://github.com/golang/proposal/blob/master/design/12416-cgo-pointers.md

// Concurrent safe map for mapping pointers between C and Go.
type decoderPtrMap struct {
	sync.RWMutex
	ptrs map[uintptr]*Decoder
}

func (m *decoderPtrMap) get(d *C.FLAC__StreamDecoder) *Decoder {
	ptr := uintptr(unsafe.Pointer(d))
	m.RLock()
	defer m.RUnlock()
	return m.ptrs[ptr]
}

func (m *decoderPtrMap) add(d *Decoder) {
	m.Lock()
	defer m.Unlock()
	m.ptrs[uintptr(unsafe.Pointer(d.d))] = d
}

func (m *decoderPtrMap) del(d *Decoder) {
	m.Lock()
	defer m.Unlock()
	delete(m.ptrs, uintptr(unsafe.Pointer(d.d)))
}

var decoderPtrs = decoderPtrMap{ptrs: make(map[uintptr]*Decoder)}

// Concurrent safe map for mapping pointers between C and Go.
type encoderPtrMap struct {
	sync.RWMutex
	ptrs map[uintptr]*Encoder
}

func (m *encoderPtrMap) get(e *C.FLAC__StreamEncoder) *Encoder {
	ptr := uintptr(unsafe.Pointer(e))
	m.RLock()
	defer m.RUnlock()
	return m.ptrs[ptr]
}

func (m *encoderPtrMap) add(e *Encoder) {
	m.Lock()
	defer m.Unlock()
	m.ptrs[uintptr(unsafe.Pointer(e.e))] = e
}

func (m *encoderPtrMap) del(e *Encoder) {
	m.Lock()
	defer m.Unlock()
	delete(m.ptrs, uintptr(unsafe.Pointer(e.e)))
}

var encoderPtrs = encoderPtrMap{ptrs: make(map[uintptr]*Encoder)}

type FlacWriter interface {
	io.Writer
	io.Closer
	io.Seeker
}

// Frame is an interleaved buffer of audio data with the specified parameters.
type Frame struct {
	Channels int
	Depth    int
	Rate     int
	Buffer   []int32
}

// Decoder is a FLAC decoder.
type Decoder struct {
	d        *C.FLAC__StreamDecoder
	reader   io.ReadSeekCloser
	Channels int
	Depth    int
	Rate     int
	error    bool
	errorStr string
	frame    *Frame
	l        sync.Mutex
}

// Encoder is a FLAC encoder.
type Encoder struct {
	e        *C.FLAC__StreamEncoder
	writer   FlacWriter
	Channels int
	Depth    int
	Rate     int
}

//export decoderErrorCallback
func decoderErrorCallback(d *C.FLAC__StreamDecoder, status C.FLAC__StreamDecoderErrorStatus, data unsafe.Pointer) {
	decoder := decoderPtrs.get(d)
	decoder.error = true
	decoder.errorStr = C.GoString(C.get_decoder_error_str(status))
}

//export decoderMetadataCallback
func decoderMetadataCallback(d *C.FLAC__StreamDecoder, metadata *C.FLAC__StreamMetadata, data unsafe.Pointer) {
	decoder := decoderPtrs.get(d)
	if metadata._type == C.FLAC__METADATA_TYPE_STREAMINFO {
		decoder.Channels = int(C.get_decoder_channels(metadata))
		decoder.Depth = int(C.get_decoder_depth(metadata))
		decoder.Rate = int(C.get_decoder_rate(metadata))
	}
}

//export decoderWriteCallback
func decoderWriteCallback(d *C.FLAC__StreamDecoder, frame *C.FLAC__Frame, buffer **C.FLAC__int32, data unsafe.Pointer) C.FLAC__StreamDecoderWriteStatus {
	decoder := decoderPtrs.get(d)
	blocksize := int(frame.header.blocksize)
	decoder.frame = new(Frame)
	f := decoder.frame
	f.Channels = decoder.Channels
	f.Depth = decoder.Depth
	f.Rate = decoder.Rate
	f.Buffer = make([]int32, blocksize*decoder.Channels)
	C.get_audio_samples((*C.int32_t)(&f.Buffer[0]), buffer, C.uint(blocksize), C.uint(decoder.Channels))
	return C.FLAC__STREAM_DECODER_WRITE_STATUS_CONTINUE
}

//export decoderReadCallback
func decoderReadCallback(d *C.FLAC__StreamDecoder, buffer *C.FLAC__byte, bytes *C.size_t, data unsafe.Pointer) C.FLAC__StreamDecoderReadStatus {
	decoder := decoderPtrs.get(d)
	numBytes := int(*bytes)
	if numBytes <= 0 {
		return C.FLAC__STREAM_DECODER_READ_STATUS_ABORT
	}
	hdr := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(buffer)),
		Len:  numBytes,
		Cap:  numBytes,
	}
	buf := *(*[]byte)(unsafe.Pointer(&hdr))
	n, err := decoder.reader.Read(buf)
	*bytes = C.size_t(n)
	if err == io.EOF && n == 0 {
		return C.FLAC__STREAM_DECODER_READ_STATUS_END_OF_STREAM
	} else if err != nil && err != io.EOF {
		return C.FLAC__STREAM_DECODER_READ_STATUS_ABORT
	}
	return C.FLAC__STREAM_DECODER_READ_STATUS_CONTINUE
}

//export decoderSeekCallback
func decoderSeekCallback(e *C.FLAC__StreamDecoder, absPos C.FLAC__uint64, data unsafe.Pointer) C.FLAC__StreamDecoderSeekStatus {
	decoder := decoderPtrs.get(e)
	decoder.l.Lock()
	defer decoder.l.Unlock()
	_, err := decoder.reader.Seek(int64(absPos), io.SeekStart)
	if err != nil {
		return C.FLAC__STREAM_DECODER_SEEK_STATUS_ERROR
	}
	return C.FLAC__STREAM_DECODER_SEEK_STATUS_OK
}

//export decoderTellCallback
func decoderTellCallback(e *C.FLAC__StreamDecoder, absPos *C.FLAC__uint64, data unsafe.Pointer) C.FLAC__StreamDecoderTellStatus {
	decoder := decoderPtrs.get(e)
	decoder.l.Lock()
	defer decoder.l.Unlock()
	pos, err := decoder.reader.Seek(0, io.SeekCurrent)
	if err != nil {
		return C.FLAC__STREAM_DECODER_TELL_STATUS_ERROR
	}
	*absPos = C.FLAC__uint64(pos)
	return C.FLAC__STREAM_DECODER_TELL_STATUS_OK
}

//export decoderLengthCallback
func decoderLengthCallback(e *C.FLAC__StreamDecoder, len *C.FLAC__uint64, data unsafe.Pointer) C.FLAC__StreamDecoderLengthStatus {
	decoder := decoderPtrs.get(e)
	_, l, err := decoder.Tell()
	if err != nil {
		return C.FLAC__STREAM_DECODER_LENGTH_STATUS_ERROR
	}

	*len = C.FLAC__uint64(l)
	return C.FLAC__STREAM_DECODER_LENGTH_STATUS_OK
}

//export decoderEofCallback
func decoderEofCallback(e *C.FLAC__StreamDecoder, data unsafe.Pointer) C.FLAC__bool {
	decoder := decoderPtrs.get(e)
	nowPos, l, err := decoder.Tell()
	if err != nil {
		return C.FLAC__bool(0)
	}
	if nowPos < l {
		return C.FLAC__bool(0)
	}
	return C.FLAC__bool(1)
}

// NewDecoder creates a new Decoder object.
func NewDecoder(name string) (d *Decoder, err error) {
	d = new(Decoder)
	d.d = C.FLAC__stream_decoder_new()
	if d.d == nil {
		return nil, errors.New("failed to create decoder")
	}
	c := C.CString(name)
	defer C.free(unsafe.Pointer(c))
	//runtime.SetFinalizer(d, (*Decoder).Close)
	status := C.FLAC__stream_decoder_init_file(d.d, c,
		(C.FLAC__StreamDecoderWriteCallback)(unsafe.Pointer(C.decoderWriteCallback_cgo)),
		(C.FLAC__StreamDecoderMetadataCallback)(unsafe.Pointer(C.decoderMetadataCallback_cgo)),
		(C.FLAC__StreamDecoderErrorCallback)(unsafe.Pointer(C.decoderErrorCallback_cgo)),
		nil)
	if status != C.FLAC__STREAM_DECODER_INIT_STATUS_OK {
		return nil, errors.New("failed to open file")
	}
	decoderPtrs.add(d)
	ret := C.FLAC__stream_decoder_process_until_end_of_metadata(d.d)
	if ret == 0 || d.error || d.Channels == 0 {
		return nil, fmt.Errorf("failed to process metadata %s", d.errorStr)
	}
	return
}

// NewDecoderReader creates a new Decoder object from a Reader.
func NewDecoderReader(reader io.ReadSeekCloser) (d *Decoder, err error) {
	d = new(Decoder)
	d.d = C.FLAC__stream_decoder_new()
	if d.d == nil {
		return nil, errors.New("failed to create decoder")
	}
	d.reader = reader
	//runtime.SetFinalizer(d, (*Decoder).Close)
	status := C.FLAC__stream_decoder_init_stream(d.d,
		(C.FLAC__StreamDecoderReadCallback)(unsafe.Pointer(C.decoderReadCallback_cgo)),
		(C.FLAC__StreamDecoderSeekCallback)(unsafe.Pointer(C.decoderSeekCallback_cgo)),
		(C.FLAC__StreamDecoderTellCallback)(unsafe.Pointer(C.decoderTellCallback_cgo)),
		(C.FLAC__StreamDecoderLengthCallback)(unsafe.Pointer(C.decoderLengthCallback_cgo)),
		(C.FLAC__StreamDecoderEofCallback)(unsafe.Pointer(C.decoderEofCallback_cgo)),
		(C.FLAC__StreamDecoderWriteCallback)(unsafe.Pointer(C.decoderWriteCallback_cgo)),
		(C.FLAC__StreamDecoderMetadataCallback)(unsafe.Pointer(C.decoderMetadataCallback_cgo)),
		(C.FLAC__StreamDecoderErrorCallback)(unsafe.Pointer(C.decoderErrorCallback_cgo)),
		nil)
	if status != C.FLAC__STREAM_DECODER_INIT_STATUS_OK {
		return nil, errors.New("failed to open stream")
	}
	decoderPtrs.add(d)
	ret := C.FLAC__stream_decoder_process_until_end_of_metadata(d.d)
	if ret == 0 || d.error || d.Channels == 0 {
		return nil, fmt.Errorf("failed to process metadata %s", d.errorStr)
	}
	return
}

// Close closes a decoder and frees the resources.
func (d *Decoder) Close() {
	d.l.Lock()
	defer d.l.Unlock()
	if d.d != nil {
		C.FLAC__stream_decoder_delete(d.d)
		decoderPtrs.del(d)
		d.d = nil
	}
	if d.reader != nil {
		_ = d.reader.Close()
	}
	//runtime.SetFinalizer(d, nil)
}

// ReadFrame reads a frame of audio data from the decoder.
func (d *Decoder) ReadFrame() (f *Frame, err error) {
	ret := C.FLAC__stream_decoder_process_single(d.d)
	if ret == 0 || d.error {
		return nil, errors.New("error reading frame")
	}
	state := C.FLAC__stream_decoder_get_state(d.d)
	if state == C.FLAC__STREAM_DECODER_END_OF_STREAM {
		err = io.EOF
	}
	f = d.frame
	d.frame = nil
	return
}

func (d *Decoder) Seek(pos uint64) (uint64, error) {
	var res C.FLAC__bool = C.FLAC__stream_decoder_seek_absolute(d.d, C.FLAC__uint64(pos))
	if int32(res) == 0 {
		return 0, errors.New("seek failed")
	}
	return pos, nil
}

func (d *Decoder) Tell() (curPos int64, len int64, err error) {
	d.l.Lock()
	defer d.l.Unlock()
	curPos, err = d.reader.Seek(0, io.SeekCurrent)
	if err != nil {
		return
	}
	len, err = d.reader.Seek(0, io.SeekEnd)
	if err != nil {
		return
	}
	_, err = d.reader.Seek(curPos, io.SeekStart)
	if err != nil {
		return
	}
	return
}

// NewEncoder creates a new Encoder object.
func NewEncoder(name string, channels int, depth int, rate int) (e *Encoder, err error) {
	if channels == 0 {
		return nil, errors.New("channels must be greater than 0")
	}
	if !(depth == 16 || depth == 24) {
		return nil, errors.New("depth must be 16 or 24")
	}
	e = new(Encoder)
	e.e = C.FLAC__stream_encoder_new()
	if e.e == nil {
		return nil, errors.New("failed to create decoder")
	}
	encoderPtrs.add(e)
	c := C.CString(name)
	defer C.free(unsafe.Pointer(c))
	//runtime.SetFinalizer(e, (*Encoder).Close)
	C.FLAC__stream_encoder_set_channels(e.e, C.uint(channels))
	C.FLAC__stream_encoder_set_bits_per_sample(e.e, C.uint(depth))
	C.FLAC__stream_encoder_set_sample_rate(e.e, C.uint(rate))
	status := C.FLAC__stream_encoder_init_file(e.e, c, nil, nil)
	if status != C.FLAC__STREAM_ENCODER_INIT_STATUS_OK {
		return nil, errors.New("failed to open file")
	}
	e.Channels = channels
	e.Depth = depth
	e.Rate = rate
	return
}

//export encoderWriteCallback
func encoderWriteCallback(e *C.FLAC__StreamEncoder, buffer *C.FLAC__byte, bytes C.size_t, samples, current_frame C.unsigned, data unsafe.Pointer) C.FLAC__StreamEncoderWriteStatus {
	encoder := encoderPtrs.get(e)
	numBytes := int(bytes)
	if numBytes <= 0 {
		return C.FLAC__STREAM_DECODER_READ_STATUS_ABORT
	}
	hdr := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(buffer)),
		Len:  numBytes,
		Cap:  numBytes,
	}
	buf := *(*[]byte)(unsafe.Pointer(&hdr))
	_, err := encoder.writer.Write(buf)
	if err != nil {
		return C.FLAC__STREAM_ENCODER_WRITE_STATUS_FATAL_ERROR
	}
	return C.FLAC__STREAM_ENCODER_WRITE_STATUS_OK
}

//export encoderSeekCallback
func encoderSeekCallback(e *C.FLAC__StreamEncoder, absPos C.FLAC__uint64, data unsafe.Pointer) C.FLAC__StreamEncoderWriteStatus {
	encoder := encoderPtrs.get(e)
	_, err := encoder.writer.Seek(int64(absPos), 0)
	if err != nil {
		return C.FLAC__STREAM_ENCODER_SEEK_STATUS_ERROR
	}
	return C.FLAC__STREAM_ENCODER_SEEK_STATUS_OK
}

//export encoderTellCallback
func encoderTellCallback(e *C.FLAC__StreamEncoder, absPos *C.FLAC__uint64, data unsafe.Pointer) C.FLAC__StreamEncoderWriteStatus {
	encoder := encoderPtrs.get(e)
	newPos, err := encoder.writer.Seek(0, 1)
	if err != nil {
		return C.FLAC__STREAM_ENCODER_TELL_STATUS_ERROR
	}
	*absPos = C.FLAC__uint64(newPos)
	return C.FLAC__STREAM_ENCODER_TELL_STATUS_OK
}

// NewEncoderWriter creates a new Encoder object from a FlacWriter.
func NewEncoderWriter(writer FlacWriter, channels int, depth int, rate int) (e *Encoder, err error) {
	if channels == 0 {
		return nil, errors.New("channels must be greater than 0")
	}
	if !(depth == 16 || depth == 24) {
		return nil, errors.New("depth must be 16 or 24")
	}
	e = new(Encoder)
	e.e = C.FLAC__stream_encoder_new()
	if e.e == nil {
		return nil, errors.New("failed to create decoder")
	}
	encoderPtrs.add(e)
	e.writer = writer
	//runtime.SetFinalizer(e, (*Encoder).Close)
	C.FLAC__stream_encoder_set_channels(e.e, C.uint(channels))
	C.FLAC__stream_encoder_set_bits_per_sample(e.e, C.uint(depth))
	C.FLAC__stream_encoder_set_sample_rate(e.e, C.uint(rate))
	status := C.FLAC__stream_encoder_init_stream(e.e,
		(C.FLAC__StreamEncoderWriteCallback)(unsafe.Pointer(C.encoderWriteCallback_cgo)),
		(C.FLAC__StreamEncoderSeekCallback)(unsafe.Pointer(C.encoderSeekCallback_cgo)),
		(C.FLAC__StreamEncoderTellCallback)(unsafe.Pointer(C.encoderTellCallback_cgo)),
		nil, nil)
	if status != C.FLAC__STREAM_ENCODER_INIT_STATUS_OK {
		return nil, errors.New("failed to open file")
	}
	e.Channels = channels
	e.Depth = depth
	e.Rate = rate
	return
}

// WriteFrame writes a frame of audio data to the encoder.
func (e *Encoder) WriteFrame(f Frame) (err error) {
	if f.Channels != e.Channels || f.Depth != e.Depth || f.Rate != e.Rate {
		return errors.New("frame type does not match encoder")
	}
	if len(f.Buffer) == 0 {
		return
	}
	ret := C.FLAC__stream_encoder_process_interleaved(e.e, (*C.FLAC__int32)(&f.Buffer[0]), C.uint(len(f.Buffer)/e.Channels))
	if ret == 0 {
		return errors.New("error encoding frame")
	}
	return
}

// Close closes an encoder and frees the resources.
func (e *Encoder) Close() {
	if e.e != nil {
		C.FLAC__stream_encoder_finish(e.e)
		C.FLAC__stream_encoder_delete(e.e)
		encoderPtrs.del(e)
		e.e = nil
	}
	//runtime.SetFinalizer(e, nil)
}
