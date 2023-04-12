package minimp3

/*
#define MINIMP3_IMPLEMENTATION

#include "minimp3.h"
#include <stdlib.h>
#include <stdio.h>

int decode(mp3dec_t *dec, mp3dec_frame_info_t *info, unsigned char *data, int *length, unsigned char *decoded, int *decoded_length) {
    int samples;
    short pcm[MINIMP3_MAX_SAMPLES_PER_FRAME];
    samples = mp3dec_decode_frame(dec, data, *length, pcm, info);
    *decoded_length = samples * info->channels * 2;
    *length -= info->frame_bytes;
    unsigned char buffer[samples * info->channels * 2];
    memcpy(buffer, (unsigned char*)&(pcm), sizeof(short) * samples * info->channels);
    memcpy(decoded, buffer, sizeof(short) * samples * info->channels);
    return info->frame_bytes;
}
*/
import "C"
import (
	"context"
	"io"
	"sync"
	"time"
	"unsafe"
)

const maxSamplesPerFrame = 1152 * 2

// Decoder decode the mp3 stream by minimp3
type Decoder struct {
	data []byte

	needFillChan  chan struct{}
	decoderLocker *sync.Mutex
	decodedData   []byte
	decode        C.mp3dec_t
	info          C.mp3dec_frame_info_t
	context       context.Context
	contextCancel context.CancelFunc
	SampleRate    int
	Channels      int
	Kbps          int
	Layer         int
	err           error
}

// BufferSize Decoded data buffer size.
var BufferSize = 1024 * 10

// WaitForDataDuration wait for the data time duration.
var WaitForDataDuration = time.Millisecond * 10

// NewDecoder decode mp3 stream and get a Decoder for read the raw data to play.
func NewDecoder(reader io.Reader) (dec *Decoder, err error) {
	dec = new(Decoder)
	dec.decoderLocker = &sync.Mutex{}
	dec.needFillChan = make(chan struct{}, 1)
	dec.context, dec.contextCancel = context.WithCancel(context.Background())
	dec.decode = C.mp3dec_t{}
	C.mp3dec_init(&dec.decode)
	dec.info = C.mp3dec_frame_info_t{}
	go func() {
		for {
			select {
			case <-dec.context.Done():
				return
			default:
			}
			if len(dec.decodedData) > BufferSize {
				<-dec.needFillChan
				continue
			}
			var decoded = [maxSamplesPerFrame * 2]byte{}
			var decodedLength = C.int(0)
			var length = C.int(len(dec.data))
			for len(dec.data) < maxSamplesPerFrame*2 {
				var data = make([]byte, 512)
				var n int
				n, err = io.ReadFull(reader, data)
				dec.err = err

				dec.data = append(dec.data, data[:n]...)
				if err != nil {
					break
				}
			}
			if err != nil && err != io.ErrUnexpectedEOF {
				break
			}
			frameSize := C.decode(&dec.decode, &dec.info,
				(*C.uchar)(unsafe.Pointer(&dec.data[0])),
				&length, (*C.uchar)(unsafe.Pointer(&decoded[0])),
				&decodedLength)
			if int(frameSize) == 0 {
				<-time.After(WaitForDataDuration)
				continue
			}
			dec.SampleRate = int(dec.info.hz)
			dec.Channels = int(dec.info.channels)
			dec.Kbps = int(dec.info.bitrate_kbps)
			dec.Layer = int(dec.info.layer)
			dec.decoderLocker.Lock()
			dec.decodedData = append(dec.decodedData, decoded[:decodedLength]...)
			if int(frameSize) <= len(dec.data) {
				dec.data = dec.data[int(frameSize):]
			}
			dec.decoderLocker.Unlock()
		}
	}()
	return
}

// Started check the record mp3 stream started ot not.
func (dec *Decoder) Started() (channel chan error) {
	channel = make(chan error)
	go func() {
		for {
			select {
			case <-dec.context.Done():
				channel <- nil
				return
			default:
			}
			if len(dec.decodedData) != 0 {
				channel <- nil
				return
			} else {
				if dec.err != nil {
					channel <- dec.err
				}
				<-time.After(time.Millisecond * 100)
			}
		}
	}()
	return
}

// Read read the raw stream
func (dec *Decoder) Read(data []byte) (n int, err error) {
	for len(dec.decodedData) == 0 {
		select {
		case <-dec.context.Done():
			return 0, io.EOF
		default:
		}
		if dec.err == io.EOF {
			err = io.EOF
			return
		}
		<-time.After(WaitForDataDuration)
	}
	dec.decoderLocker.Lock()
	defer dec.decoderLocker.Unlock()
	n = copy(data, dec.decodedData[:])
	dec.decodedData = dec.decodedData[n:]
	if len(dec.decodedData) <= BufferSize {
		select {
		case dec.needFillChan <- struct{}{}:
		default:
		}
	}
	return
}

// Close stop the decode mp3 stream cycle.
func (dec *Decoder) Close() {
	if dec.contextCancel != nil {
		dec.contextCancel()
	}
}

// DecodeFull put all of the mp3 data to decode.
func DecodeFull(mp3 []byte) (dec *Decoder, decodedData []byte, err error) {
	dec = new(Decoder)
	dec.decode = C.mp3dec_t{}
	C.mp3dec_init(&dec.decode)
	info := C.mp3dec_frame_info_t{}
	var length = C.int(len(mp3))
	for {
		var decoded = [maxSamplesPerFrame * 2]byte{}
		var decodedLength = C.int(0)
		frameSize := C.decode(&dec.decode,
			&info, (*C.uchar)(unsafe.Pointer(&mp3[0])),
			&length, (*C.uchar)(unsafe.Pointer(&decoded[0])),
			&decodedLength)
		if int(frameSize) == 0 {
			break
		}
		decodedData = append(decodedData, decoded[:decodedLength]...)
		if int(frameSize) < len(mp3) {
			mp3 = mp3[int(frameSize):]
		}
		dec.SampleRate = int(info.hz)
		dec.Channels = int(info.channels)
		dec.Kbps = int(info.bitrate_kbps)
		dec.Layer = int(info.layer)
	}
	return
}
