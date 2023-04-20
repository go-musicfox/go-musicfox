#include "FLAC/stream_decoder.h"
#include "FLAC/stream_encoder.h"

#include "_cgo_export.h"

void
decoderErrorCallback_cgo(const FLAC__StreamDecoder *decoder,
		         FLAC__StreamDecoderErrorStatus status,
		         void *data)
{
    decoderErrorCallback((FLAC__StreamDecoder *)decoder, status, data);
}

void
decoderMetadataCallback_cgo(const FLAC__StreamDecoder *decoder,
			    const FLAC__StreamMetadata *metadata,
			    void *data)
{
    decoderMetadataCallback((FLAC__StreamDecoder *)decoder,
			    (FLAC__StreamMetadata *)metadata, data);
}

FLAC__StreamDecoderSeekStatus
decoderSeekCallback_cgo(const FLAC__StreamDecoder *decoder,
                FLAC__uint64 absolute_byte_offset,
                void *data)
{
	return decoderSeekCallback((FLAC__StreamDecoder *)decoder,
    			       absolute_byte_offset, data);
}

FLAC__StreamDecoderTellStatus
decoderTellCallback_cgo(const FLAC__StreamDecoder *decoder,
                FLAC__uint64 *absolute_byte_offset,
                void *data)
{
	return decoderTellCallback((FLAC__StreamDecoder *)decoder,
    			       (FLAC__uint64 *)absolute_byte_offset, data);
}

FLAC__StreamDecoderLengthStatus
decoderLengthCallback_cgo(const FLAC__StreamDecoder *decoder,
                FLAC__uint64 *stream_length,
                void *data)
{
	return decoderLengthCallback((FLAC__StreamDecoder *)decoder,
    			       (FLAC__uint64 *)stream_length, data);
}

FLAC__bool
decoderEofCallback_cgo(const FLAC__StreamDecoder *decoder,
                void *data)
{
	return decoderEofCallback((FLAC__StreamDecoder *)decoder, data);
}

FLAC__StreamDecoderWriteStatus
decoderWriteCallback_cgo(const FLAC__StreamDecoder *decoder,
		         const FLAC__Frame *frame,
		         const FLAC__int32 **buffer,
		         void *data)
{
    return decoderWriteCallback((FLAC__StreamDecoder *)decoder,
				(FLAC__Frame *)frame,
				(FLAC__int32 **)buffer, data);
}

FLAC__StreamDecoderReadStatus
decoderReadCallback_cgo(const FLAC__StreamDecoder *decoder,
		        const FLAC__byte buffer[],
			size_t *bytes,
		        void *data)
{
    return decoderReadCallback((FLAC__StreamDecoder *)decoder,
			       (FLAC__byte *)buffer,
			       bytes, data);
}

FLAC__StreamEncoderWriteStatus
encoderWriteCallback_cgo(const FLAC__StreamEncoder *encoder,
			 const FLAC__byte buffer[],
			 size_t bytes, unsigned samples,
			 unsigned current_frame,
		         void *data)
{
    return encoderWriteCallback((FLAC__StreamEncoder *)encoder,
				(FLAC__byte *)buffer, bytes, samples,
				current_frame, data);
}

FLAC__StreamEncoderSeekStatus
encoderSeekCallback_cgo(const FLAC__StreamEncoder *encoder,
			FLAC__uint64 absolute_byte_offset,
		        void *data)
{
    return encoderSeekCallback((FLAC__StreamEncoder *)encoder,
			       absolute_byte_offset, data);
}

FLAC__StreamEncoderTellStatus
encoderTellCallback_cgo(const FLAC__StreamEncoder *encoder,
			FLAC__uint64 *absolute_byte_offset,
		        void *data)
{
    return encoderTellCallback((FLAC__StreamEncoder *)encoder,
			       absolute_byte_offset, data);
}

extern const char *
get_decoder_error_str(FLAC__StreamDecoderErrorStatus status)
{
     return FLAC__StreamDecoderErrorStatusString[status];
}

extern int
get_decoder_channels(FLAC__StreamMetadata *metadata)
{
     return metadata->data.stream_info.channels;
}

extern int
get_decoder_depth(FLAC__StreamMetadata *metadata)
{
     return metadata->data.stream_info.bits_per_sample;
}

extern int
get_decoder_rate(FLAC__StreamMetadata *metadata)
{
     return metadata->data.stream_info.sample_rate;
}

extern void
get_audio_samples(int32_t *output, const FLAC__int32 **input,
                  unsigned int blocksize, unsigned int channels)
{
    unsigned int i, j, samples = blocksize * channels;
    for (i = 0; i < blocksize; i++)
        for (j = 0; j < channels; j++)
            output[i * channels + j] = input[j][i];
}
