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
	// DefaultRotationSize is the target size for rotation animations
	DefaultRotationSize = 320
)

// rgbaPool provides reusable RGBA image buffers for rotation frames.
// Avoids 180+ allocations per song change when spinning cover is enabled.
var rgbaPool = sync.Pool{
	New: func() interface{} {
		return image.NewRGBA(image.Rect(0, 0, DefaultRotationSize, DefaultRotationSize))
	},
}

// getPooledRGBA retrieves an RGBA image from the pool and clears it.
func getPooledRGBA() *image.RGBA {
	img := rgbaPool.Get().(*image.RGBA)
	// Clear pixel data to prevent ghost images
	clear(img.Pix)
	return img
}

// putPooledRGBA returns an RGBA image to the pool for reuse.
func putPooledRGBA(img *image.RGBA) {
	if img == nil {
		return
	}
	bounds := img.Bounds()
	if bounds.Dx() == DefaultRotationSize && bounds.Dy() == DefaultRotationSize {
		rgbaPool.Put(img)
	}
}

// RotateImagePooled rotates an image using a pooled destination buffer.
// Returns a pooled RGBA image - caller should call PutPooledRGBA when done.
func RotateImagePooled(src *image.RGBA, angle float64) *image.RGBA {
	size := src.Bounds().Dx()
	var dst *image.RGBA
	if size == DefaultRotationSize {
		dst = getPooledRGBA()
	} else {
		dst = image.NewRGBA(image.Rect(0, 0, size, size))
	}
	RotateImageFast(src, angle, dst)
	return dst
}

// PutPooledRGBA returns an RGBA image to the pool for reuse.
func PutPooledRGBA(img *image.RGBA) {
	putPooledRGBA(img)
}

// EnsureRGBA converts an image.Image to *image.RGBA if needed.
func EnsureRGBA(img image.Image) *image.RGBA {
	if rgba, ok := img.(*image.RGBA); ok {
		return rgba
	}
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)
	return rgba
}

// ImageCache manages downloaded and processed cover images.
type ImageCache struct {
	mu         sync.RWMutex
	cache      map[string]*cacheEntry
	order      []string // LRU order: oldest at front, newest at back
	maxSize    int
	httpClient *http.Client
	// Precomputed corner mask cache for performance
	cornerMaskCache map[string]*image.Alpha
	cornerMaskMu    sync.RWMutex
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
		cache:           make(map[string]*cacheEntry),
		order:           make([]string, 0, maxSize),
		maxSize:         maxSize,
		httpClient:      &http.Client{Timeout: DefaultHTTPTimeout},
		cornerMaskCache: make(map[string]*image.Alpha),
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

	// Get precomputed corner mask for this size (cached for performance)
	radiusPercent := float64(configs.AppConfig.Main.Lyric.Cover.CornerRadius) / 100.0 / 2.0
	if radiusPercent >= 0.5 {
		radiusPercent = 0.5
	}
	mask := c.getOrCreateCornerMask(targetSize, radiusPercent)

	// Apply rounded corners using precomputed mask (much faster than pixel-by-pixel)
	rounded := applyRoundedCornersFast(resized, mask)

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

// getOrCreateCornerMask returns a cached corner mask or creates a new one.
// The mask is precomputed once per (size, radiusPercent) combination.
func (c *ImageCache) getOrCreateCornerMask(size int, radiusPercent float64) *image.Alpha {
	if radiusPercent <= 0 {
		return nil
	}

	// Create cache key from size and radius percent
	cacheKey := fmt.Sprintf("%d_%.3f", size, radiusPercent)

	// Check cache first
	c.cornerMaskMu.RLock()
	if mask, ok := c.cornerMaskCache[cacheKey]; ok {
		c.cornerMaskMu.RUnlock()
		return mask
	}
	c.cornerMaskMu.RUnlock()

	// Create new corner mask (only for corners, not for full image)
	c.cornerMaskMu.Lock()
	defer c.cornerMaskMu.Unlock()

	// Double-check after acquiring write lock
	if mask, ok := c.cornerMaskCache[cacheKey]; ok {
		return mask
	}

	// Precompute corner radius
	radius := int(float64(size) * radiusPercent)
	if radius <= 0 {
		return nil
	}

	// Create a minimal alpha mask containing only the corner regions
	// This is more memory efficient than a full-size mask
	mask := &image.Alpha{
		Pix:    make([]uint8, size*size), // 1 byte per pixel for Alpha
		Stride: size,
		Rect:   image.Rect(0, 0, size, size),
	}

	// Initialize all pixels to opaque (255) - we only want to transparent corners
	for i := range mask.Pix {
		mask.Pix[i] = 255
	}

	// Precompute radius squared to avoid repeated calculation
	radiusSq := radius * radius

	// Top-left corner - set transparent pixels outside the corner arc
	for y := 0; y < radius; y++ {
		for x := 0; x < radius; x++ {
			dx := radius - x
			dy := radius - y
			offset := y*mask.Stride + x
			if dx*dx+dy*dy > radiusSq {
				// Transparent corner - set alpha to 0
				mask.Pix[offset] = 0
			}
		}
	}

	// Top-right corner
	for y := 0; y < radius; y++ {
		for x := 0; x < radius; x++ {
			dx := x
			dy := radius - y
			col := size - radius + x
			offset := y*mask.Stride + col
			if dx*dx+dy*dy > radiusSq {
				mask.Pix[offset] = 0
			}
		}
	}

	// Bottom-left corner
	for y := 0; y < radius; y++ {
		for x := 0; x < radius; x++ {
			dx := radius - x
			dy := y
			row := size - radius + y
			offset := row*mask.Stride + x
			if dx*dx+dy*dy > radiusSq {
				mask.Pix[offset] = 0
			}
		}
	}

	// Bottom-right corner
	for y := 0; y < radius; y++ {
		for x := 0; x < radius; x++ {
			dx := x
			dy := y
			col := size - radius + x
			row := size - radius + y
			offset := row*mask.Stride + col
			if dx*dx+dy*dy > radiusSq {
				mask.Pix[offset] = 0
			}
		}
	}

	c.cornerMaskCache[cacheKey] = mask
	return mask
}

// applyRoundedCornersFast applies rounded corners using a precomputed alpha mask.
func applyRoundedCornersFast(src image.Image, mask *image.Alpha) image.Image {
	if mask == nil {
		return src
	}

	bounds := src.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	dst := image.NewRGBA(bounds)
	draw.Draw(dst, bounds, src, bounds.Min, draw.Src)

	maskPix := mask.Pix
	maskStride := mask.Stride
	maskSize := mask.Bounds().Dx()
	radius := maskSize / 2

	// Top-left corner
	for y := 0; y < radius; y++ {
		for x := 0; x < radius; x++ {
			maskOffset := y*maskStride + x
			if maskPix[maskOffset] == 0 {
				dst.SetRGBA(x, y, color.RGBA{0, 0, 0, 0})
			}
		}
	}

	// Top-right corner
	for y := 0; y < radius; y++ {
		for x := 0; x < radius; x++ {
			maskOffset := y*maskStride + (radius + x)
			if maskPix[maskOffset] == 0 {
				dst.SetRGBA(width-radius+x, y, color.RGBA{0, 0, 0, 0})
			}
		}
	}

	// Bottom-left corner
	for y := 0; y < radius; y++ {
		for x := 0; x < radius; x++ {
			maskOffset := (radius+y)*maskStride + x
			if maskPix[maskOffset] == 0 {
				dst.SetRGBA(x, height-radius+y, color.RGBA{0, 0, 0, 0})
			}
		}
	}

	// Bottom-right corner
	for y := 0; y < radius; y++ {
		for x := 0; x < radius; x++ {
			maskOffset := (radius+y)*maskStride + (radius + x)
			if maskPix[maskOffset] == 0 {
				dst.SetRGBA(width-radius+x, height-radius+y, color.RGBA{0, 0, 0, 0})
			}
		}
	}

	return dst
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
// This is the legacy interface that allocates a new image each time.
func RotateImage(src image.Image, angle float64) image.Image {
	// Convert to RGBA if needed
	srcRGBA, ok := src.(*image.RGBA)
	if !ok {
		bounds := src.Bounds()
		srcRGBA = image.NewRGBA(bounds)
		draw.Draw(srcRGBA, bounds, src, bounds.Min, draw.Src)
	}

	// Use pooled rotation for better performance
	return RotateImagePooled(srcRGBA, angle)
}

// RotateImageFast rotates an RGBA image by the given angle in degrees.
// Optimized version: directly operates on pixel arrays to avoid interface overhead.
// The dst buffer must be the same size as src.
func RotateImageFast(src *image.RGBA, angle float64, dst *image.RGBA) {
	rad := angle * math.Pi / 180.0
	sin, cos := math.Sin(rad), math.Cos(rad)

	srcBounds := src.Bounds()
	srcW := srcBounds.Dx()
	srcH := srcBounds.Dy()
	dstW := dst.Bounds().Dx()
	dstH := dst.Bounds().Dy()

	// Direct access to pixel data
	srcPix := src.Pix
	srcStride := src.Stride
	dstPix := dst.Pix
	dstStride := dst.Stride

	centerX := float64(srcW) / 2.0
	centerY := float64(srcH) / 2.0
	dstCenterX := float64(dstW) / 2.0
	dstCenterY := float64(dstH) / 2.0

	srcWf := float64(srcW) - 1
	srcHf := float64(srcH) - 1

	for y := 0; y < dstH; y++ {
		dy := float64(y) - dstCenterY
		// Pre-calculate row offset
		rowOffset := y * dstStride

		for x := 0; x < dstW; x++ {
			dx := float64(x) - dstCenterX

			// Inverse mapping
			srcX := dx*cos + dy*sin + centerX
			srcY := -dx*sin + dy*cos + centerY

			dstOffset := rowOffset + x*4

			if srcX < 0 || srcX >= srcWf || srcY < 0 || srcY >= srcHf {
				// Transparent pixel
				dstPix[dstOffset] = 0
				dstPix[dstOffset+1] = 0
				dstPix[dstOffset+2] = 0
				dstPix[dstOffset+3] = 0
				continue
			}

			x0 := int(srcX)
			y0 := int(srcY)
			x1 := x0 + 1
			y1 := y0 + 1

			xWeight := srcX - float64(x0)
			yWeight := srcY - float64(y0)
			xWeight1 := 1 - xWeight
			yWeight1 := 1 - yWeight

			// Direct pixel access (4 bytes per pixel: RGBA)
			off00 := y0*srcStride + x0*4
			off10 := y0*srcStride + x1*4
			off01 := y1*srcStride + x0*4
			off11 := y1*srcStride + x1*4

			// Bilinear interpolation weights
			w00 := xWeight1 * yWeight1
			w10 := xWeight * yWeight1
			w01 := xWeight1 * yWeight
			w11 := xWeight * yWeight

			// Inline bilinear interpolation for performance
			dstPix[dstOffset] = uint8(float64(srcPix[off00])*w00 + float64(srcPix[off10])*w10 +
				float64(srcPix[off01])*w01 + float64(srcPix[off11])*w11)
			dstPix[dstOffset+1] = uint8(float64(srcPix[off00+1])*w00 + float64(srcPix[off10+1])*w10 +
				float64(srcPix[off01+1])*w01 + float64(srcPix[off11+1])*w11)
			dstPix[dstOffset+2] = uint8(float64(srcPix[off00+2])*w00 + float64(srcPix[off10+2])*w10 +
				float64(srcPix[off01+2])*w01 + float64(srcPix[off11+2])*w11)
			dstPix[dstOffset+3] = uint8(float64(srcPix[off00+3])*w00 + float64(srcPix[off10+3])*w10 +
				float64(srcPix[off01+3])*w01 + float64(srcPix[off11+3])*w11)
		}
	}
}
