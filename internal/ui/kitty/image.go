package kitty

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/go-musicfox/go-musicfox/internal/configs"
)

const (
	// DefaultCacheSize is the default number of images to cache
	DefaultCacheSize = 10
	// DefaultHTTPTimeout is the default timeout for HTTP requests
	DefaultHTTPTimeout = 10 * time.Second
)

// ImageCache manages downloaded and processed cover images.
type ImageCache struct {
	mu         sync.RWMutex
	cache      map[string]*cacheEntry
	order      []string // LRU order: oldest at front, newest at back
	maxSize    int
	httpClient *http.Client
}

type cacheEntry struct {
	img      image.Image
	cols     int
	rows     int
	kittySeq string // Pre-computed kitty sequence
}

// NewImageCache creates a new image cache with the specified maximum size.
func NewImageCache(maxSize int) *ImageCache {
	if maxSize <= 0 {
		maxSize = DefaultCacheSize
	}
	return &ImageCache{
		cache:   make(map[string]*cacheEntry),
		order:   make([]string, 0, maxSize),
		maxSize: maxSize,
		httpClient: &http.Client{
			Timeout: DefaultHTTPTimeout,
		},
	}
}

// GetOrFetch retrieves an image from cache or fetches it from the URL.
// cols and rows specify the desired display size in terminal cells.
// Returns the kitty escape sequence ready to be written to the terminal.
func (c *ImageCache) GetOrFetch(ctx context.Context, url string, cols, rows int) (string, error) {
	entry, err := c.getOrFetchEntry(ctx, url, cols, rows)
	if err != nil {
		return "", err
	}
	return entry.kittySeq, nil
}

// GetImage retrieves an image from cache or fetches it from the URL.
// It returns the processed image.Image (resized and rounded).
func (c *ImageCache) GetImage(ctx context.Context, url string, cols, rows int) (image.Image, error) {
	entry, err := c.getOrFetchEntry(ctx, url, cols, rows)
	if err != nil {
		return nil, err
	}
	return entry.img, nil
}

func (c *ImageCache) getOrFetchEntry(ctx context.Context, url string, cols, rows int) (*cacheEntry, error) {
	if url == "" {
		return nil, nil // Or error? GetOrFetch used to return "", nil
	}

	cacheKey := fmt.Sprintf("%s_%d_%d", url, cols, rows)

	// Check cache first
	c.mu.RLock()
	if entry, ok := c.cache[cacheKey]; ok {
		c.mu.RUnlock()
		// Move to end of LRU list (most recently used)
		c.touch(cacheKey)
		return entry, nil
	}
	c.mu.RUnlock()

	// Fetch and process the image
	img, err := c.fetchImage(ctx, url)
	if err != nil {
		return nil, err
	}

	// Resize image for terminal display.
	// Use 1:1 ratio (no stretch) - the image will be displayed as a square.
	// 320px is usually enough for terminal covers (approx 30-40 columns)
	// and much faster to process/encode for animations than 512px.
	targetSize := 512
	if configs.AppConfig.Main.Lyric.Cover.Spin {
		targetSize = 320
	}
	resized := resizeImageSquare(img, targetSize)

	// Apply rounded corners based on config
	radiusPercent := float64(configs.AppConfig.Main.Lyric.Cover.CornerRadius) / 100.0 / 2.0
	if radiusPercent >= 0.5 {
		radiusPercent = 0.5
	}
	rounded := applyRoundedCorners(resized, radiusPercent)

	// Generate kitty sequence
	kittySeq, err := TransmitAndDisplay(rounded, cols, rows)
	if err != nil {
		return nil, fmt.Errorf("failed to generate kitty sequence: %w", err)
	}

	// Store in cache
	c.mu.Lock()
	entry := &cacheEntry{
		img:      rounded,
		cols:     cols,
		rows:     rows,
		kittySeq: kittySeq,
	}
	c.cache[cacheKey] = entry
	c.order = append(c.order, cacheKey)

	// Evict oldest entries if cache is full
	for len(c.cache) > c.maxSize && len(c.order) > 0 {
		oldest := c.order[0]
		c.order = c.order[1:]
		delete(c.cache, oldest)
	}
	c.mu.Unlock()

	return entry, nil
}

// touch moves a cache key to the end of the LRU list.
func (c *ImageCache) touch(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i, k := range c.order {
		if k == key {
			c.order = append(c.order[:i], c.order[i+1:]...)
			c.order = append(c.order, key)
			break
		}
	}
}

// fetchImage downloads an image from the given URL.
func (c *ImageCache) fetchImage(ctx context.Context, url string) (image.Image, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Limit the size we'll read to prevent memory issues (10MB max)
	limited := io.LimitReader(resp.Body, 10*1024*1024)

	img, _, err := image.Decode(limited)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	return img, nil
}

// Clear removes all entries from the cache.
func (c *ImageCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[string]*cacheEntry)
	c.order = c.order[:0]
}

// resizeImageSquare resizes an image to a square of the given size.
// The image is center-cropped to maintain 1:1 aspect ratio before scaling.
func resizeImageSquare(src image.Image, targetSize int) image.Image {
	srcBounds := src.Bounds()
	srcWidth := srcBounds.Dx()
	srcHeight := srcBounds.Dy()

	// Determine the crop region to make it square (center crop)
	var cropX, cropY, cropSize int
	if srcWidth > srcHeight {
		cropSize = srcHeight
		cropX = (srcWidth - srcHeight) / 2
		cropY = 0
	} else {
		cropSize = srcWidth
		cropX = 0
		cropY = (srcHeight - srcWidth) / 2
	}

	// Create a square cropped image
	cropped := image.NewRGBA(image.Rect(0, 0, cropSize, cropSize))
	for y := 0; y < cropSize; y++ {
		for x := 0; x < cropSize; x++ {
			cropped.Set(x, y, src.At(srcBounds.Min.X+cropX+x, srcBounds.Min.Y+cropY+y))
		}
	}

	// Resize the cropped square to target size
	if targetSize <= 0 {
		targetSize = 1
	}

	dst := image.NewRGBA(image.Rect(0, 0, targetSize, targetSize))
	bilinearScale(dst, cropped)

	return dst
}

// bilinearScale performs bilinear interpolation scaling.
func bilinearScale(dst *image.RGBA, src image.Image) {
	dstBounds := dst.Bounds()
	srcBounds := src.Bounds()

	dstW := dstBounds.Dx()
	dstH := dstBounds.Dy()
	srcW := srcBounds.Dx()
	srcH := srcBounds.Dy()

	for y := 0; y < dstH; y++ {
		for x := 0; x < dstW; x++ {
			// Map destination coordinates to source coordinates
			srcX := float64(x) * float64(srcW-1) / float64(dstW-1)
			srcY := float64(y) * float64(srcH-1) / float64(dstH-1)

			// Handle edge case when dstW or dstH is 1
			if dstW == 1 {
				srcX = float64(srcW-1) / 2
			}
			if dstH == 1 {
				srcY = float64(srcH-1) / 2
			}

			// Get the four nearest pixels
			x0 := int(srcX)
			y0 := int(srcY)
			x1 := x0 + 1
			y1 := y0 + 1

			// Clamp to source bounds
			if x1 >= srcW {
				x1 = srcW - 1
			}
			if y1 >= srcH {
				y1 = srcH - 1
			}

			// Get interpolation weights
			xWeight := srcX - float64(x0)
			yWeight := srcY - float64(y0)

			// Get the four pixel colors
			c00 := src.At(srcBounds.Min.X+x0, srcBounds.Min.Y+y0)
			c10 := src.At(srcBounds.Min.X+x1, srcBounds.Min.Y+y0)
			c01 := src.At(srcBounds.Min.X+x0, srcBounds.Min.Y+y1)
			c11 := src.At(srcBounds.Min.X+x1, srcBounds.Min.Y+y1)

			// Convert to RGBA
			r00, g00, b00, a00 := c00.RGBA()
			r10, g10, b10, a10 := c10.RGBA()
			r01, g01, b01, a01 := c01.RGBA()
			r11, g11, b11, a11 := c11.RGBA()

			// Bilinear interpolation
			r := bilinearInterp(r00, r10, r01, r11, xWeight, yWeight)
			g := bilinearInterp(g00, g10, g01, g11, xWeight, yWeight)
			b := bilinearInterp(b00, b10, b01, b11, xWeight, yWeight)
			a := bilinearInterp(a00, a10, a01, a11, xWeight, yWeight)

			// Set the pixel - convert from 16-bit to 8-bit color
			dst.SetRGBA(dstBounds.Min.X+x, dstBounds.Min.Y+y, color.RGBA{
				R: uint8(r >> 8),
				G: uint8(g >> 8),
				B: uint8(b >> 8),
				A: uint8(a >> 8),
			})
		}
	}
}

// bilinearInterp performs bilinear interpolation on a single channel.
func bilinearInterp(c00, c10, c01, c11 uint32, xWeight, yWeight float64) uint32 {
	// Interpolate in x direction
	top := float64(c00)*(1-xWeight) + float64(c10)*xWeight
	bottom := float64(c01)*(1-xWeight) + float64(c11)*xWeight
	// Interpolate in y direction
	return uint32(top*(1-yWeight) + bottom*yWeight)
}

// applyRoundedCorners applies rounded corners to an image.
// radiusPercent is the corner radius as a percentage of the image size (0.0-0.5).
// applyRoundedCorners applies rounded corners to an image with anti-aliasing.
// radiusPercent is the corner radius as a percentage of the image size (0.0-0.5).
func applyRoundedCorners(src image.Image, radiusPercent float64) image.Image {
	bounds := src.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Calculate corner radius
	minDim := width
	if height < minDim {
		minDim = height
	}
	radius := float64(minDim) * radiusPercent // Use float for precision
	if radius <= 0 {
		return src
	}

	// Create output image with alpha channel
	dst := image.NewRGBA(bounds)

	// Draw the source image first
	draw.Draw(dst, bounds, src, bounds.Min, draw.Src)

	// Scan only the corner regions to apply transparency
	// This optimization avoids calculating sqrt for the center of the image
	intRadius := int(math.Ceil(radius))

	for y := 0; y < height; y++ {
		if y >= intRadius && y < height-intRadius {
			continue
		}

		for x := 0; x < width; x++ {
			if x >= intRadius && x < width-intRadius {
				continue
			}
			factor := getCornerAlphaFactor(x, y, width, height, radius)
			if factor < 1.0 {
				r, g, b, a := dst.At(x, y).RGBA()
				newA := uint8((float64(a>>8) * factor))
				newR := uint8((float64(r>>8) * factor))
				newG := uint8((float64(g>>8) * factor))
				newB := uint8((float64(b>>8) * factor))
				dst.SetRGBA(x, y, color.RGBA{R: newR, G: newG, B: newB, A: newA})
			}
		}
	}

	return dst
}

// getCornerAlphaFactor calculates the alpha factor for a pixel at (x, y).
// based on its distance from the nearest corner arc center.
// Returns a value between 0.0 (fully transparent) and 1.0 (fully opaque).
func getCornerAlphaFactor(x, y, width, height int, radius float64) float64 {
	// Determine which corner we are checking
	var cx, cy float64 // Center of the corner circle

	if x < int(radius) {
		cx = radius
	} else if x >= width-int(radius) {
		cx = float64(width) - radius
	} else {
		return 1.0
	}

	if y < int(radius) {
		cy = radius
	} else if y >= height-int(radius) {
		cy = float64(height) - radius
	} else {
		return 1.0
	}

	dx := float64(x) + 0.5 - cx
	dy := float64(y) + 0.5 - cy
	distance := math.Sqrt(dx*dx + dy*dy)

	if distance < radius-0.5 {
		return 1.0
	} else if distance > radius+0.5 {
		return 0.0
	} else {
		return 1.0 - (distance - (radius - 0.5))
	}
}

// RotateImage rotates an image by the given angle in degrees.
func RotateImage(src image.Image, angle float64) image.Image {
	rad := angle * math.Pi / 180.0
	sin, cos := math.Sin(rad), math.Cos(rad)

	srcBounds := src.Bounds()
	srcW := srcBounds.Dx()
	srcH := srcBounds.Dy()

	// Use the original max dimension as the canvas size to keep it consistent.
	// We want the rotated image to fit WITHIN the original bounding box if we zoomed out,
	// but here we just want a large enough canvas.
	// Actually, just making the canvas larger (diagonal) makes the image visually shrink
	// because kitty fits the whole canvas into the cell area.

	// FIX: Use the SAME size as source.
	// This means we will crop corners when rotating, BUT the visual size will remain constant and maximized.
	// For a circle (rounded cover), rotating inside the square bounding box is perfect.
	// Corners are transparent anyway.
	dstW, dstH := srcW, srcH
	dst := image.NewRGBA(image.Rect(0, 0, dstW, dstH))

	centerX := float64(srcW) / 2.0
	centerY := float64(srcH) / 2.0
	dstCenterX := float64(dstW) / 2.0
	dstCenterY := float64(dstH) / 2.0

	for y := 0; y < dstH; y++ {
		for x := 0; x < dstW; x++ {
			// Inverse mapping
			dx := float64(x) - dstCenterX
			dy := float64(y) - dstCenterY

			srcX := dx*cos + dy*sin + centerX
			srcY := -dx*sin + dy*cos + centerY

			if srcX < 0 || srcX >= float64(srcW)-1 || srcY < 0 || srcY >= float64(srcH)-1 {
				continue
			}

			x0 := int(srcX)
			y0 := int(srcY)
			x1 := x0 + 1
			y1 := y0 + 1

			xWeight := srcX - float64(x0)
			yWeight := srcY - float64(y0)

			c00 := src.At(srcBounds.Min.X+x0, srcBounds.Min.Y+y0)
			c10 := src.At(srcBounds.Min.X+x1, srcBounds.Min.Y+y0)
			c01 := src.At(srcBounds.Min.X+x0, srcBounds.Min.Y+y1)
			c11 := src.At(srcBounds.Min.X+x1, srcBounds.Min.Y+y1)

			r00, g00, b00, a00 := c00.RGBA()
			r10, g10, b10, a10 := c10.RGBA()
			r01, g01, b01, a01 := c01.RGBA()
			r11, g11, b11, a11 := c11.RGBA()

			r := bilinearInterp(r00, r10, r01, r11, xWeight, yWeight)
			g := bilinearInterp(g00, g10, g01, g11, xWeight, yWeight)
			b := bilinearInterp(b00, b10, b01, b11, xWeight, yWeight)
			a := bilinearInterp(a00, a10, a01, a11, xWeight, yWeight)

			dst.SetRGBA(x, y, color.RGBA{
				R: uint8(r >> 8),
				G: uint8(g >> 8),
				B: uint8(b >> 8),
				A: uint8(a >> 8),
			})
		}
	}

	return dst
}
