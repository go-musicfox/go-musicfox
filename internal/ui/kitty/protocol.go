package kitty

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"strings"
	"sync/atomic"
)

const (
	// APC (Application Program Command) escape sequence start
	apcStart = "\x1b_G"
	// ST (String Terminator)
	st = "\x1b\\"

	// Chunk size for base64 encoded data (4096 bytes max per chunk)
	maxChunkSize = 4096
)

var imageIDCounter uint32

// NewImageID returns the next unique image ID.
func NewImageID() uint32 {
	return nextImageID()
}

// nextImageID returns the next unique image ID.
func nextImageID() uint32 {
	return atomic.AddUint32(&imageIDCounter, 1)
}

// TransmitAndDisplay transmits an image and displays it at the current cursor position.
// Returns the escape sequence string that should be written to the terminal.
func TransmitAndDisplay(img image.Image, cols, rows int) (string, error) {
	return TransmitAndDisplayWithID(img, cols, rows, nextImageID())
}

// TransmitAndDisplayWithID transmits an image with a specific ID.
// Useful for updating an existing image (animation).
func TransmitAndDisplayWithID(img image.Image, cols, rows int, imageID uint32) (string, error) {
	return transmit(img, cols, rows, imageID, "T", 0)
}

// TransmitImage transmits an image without displaying it.
// Returns the escape sequence string.
func TransmitImage(img image.Image, cols, rows int, imageID uint32) (string, error) {
	return transmit(img, cols, rows, imageID, "t", 0)
}

// DeleteImage deletes an image by its ID.
func DeleteImage(imageID uint32) string {
	// a=d: delete action
	// d=i: delete by image ID
	// i=<id>: image ID to delete
	return fmt.Sprintf("%sa=d,d=i,i=%d,q=2%s", apcStart, imageID, st)
}

// DeleteAllImages deletes all images from the terminal.
func DeleteAllImages() string {
	// a=d: delete action
	// d=a: delete all images
	return fmt.Sprintf("%sa=d,d=a,q=2%s", apcStart, st)
}

// ClearImageArea clears the image display area by moving cursor and clearing.
// This is useful for cleaning up before displaying a new image.
func ClearImageArea(rows int) string {
	var result strings.Builder
	// Move up and clear each row
	for i := range rows {
		result.WriteString("\x1b[2K") // Clear entire line
		if i < rows-1 {
			result.WriteString("\x1b[1B") // Move down one line
		}
	}
	// Move back up to starting position
	if rows > 1 {
		fmt.Fprintf(&result, "\x1b[%dA", rows-1)
	}
	return result.String()
}

// Placeholder creates a placeholder string of the specified size.
// This is used to reserve space for the image in the terminal output.
func Placeholder(cols, rows int) string {
	var result strings.Builder
	space := strings.Repeat(" ", cols)
	for i := range rows {
		result.WriteString(space)
		if i < rows-1 {
			result.WriteString("\n")
		}
	}
	return result.String()
}

// splitIntoChunks splits a string into chunks of the specified size.
// transmit handles the common logic for transmitting image data.
func transmit(img image.Image, cols, rows int, imageID uint32, action string, gapMS int) (string, error) {
	// Encode image to PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", fmt.Errorf("failed to encode image to PNG: %w", err)
	}

	// Base64 encode the PNG data
	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())

	// Build the kitty graphics command
	var result strings.Builder

	// Parameters builder
	var params strings.Builder
	fmt.Fprintf(&params, "a=%s,f=100", action)

	// For transmission (t or T), we generally use 'd' (direct) if not specified, but the protocol default is 'd'.
	// Adding t=d explicitly is fine.
	if action == "T" || action == "t" || action == "f" {
		params.WriteString(",t=d")
	}

	fmt.Fprintf(&params, ",i=%d", imageID)

	if cols > 0 {
		fmt.Fprintf(&params, ",c=%d", cols)
	}

	if gapMS > 0 {
		fmt.Fprintf(&params, ",z=%d", gapMS)
	}

	params.WriteString(",q=2") // Suppress responses

	// For small images, we can send in one chunk
	if len(encoded) <= maxChunkSize {
		fmt.Fprintf(&result, "%s%s;%s%s", apcStart, params.String(), encoded, st)
	} else {
		// Multi-chunk transmission
		chunks := splitIntoChunks(encoded, maxChunkSize)
		for i, chunk := range chunks {
			if i == 0 {
				// First chunk: m=1 means more data follows
				fmt.Fprintf(&result, "%s%s,m=1;%s%s", apcStart, params.String(), chunk, st)
			} else if i == len(chunks)-1 {
				// Last chunk: m=0 means no more data
				if action == "f" {
					fmt.Fprintf(&result, "%sa=f,m=0;%s%s", apcStart, chunk, st)
				} else {
					fmt.Fprintf(&result, "%sm=0;%s%s", apcStart, chunk, st)
				}
			} else {
				// Middle chunks: m=1
				if action == "f" {
					fmt.Fprintf(&result, "%sa=f,m=1;%s%s", apcStart, chunk, st)
				} else {
					fmt.Fprintf(&result, "%sm=1;%s%s", apcStart, chunk, st)
				}
			}
		}
	}

	return result.String(), nil
}

// TransmitFrame transmits a frame for an existing animation.
// imageID is the ID of the base image.
// gapMS is the duration of this frame in milliseconds.
func TransmitFrame(img image.Image, imageID uint32, gapMS int) (string, error) {
	return transmit(img, 0, 0, imageID, "f", gapMS) // a=f for frame data
}

// StartAnimation starts the animation playback.
func StartAnimation(imageID uint32) string {
	// a=a: animation control
	// s=3: start (loop mode)
	return fmt.Sprintf("%sa=a,i=%d,s=3,q=2%s", apcStart, imageID, st)
}

// StopAnimation stops the animation playback.
func StopAnimation(imageID uint32) string {
	// a=a: animation control
	// s=1: stop
	return fmt.Sprintf("%sa=a,i=%d,s=1,q=2%s", apcStart, imageID, st)
}

// SetFrameGap sets the gap (duration) for a specific frame.
// frameIdx: 1-based frame index (1 is root/base frame).
func SetFrameGap(imageID uint32, frameIdx int, gapMS int) string {
	// a=a: animation control
	// r=<frame>: target frame
	// z=<gap>: duration in ms
	return fmt.Sprintf("%sa=a,i=%d,r=%d,z=%d,q=2%s", apcStart, imageID, frameIdx, gapMS, st)
}

// PlaceImage generates a command to display (place) an already transmitted image at the current cursor position.
func PlaceImage(imageID uint32, cols, rows int) string {
	// a=p: placement action
	// i=<id>: image ID to place
	// c=<cols>: columns
	// r=<rows>: rows
	// C=1: do not move cursor
	// If rows is 0, let Kitty calculate height based on image aspect ratio (square) and cols
	if rows == 0 {
		return fmt.Sprintf("%sa=p,i=%d,c=%d,C=1,q=2%s", apcStart, imageID, cols, st)
	}
	return fmt.Sprintf("%sa=p,i=%d,c=%d,r=%d,C=1,q=2%s", apcStart, imageID, cols, rows, st)
}

// splitIntoChunks splits a string into chunks of the specified size.
func splitIntoChunks(s string, chunkSize int) []string {
	var chunks []string
	for len(s) > 0 {
		if len(s) > chunkSize {
			chunks = append(chunks, s[:chunkSize])
			s = s[chunkSize:]
		} else {
			chunks = append(chunks, s)
			break
		}
	}
	return chunks
}
