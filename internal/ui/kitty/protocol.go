package kitty

import (
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"strings"
	"sync/atomic"

	"bytes"
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

// nextImageID returns the next unique image ID.
func nextImageID() uint32 {
	return atomic.AddUint32(&imageIDCounter, 1)
}

// TransmitAndDisplay transmits an image and displays it at the current cursor position.
// Returns the escape sequence string that should be written to the terminal.
// cols specifies the display width in terminal cells.
// rows is kept for interface compatibility but not used - Kitty calculates rows
// automatically based on the image's aspect ratio.
// For a visually square display, send a 2:1 stretched image (width = height * 2).
func TransmitAndDisplay(img image.Image, cols, rows int) (string, error) {
	// Encode image to PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", fmt.Errorf("failed to encode image to PNG: %w", err)
	}

	// Base64 encode the PNG data
	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())

	// Generate unique image ID
	imageID := nextImageID()

	// Build the kitty graphics command
	var result strings.Builder

	// For small images, we can send in one chunk
	if len(encoded) <= maxChunkSize {
		// a=T: transmit and display
		// f=100: PNG format
		// t=d: direct data transmission
		// i=<id>: image ID
		// c=<cols>: display columns (rows calculated automatically from image aspect ratio)
		// q=2: suppress responses
		result.WriteString(fmt.Sprintf("%sa=T,f=100,t=d,i=%d,c=%d,q=2;%s%s",
			apcStart, imageID, cols, encoded, st))
	} else {
		// Multi-chunk transmission
		chunks := splitIntoChunks(encoded, maxChunkSize)
		for i, chunk := range chunks {
			if i == 0 {
				// First chunk: m=1 means more data follows
				result.WriteString(fmt.Sprintf("%sa=T,f=100,t=d,i=%d,c=%d,m=1,q=2;%s%s",
					apcStart, imageID, cols, chunk, st))
			} else if i == len(chunks)-1 {
				// Last chunk: m=0 means no more data
				result.WriteString(fmt.Sprintf("%sm=0;%s%s",
					apcStart, chunk, st))
			} else {
				// Middle chunks: m=1 means more data follows
				result.WriteString(fmt.Sprintf("%sm=1;%s%s",
					apcStart, chunk, st))
			}
		}
	}

	return result.String(), nil
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
	for i := 0; i < rows; i++ {
		result.WriteString("\x1b[2K") // Clear entire line
		if i < rows-1 {
			result.WriteString("\x1b[1B") // Move down one line
		}
	}
	// Move back up to starting position
	if rows > 1 {
		result.WriteString(fmt.Sprintf("\x1b[%dA", rows-1))
	}
	return result.String()
}

// Placeholder creates a placeholder string of the specified size.
// This is used to reserve space for the image in the terminal output.
func Placeholder(cols, rows int) string {
	var result strings.Builder
	space := strings.Repeat(" ", cols)
	for i := 0; i < rows; i++ {
		result.WriteString(space)
		if i < rows-1 {
			result.WriteString("\n")
		}
	}
	return result.String()
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
