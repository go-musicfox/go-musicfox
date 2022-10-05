package flac

import (
	"github.com/icza/bitio"
	"github.com/mewkiz/flac/frame"
	iobits "github.com/mewkiz/flac/internal/bits"
	"github.com/mewkiz/pkg/errutil"
)

// --- [ Subframe ] ------------------------------------------------------------

// encodeSubframe encodes the given subframe, writing to bw.
func encodeSubframe(bw *bitio.Writer, hdr frame.Header, subframe *frame.Subframe) error {
	// Encode subframe header.
	if err := encodeSubframeHeader(bw, subframe.SubHeader); err != nil {
		return errutil.Err(err)
	}

	// Encode audio samples.
	switch subframe.Pred {
	case frame.PredConstant:
		if err := encodeConstantSamples(bw, hdr.BitsPerSample, subframe.Samples); err != nil {
			return errutil.Err(err)
		}
	case frame.PredVerbatim:
		if err := encodeVerbatimSamples(bw, hdr, subframe.Samples); err != nil {
			return errutil.Err(err)
		}
	//case frame.PredFixed:
	//	if err := encodeFixedSamples(bw, hdr, subframe.Samples, subframe.Order); err != nil {
	//		return errutil.Err(err)
	//	}
	//case frame.PredFIR:
	//	if err := encodeFIRSamples(bw, hdr, subframe.Samples, subframe.Order); err != nil {
	//		return errutil.Err(err)
	//	}
	default:
		return errutil.Newf("support for prediction method %v not yet implemented", subframe.Pred)
	}
	return nil
}

// --- [ Subframe header ] -----------------------------------------------------

// encodeSubframeHeader encodes the given subframe header, writing to bw.
func encodeSubframeHeader(bw *bitio.Writer, hdr frame.SubHeader) error {
	// Zero bit padding, to prevent sync-fooling string of 1s.
	if err := bw.WriteBits(0x0, 1); err != nil {
		return errutil.Err(err)
	}

	// Subframe type:
	//     000000 : SUBFRAME_CONSTANT
	//     000001 : SUBFRAME_VERBATIM
	//     00001x : reserved
	//     0001xx : reserved
	//     001xxx : if(xxx <= 4) SUBFRAME_FIXED, xxx=order ; else reserved
	//     01xxxx : reserved
	//     1xxxxx : SUBFRAME_LPC, xxxxx=order-1
	var bits uint64
	switch hdr.Pred {
	case frame.PredConstant:
		// 000000 : SUBFRAME_CONSTANT
		bits = 0x00
	case frame.PredVerbatim:
		// 000001 : SUBFRAME_VERBATIM
		bits = 0x01
	case frame.PredFixed:
		// 001xxx : if(xxx <= 4) SUBFRAME_FIXED, xxx=order ; else reserved
		bits = 0x08 | uint64(hdr.Order)
	case frame.PredFIR:
		// 1xxxxx : SUBFRAME_LPC, xxxxx=order-1
		bits = 0x20 | uint64(hdr.Order-1)
	}
	if err := bw.WriteBits(bits, 6); err != nil {
		return errutil.Err(err)
	}

	// <1+k> 'Wasted bits-per-sample' flag:
	//
	//     0 : no wasted bits-per-sample in source subblock, k=0
	//     1 : k wasted bits-per-sample in source subblock, k-1 follows, unary coded; e.g. k=3 => 001 follows, k=7 => 0000001 follows.
	hasWastedBits := hdr.Wasted > 0
	if err := bw.WriteBool(hasWastedBits); err != nil {
		return errutil.Err(err)
	}
	if hasWastedBits {
		if err := iobits.WriteUnary(bw, uint64(hdr.Wasted)); err != nil {
			return errutil.Err(err)
		}
	}
	return nil
}

// --- [ Constant samples ] ----------------------------------------------------

// encodeConstantSamples stores the given constant sample, writing to bw.
func encodeConstantSamples(bw *bitio.Writer, bps byte, samples []int32) error {
	sample := samples[0]
	for _, s := range samples[1:] {
		if sample != s {
			return errutil.Newf("constant sample mismatch; expected %v, got %v", sample, s)
		}
	}
	// Unencoded constant value of the subblock, n = frame's bits-per-sample.
	if err := bw.WriteBits(uint64(sample), bps); err != nil {
		return errutil.Err(err)
	}
	return nil
}

// --- [ Verbatim samples ] ----------------------------------------------------

// encodeVerbatimSamples stores the given samples verbatim (uncompressed),
// writing to bw.
func encodeVerbatimSamples(bw *bitio.Writer, hdr frame.Header, samples []int32) error {
	// Unencoded subblock; n = frame's bits-per-sample, i = frame's blocksize.
	if int(hdr.BlockSize) != len(samples) {
		return errutil.Newf("block size and sample count mismatch; expected %d, got %d", hdr.BlockSize, len(samples))
	}
	for _, sample := range samples {
		if err := bw.WriteBits(uint64(sample), hdr.BitsPerSample); err != nil {
			return errutil.Err(err)
		}
	}
	return nil
}
