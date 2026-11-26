package kitty

import (
	"image"
	"image/color"
	"math/rand/v2"
	"strings"
	"testing"
)

func setTmuxPassthrough(t *testing.T, enabled bool) {
	t.Helper()
	orig := tmuxPassthrough
	tmuxPassthrough = enabled
	t.Cleanup(func() { tmuxPassthrough = orig })
}

func TestDeleteAllImagesWithoutPassthrough(t *testing.T) {
	setTmuxPassthrough(t, false)

	seq := DeleteAllImages()
	if !strings.HasPrefix(seq, "\x1b_G") {
		t.Errorf("expected sequence to start with APC start marker, got %q", seq)
	}
	if !strings.HasSuffix(seq, "\x1b\\") {
		t.Errorf("expected sequence to end with ST, got %q", seq)
	}
	if strings.Contains(seq, "tmux;") {
		t.Errorf("expected no tmux passthrough wrapper, got %q", seq)
	}
}

func TestGeneratorsReturnBareSequencesEvenWithPassthrough(t *testing.T) {
	setTmuxPassthrough(t, true)

	// Generators must always return bare APC sequences: tmux passthrough
	// wrapping is applied once at the write site. A pre-wrapped generator
	// output would get double-wrapped when the write site wraps a composite
	// payload.
	for name, seq := range map[string]string{
		"DeleteAllImages": DeleteAllImages(),
		"DeleteImage":     DeleteImage(1),
		"PlaceImage":      PlaceImage(1, 2, 0, 1),
		"StartAnimation":  StartAnimation(1),
		"StopAnimation":   StopAnimation(1),
		"SetFrameGap":     SetFrameGap(1, 1, 40),
	} {
		if !strings.HasPrefix(seq, "\x1b_G") {
			t.Errorf("%s: expected bare APC sequence, got %q", name, seq)
		}
	}
}

func TestDeleteAllImagesWithPassthrough(t *testing.T) {
	setTmuxPassthrough(t, true)

	seq := Wrap(DeleteAllImages())
	if !strings.HasPrefix(seq, "\x1bPtmux;") {
		t.Errorf("expected sequence to start with DCS tmux; passthrough, got %q", seq)
	}
	if !strings.HasSuffix(seq, "\x1b\\") {
		t.Errorf("expected sequence to end with ST, got %q", seq)
	}

	// All inner ESC bytes must be doubled: after removing doubled ESC pairs
	// from the payload, no single ESC may remain.
	payload := seq[len("\x1bPtmux;") : len(seq)-len("\x1b\\")]
	if remaining := strings.ReplaceAll(payload, "\x1b\x1b", ""); strings.Contains(remaining, "\x1b") {
		t.Errorf("expected all inner ESC bytes to be doubled, got %q", seq)
	}
}

func TestWrapWithoutPassthrough(t *testing.T) {
	setTmuxPassthrough(t, false)

	raw := DeleteAllImages()
	if got := Wrap(raw); got != raw {
		t.Errorf("expected Wrap to return the sequence unchanged, got %q", got)
	}
}

func TestTransmitLargeImageWithPassthrough(t *testing.T) {
	setTmuxPassthrough(t, true)

	// Build a large noisy image so the PNG data (and thus base64) exceeds the
	// 4096-byte chunk size and triggers multi-chunk transmission.
	const size = 200
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			img.Set(x, y, color.RGBA{
				R: uint8(rand.IntN(256)),
				G: uint8(rand.IntN(256)),
				B: uint8(rand.IntN(256)),
				A: 255,
			})
		}
	}

	raw, err := TransmitAndDisplay(img, 10, 5)
	if err != nil {
		t.Fatalf("TransmitAndDisplay failed: %v", err)
	}
	if !strings.Contains(raw, "m=1;") {
		t.Errorf("expected multi-chunk transmission (m=1 markers), got single chunk of len %d", len(raw))
	}

	// The generator returns a bare sequence; the write site wraps the whole
	// multi-chunk transmission as one DCS passthrough.
	seq := Wrap(raw)
	if !strings.HasPrefix(seq, "\x1bPtmux;") {
		t.Errorf("expected multi-chunk sequence to start with DCS tmux; passthrough, got prefix %q", seq[:20])
	}
	if !strings.HasSuffix(seq, "\x1b\\") {
		t.Errorf("expected sequence to end with ST, got %q", seq[len(seq)-10:])
	}

	// All inner ESC bytes must be doubled.
	payload := seq[len("\x1bPtmux;") : len(seq)-len("\x1b\\")]
	if remaining := strings.ReplaceAll(payload, "\x1b\x1b", ""); strings.Contains(remaining, "\x1b") {
		t.Errorf("expected all inner ESC bytes to be doubled in multi-chunk sequence")
	}
}
