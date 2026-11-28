//go:build darwin

package cocoa

import (
	"sync"

	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/objc"

	"github.com/go-musicfox/go-musicfox/internal/macdriver"
	"github.com/go-musicfox/go-musicfox/internal/macdriver/core"
)

// CFRange mirrors Core Foundation's CFRange. CFIndex is a signed 64-bit value
// on all supported macOS architectures.
type CFRange struct {
	Location int64
	Length   int64
}

var (
	coreTextOnce sync.Once

	ctFontCreateForString func(currentFont, str uintptr, rng CFRange) uintptr
	ctFontGetAscent       func(currentFont uintptr) CGFloat
	ctFontGetDescent      func(currentFont uintptr) CGFloat
	ctFontGetLeading      func(currentFont uintptr) CGFloat
	cfRelease             func(cf uintptr)
)

func importCoreText() {
	coreTextOnce.Do(func() {
		ct, err := purego.Dlopen("/System/Library/Frameworks/CoreText.framework/CoreText", purego.RTLD_GLOBAL)
		if err != nil {
			return
		}
		cf, err := purego.Dlopen("/System/Library/Frameworks/CoreFoundation.framework/CoreFoundation", purego.RTLD_GLOBAL)
		if err != nil {
			return
		}

		purego.RegisterLibFunc(&ctFontCreateForString, ct, "CTFontCreateForString")
		purego.RegisterLibFunc(&ctFontGetAscent, ct, "CTFontGetAscent")
		purego.RegisterLibFunc(&ctFontGetDescent, ct, "CTFontGetDescent")
		purego.RegisterLibFunc(&ctFontGetLeading, ct, "CTFontGetLeading")
		purego.RegisterLibFunc(&cfRelease, cf, "CFRelease")
	})
}

func init() {
	importCoreText()
}

// FontForString resolves the font Core Text would use to render text with base,
// following the system cascade list. The returned font is owned by the caller
// and must be released with ReleaseFont.
func FontForString(base NSFont, text string) NSFont {
	nsText := core.NSString_alloc().InitWithUTF8String(text)
	defer nsText.Release()
	length := int(nsText.Send(macdriver.SEL_length))
	return FontForStringRange(base, nsText, 0, length)
}

// FontForStringRange resolves the font Core Text would use to render the given
// UTF-16 range of str with base. The returned font is owned by the caller and
// must be released with ReleaseFont. A zero ID means resolution is unavailable.
func FontForStringRange(base NSFont, str core.NSString, location, length int) NSFont {
	if ctFontCreateForString == nil || str.ID == 0 || length <= 0 {
		return NSFont{}
	}
	id := ctFontCreateForString(uintptr(base.ID), uintptr(str.ID), CFRange{
		Location: int64(location),
		Length:   int64(length),
	})
	return NSFont{core.NSObject{ID: objc.ID(id)}}
}

// FontLineHeight returns the font's rendered line height (ascender - descender
// + leading), matching the vertical metrics the text system uses for layout.
func FontLineHeight(font NSFont) CGFloat {
	if ctFontGetAscent == nil || font.ID == 0 {
		return 0
	}
	ascent := ctFontGetAscent(uintptr(font.ID))
	descent := ctFontGetDescent(uintptr(font.ID))
	leading := ctFontGetLeading(uintptr(font.ID))
	return ascent - descent + leading
}

// ReleaseFont releases a font returned by FontForString/FontForStringRange.
func (f NSFont) ReleaseFont() {
	if f.ID != 0 && cfRelease != nil {
		cfRelease(uintptr(f.ID))
	}
}
