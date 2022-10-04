package frame

import (
	"errors"
	"fmt"
	"log"

	"github.com/mewkiz/flac/internal/bits"
)

// A Subframe contains the encoded audio samples from one channel of an audio
// block (a part of the audio stream).
//
// ref: https://www.xiph.org/flac/format.html#subframe
type Subframe struct {
	// Subframe header.
	SubHeader
	// Unencoded audio samples. Samples is initially nil, and gets populated by a
	// call to Frame.Parse.
	//
	// Samples is used by decodeFixed and decodeFIR to temporarily store
	// residuals. Before returning they call decodeLPC which decodes the audio
	// samples.
	Samples []int32
	// Number of audio samples in the subframe.
	NSamples int
}

// parseSubframe reads and parses the header, and the audio samples of a
// subframe.
func (frame *Frame) parseSubframe(br *bits.Reader, bps uint) (subframe *Subframe, err error) {
	// Parse subframe header.
	subframe = new(Subframe)
	if err = subframe.parseHeader(br); err != nil {
		return subframe, err
	}
	// Adjust bps of subframe for wasted bits-per-sample.
	bps -= subframe.Wasted

	// Decode subframe audio samples.
	subframe.NSamples = int(frame.BlockSize)
	subframe.Samples = make([]int32, 0, subframe.NSamples)
	switch subframe.Pred {
	case PredConstant:
		err = subframe.decodeConstant(br, bps)
	case PredVerbatim:
		err = subframe.decodeVerbatim(br, bps)
	case PredFixed:
		err = subframe.decodeFixed(br, bps)
	case PredFIR:
		err = subframe.decodeFIR(br, bps)
	}

	// Left shift to accout for wasted bits-per-sample.
	for i, sample := range subframe.Samples {
		subframe.Samples[i] = sample << subframe.Wasted
	}
	return subframe, err
}

// A SubHeader specifies the prediction method and order of a subframe.
//
// ref: https://www.xiph.org/flac/format.html#subframe_header
type SubHeader struct {
	// Specifies the prediction method used to encode the audio sample of the
	// subframe.
	Pred Pred
	// Prediction order used by fixed and FIR linear prediction decoding.
	Order int
	// Wasted bits-per-sample.
	Wasted uint
}

// parseHeader reads and parses the header of a subframe.
func (subframe *Subframe) parseHeader(br *bits.Reader) error {
	// 1 bit: zero-padding.
	x, err := br.Read(1)
	if err != nil {
		return unexpected(err)
	}
	if x != 0 {
		return errors.New("frame.Subframe.parseHeader: non-zero padding")
	}

	// 6 bits: Pred.
	x, err = br.Read(6)
	if err != nil {
		return unexpected(err)
	}
	// The 6 bits are used to specify the prediction method and order as follows:
	//    000000: Constant prediction method.
	//    000001: Verbatim prediction method.
	//    00001x: reserved.
	//    0001xx: reserved.
	//    001xxx:
	//       if (xxx <= 4)
	//          Fixed prediction method; xxx=order
	//       else
	//          reserved.
	//    01xxxx: reserved.
	//    1xxxxx: FIR prediction method; xxxxx=order-1
	switch {
	case x < 1:
		// 000000: Constant prediction method.
		subframe.Pred = PredConstant
	case x < 2:
		// 000001: Verbatim prediction method.
		subframe.Pred = PredVerbatim
	case x < 8:
		// 00001x: reserved.
		// 0001xx: reserved.
		return fmt.Errorf("frame.Subframe.parseHeader: reserved prediction method bit pattern (%06b)", x)
	case x < 16:
		// 001xxx:
		//    if (xxx <= 4)
		//       Fixed prediction method; xxx=order
		//    else
		//       reserved.
		order := int(x & 0x07)
		if order > 4 {
			return fmt.Errorf("frame.Subframe.parseHeader: reserved prediction method bit pattern (%06b)", x)
		}
		subframe.Pred = PredFixed
		subframe.Order = order
	case x < 32:
		// 01xxxx: reserved.
		return fmt.Errorf("frame.Subframe.parseHeader: reserved prediction method bit pattern (%06b)", x)
	default:
		// 1xxxxx: FIR prediction method; xxxxx=order-1
		subframe.Pred = PredFIR
		subframe.Order = int(x&0x1F) + 1
	}

	// 1 bit: hasWastedBits.
	x, err = br.Read(1)
	if err != nil {
		return unexpected(err)
	}
	if x != 0 {
		// k wasted bits-per-sample in source subblock, k-1 follows, unary coded;
		// e.g. k=3 => 001 follows, k=7 => 0000001 follows.
		x, err = br.ReadUnary()
		if err != nil {
			return unexpected(err)
		}
		subframe.Wasted = uint(x) + 1
	}

	return nil
}

// Pred specifies the prediction method used to encode the audio samples of a
// subframe.
type Pred uint8

// Prediction methods.
const (
	// PredConstant specifies that the subframe contains a constant sound. The
	// audio samples are encoded using run-length encoding. Since every audio
	// sample has the same constant value, a single unencoded audio sample is
	// stored in practice. It is replicated a number of times, as specified by
	// BlockSize in the frame header.
	PredConstant Pred = iota
	// PredVerbatim specifies that the subframe contains unencoded audio samples.
	// Random sound is often stored verbatim, since no prediction method can
	// compress it sufficiently.
	PredVerbatim
	// PredFixed specifies that the subframe contains linear prediction coded
	// audio samples. The coefficients of the prediction polynomial are selected
	// from a fixed set, and can represent 0th through fourth-order polynomials.
	// The prediction order (0 through 4) is stored within the subframe along
	// with the same number of unencoded warm-up samples, which are used to kick
	// start the prediction polynomial. The remainder of the subframe stores
	// encoded residuals (signal errors) which specify the difference between the
	// predicted and the original audio samples.
	PredFixed
	// PredFIR specifies that the subframe contains linear prediction coded audio
	// samples. The coefficients of the prediction polynomial are stored in the
	// subframe, and can represent 0th through 32nd-order polynomials. The
	// prediction order (0 through 32) is stored within the subframe along with
	// the same number of unencoded warm-up samples, which are used to kick start
	// the prediction polynomial. The remainder of the subframe stores encoded
	// residuals (signal errors) which specify the difference between the
	// predicted and the original audio samples.
	PredFIR
)

// signExtend interprets x as a signed n-bit integer value and sign extends it
// to 32 bits.
func signExtend(x uint64, n uint) int32 {
	// x is signed if its most significant bit is set.
	if x&(1<<(n-1)) != 0 {
		// Sign extend x.
		return int32(x | ^uint64(0)<<n)
	}
	return int32(x)
}

// decodeConstant reads an unencoded audio sample of the subframe. Each sample
// of the subframe has this constant value. The constant encoding can be thought
// of as run-length encoding.
//
// ref: https://www.xiph.org/flac/format.html#subframe_constant
func (subframe *Subframe) decodeConstant(br *bits.Reader, bps uint) error {
	// (bits-per-sample) bits: Unencoded constant value of the subblock.
	x, err := br.Read(bps)
	if err != nil {
		return unexpected(err)
	}

	// Each sample of the subframe has the same constant value.
	sample := signExtend(x, bps)
	for i := 0; i < subframe.NSamples; i++ {
		subframe.Samples = append(subframe.Samples, sample)
	}

	return nil
}

// decodeVerbatim reads the unencoded audio samples of the subframe.
//
// ref: https://www.xiph.org/flac/format.html#subframe_verbatim
func (subframe *Subframe) decodeVerbatim(br *bits.Reader, bps uint) error {
	// Parse the unencoded audio samples of the subframe.
	for i := 0; i < subframe.NSamples; i++ {
		// (bits-per-sample) bits: Unencoded constant value of the subblock.
		x, err := br.Read(bps)
		if err != nil {
			return unexpected(err)
		}
		sample := signExtend(x, bps)
		subframe.Samples = append(subframe.Samples, sample)
	}
	return nil
}

// fixedCoeffs maps from prediction order to the LPC coefficients used in fixed
// encoding.
//
//    x_0[n] = 0
//    x_1[n] = x[n-1]
//    x_2[n] = 2*x[n-1] - x[n-2]
//    x_3[n] = 3*x[n-1] - 3*x[n-2] + x[n-3]
//    x_4[n] = 4*x[n-1] - 6*x[n-2] + 4*x[n-3] - x[n-4]
var fixedCoeffs = [...][]int32{
	// ref: Section 2.2 of http://www.hpl.hp.com/techreports/1999/HPL-1999-144.pdf
	1: {1},
	2: {2, -1},
	3: {3, -3, 1},
	// ref: Data Compression: The Complete Reference (7.10.1)
	4: {4, -6, 4, -1},
}

// decodeFixed decodes the linear prediction coded samples of the subframe,
// using a fixed set of predefined polynomial coefficients.
//
// ref: https://www.xiph.org/flac/format.html#subframe_fixed
func (subframe *Subframe) decodeFixed(br *bits.Reader, bps uint) error {
	// Parse unencoded warm-up samples.
	for i := 0; i < subframe.Order; i++ {
		// (bits-per-sample) bits: Unencoded warm-up sample.
		x, err := br.Read(bps)
		if err != nil {
			return unexpected(err)
		}
		sample := signExtend(x, bps)
		subframe.Samples = append(subframe.Samples, sample)
	}

	// Decode subframe residuals.
	err := subframe.decodeResidual(br)
	if err != nil {
		return err
	}

	// Predict the audio samples of the subframe using a polynomial with
	// predefined coefficients of a given order. Correct signal errors using the
	// decoded residuals.
	return subframe.decodeLPC(fixedCoeffs[subframe.Order], 0)
}

// decodeFIR decodes the linear prediction coded samples of the subframe, using
// polynomial coefficients stored in the stream.
//
// ref: https://www.xiph.org/flac/format.html#subframe_lpc
func (subframe *Subframe) decodeFIR(br *bits.Reader, bps uint) error {
	// Parse unencoded warm-up samples.
	for i := 0; i < subframe.Order; i++ {
		// (bits-per-sample) bits: Unencoded warm-up sample.
		x, err := br.Read(bps)
		if err != nil {
			return unexpected(err)
		}
		sample := signExtend(x, bps)
		subframe.Samples = append(subframe.Samples, sample)
	}

	// 4 bits: (coefficients' precision in bits) - 1.
	x, err := br.Read(4)
	if err != nil {
		return unexpected(err)
	}
	if x == 0xF {
		return errors.New("frame.Subframe.decodeFIR: invalid coefficient precision bit pattern (1111)")
	}
	prec := uint(x) + 1

	// 5 bits: predictor coefficient shift needed in bits.
	x, err = br.Read(5)
	if err != nil {
		return unexpected(err)
	}
	shift := signExtend(x, 5)

	// Parse coefficients.
	coeffs := make([]int32, subframe.Order)
	for i := range coeffs {
		// (prec) bits: Predictor coefficient.
		x, err = br.Read(prec)
		if err != nil {
			return unexpected(err)
		}
		coeffs[i] = signExtend(x, prec)
	}

	// Decode subframe residuals.
	if err = subframe.decodeResidual(br); err != nil {
		return err
	}

	// Predict the audio samples of the subframe using a polynomial with
	// predefined coefficients of a given order. Correct signal errors using the
	// decoded residuals.
	return subframe.decodeLPC(coeffs, shift)
}

// decodeResidual decodes the encoded residuals (prediction method error
// signals) of the subframe.
//
// ref: https://www.xiph.org/flac/format.html#residual
func (subframe *Subframe) decodeResidual(br *bits.Reader) error {
	// 2 bits: Residual coding method.
	x, err := br.Read(2)
	if err != nil {
		return unexpected(err)
	}
	// The 2 bits are used to specify the residual coding method as follows:
	//    00: Rice coding with a 4-bit Rice parameter.
	//    01: Rice coding with a 5-bit Rice parameter.
	//    10: reserved.
	//    11: reserved.
	switch x {
	case 0x0:
		return subframe.decodeRicePart(br, 4)
	case 0x1:
		return subframe.decodeRicePart(br, 5)
	default:
		return fmt.Errorf("frame.Subframe.decodeResidual: reserved residual coding method bit pattern (%02b)", x)
	}
}

// decodeRicePart decodes a Rice partition of encoded residuals from the
// subframe, using a Rice parameter of the specified size in bits.
//
// ref: https://www.xiph.org/flac/format.html#partitioned_rice
// ref: https://www.xiph.org/flac/format.html#partitioned_rice2
func (subframe *Subframe) decodeRicePart(br *bits.Reader, paramSize uint) error {
	// 4 bits: Partition order.
	x, err := br.Read(4)
	if err != nil {
		return unexpected(err)
	}
	partOrder := x

	// Parse Rice partitions; in total 2^partOrder partitions.
	//
	// ref: https://www.xiph.org/flac/format.html#rice_partition
	// ref: https://www.xiph.org/flac/format.html#rice2_partition
	nparts := 1 << partOrder
	for i := 0; i < nparts; i++ {
		// (4 or 5) bits: Rice parameter.
		x, err = br.Read(paramSize)
		if err != nil {
			return unexpected(err)
		}
		param := uint(x)

		// Determine the number of Rice encoded samples in the partition.
		var nsamples int
		if partOrder == 0 {
			nsamples = subframe.NSamples - subframe.Order
		} else if i != 0 {
			nsamples = subframe.NSamples / nparts
		} else {
			nsamples = subframe.NSamples/nparts - subframe.Order
		}

		// TODO(u): Verify that decoding of subframes with Rice parameter escape
		// codes have been implemented correctly.
		if paramSize == 4 && param == 0xF || paramSize == 5 && param == 0x1F {
			// 1111 or 11111: Escape code, meaning the partition is in unencoded
			// binary form using n bits per sample; n follows as a 5-bit number.
			x, err := br.Read(5)
			if err != nil {
				return unexpected(err)
			}
			n := uint(x)
			for j := 0; j < nsamples; j++ {
				sample, err := br.Read(n)
				if err != nil {
					return unexpected(err)
				}
				subframe.Samples = append(subframe.Samples, int32(sample))
			}
			// TODO(u): Remove log message when the test cases have been extended.
			log.Print("frame.Subframe.decodeRicePart: The flac library test cases do not yet include any audio files with Rice parameter escape codes. If possible please consider contributing this audio sample to improve the reliability of the test cases.")
			return nil
		}

		// Decode the Rice encoded residuals of the partition.
		for j := 0; j < nsamples; j++ {
			if err = subframe.decodeRiceResidual(br, param); err != nil {
				return err
			}
		}
	}

	return nil
}

// decodeRiceResidual decodes a Rice encoded residual (error signal).
func (subframe *Subframe) decodeRiceResidual(br *bits.Reader, k uint) error {
	// Read unary encoded most significant bits.
	high, err := br.ReadUnary()
	if err != nil {
		return unexpected(err)
	}

	// Read binary encoded least significant bits.
	low, err := br.Read(k)
	if err != nil {
		return unexpected(err)
	}
	residual := int32(high<<k | low)

	// ZigZag decode.
	residual = bits.ZigZag(residual)
	subframe.Samples = append(subframe.Samples, residual)

	return nil
}

// decodeLPC decodes linear prediction coded audio samples, using the
// coefficients of a given polynomial, a couple of unencoded warm-up samples,
// and the signal errors of the prediction as specified by the residuals.
func (subframe *Subframe) decodeLPC(coeffs []int32, shift int32) error {
	if len(coeffs) != subframe.Order {
		return fmt.Errorf("frame.Subframe.decodeLPC: prediction order (%d) differs from number of coefficients (%d)", subframe.Order, len(coeffs))
	}
	if shift < 0 {
		return fmt.Errorf("frame.Subframe.decodeLPC: invalid negative shift")
	}
	if subframe.NSamples != len(subframe.Samples) {
		return fmt.Errorf("frame.Subframe.decodeLPC: subframe sample count mismatch; expected %d, got %d", subframe.NSamples, len(subframe.Samples))
	}
	for i := subframe.Order; i < subframe.NSamples; i++ {
		var sample int64
		for j, c := range coeffs {
			sample += int64(c) * int64(subframe.Samples[i-j-1])
		}
		subframe.Samples[i] += int32(sample >> uint(shift))
	}
	return nil
}
