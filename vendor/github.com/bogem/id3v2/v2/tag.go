// Copyright 2016 Albert Nigmatzianov. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package id3v2

import (
	"errors"
	"io"
	"os"
)

var ErrNoFile = errors.New("tag was not initialized with file")

// Tag stores all information about opened tag.
type Tag struct {
	frames    map[string]Framer
	sequences map[string]*sequence

	defaultEncoding Encoding
	reader          io.Reader
	originalSize    int64
	version         byte
}

// AddFrame adds f to tag with appropriate id. If id is "" or f is nil,
// AddFrame will not add it to tag.
//
// If you want to add attached picture, comment or unsynchronised lyrics/text
// transcription frames, better use AddAttachedPicture, AddCommentFrame
// or AddUnsynchronisedLyricsFrame methods respectively.
func (tag *Tag) AddFrame(id string, f Framer) {
	if id == "" || f == nil {
		return
	}

	if mustFrameBeInSequence(id) {
		sequence := tag.sequences[id]
		if sequence == nil {
			sequence = getSequence()
		}
		sequence.AddFrame(f)
		tag.sequences[id] = sequence
	} else {
		tag.frames[id] = f
	}
}

// AddAttachedPicture adds the picture frame to tag.
func (tag *Tag) AddAttachedPicture(pf PictureFrame) {
	tag.AddFrame(tag.CommonID("Attached picture"), pf)
}

// AddChapterFrame adds the chapter frame to tag.
func (tag *Tag) AddChapterFrame(cf ChapterFrame) {
	tag.AddFrame(tag.CommonID("Chapters"), cf)
}

// AddCommentFrame adds the comment frame to tag.
func (tag *Tag) AddCommentFrame(cf CommentFrame) {
	tag.AddFrame(tag.CommonID("Comments"), cf)
}

// AddTextFrame creates the text frame with provided encoding and text
// and adds to tag.
func (tag *Tag) AddTextFrame(id string, encoding Encoding, text string) {
	tag.AddFrame(id, TextFrame{Encoding: encoding, Text: text})
}

// AddUnsynchronisedLyricsFrame adds the unsynchronised lyrics/text frame
// to tag.
func (tag *Tag) AddUnsynchronisedLyricsFrame(uslf UnsynchronisedLyricsFrame) {
	tag.AddFrame(tag.CommonID("Unsynchronised lyrics/text transcription"), uslf)
}

// AddUserDefinedTextFrame adds the custom frame (TXXX) to tag.
func (tag *Tag) AddUserDefinedTextFrame(udtf UserDefinedTextFrame) {
	tag.AddFrame(tag.CommonID("User defined text information frame"), udtf)
}

// AddUFIDFrame adds the unique file identifier frame (UFID) to tag.
func (tag *Tag) AddUFIDFrame(ufid UFIDFrame) {
	tag.AddFrame(tag.CommonID("Unique file identifier"), ufid)
}

// CommonID returns frame ID from given description.
// For example, CommonID("Language") will return "TLAN".
// If it can't find the ID with given description, it returns the description.
//
// All descriptions you can find in file common_ids.go
// or in id3 documentation.
// v2.3: http://id3.org/id3v2.3.0#Declared_ID3v2_frames
// v2.4: http://id3.org/id3v2.4.0-frames
func (tag *Tag) CommonID(description string) string {
	var ids map[string]string
	if tag.version == 3 {
		ids = V23CommonIDs
	} else {
		ids = V24CommonIDs
	}
	if id, ok := ids[description]; ok {
		return id
	}
	return description
}

// AllFrames returns map, that contains all frames in tag, that could be parsed.
// The key of this map is an ID of frame and value is an array of frames.
func (tag *Tag) AllFrames() map[string][]Framer {
	frames := make(map[string][]Framer)

	for id, f := range tag.frames {
		frames[id] = []Framer{f}
	}
	for id, sequence := range tag.sequences {
		frames[id] = sequence.Frames()
	}

	return frames
}

// DeleteAllFrames deletes all frames in tag.
func (tag *Tag) DeleteAllFrames() {
	if tag.frames == nil || len(tag.frames) > 0 {
		tag.frames = make(map[string]Framer)
	}
	if tag.sequences == nil || len(tag.sequences) > 0 {
		for _, s := range tag.sequences {
			putSequence(s)
		}
		tag.sequences = make(map[string]*sequence)
	}
}

// DeleteFrames deletes frames in tag with given id.
func (tag *Tag) DeleteFrames(id string) {
	delete(tag.frames, id)
	if s, ok := tag.sequences[id]; ok {
		putSequence(s)
		delete(tag.sequences, id)
	}
}

// Reset deletes all frames in tag and parses rd considering opts.
func (tag *Tag) Reset(rd io.Reader, opts Options) error {
	tag.DeleteAllFrames()
	return tag.parse(rd, opts)
}

// GetFrames returns frames with corresponding id.
// It returns nil if there is no frames with given id.
func (tag *Tag) GetFrames(id string) []Framer {
	if f, exists := tag.frames[id]; exists {
		return []Framer{f}
	} else if s, exists := tag.sequences[id]; exists {
		return s.Frames()
	}
	return nil
}

// GetLastFrame returns last frame from slice, that is returned from GetFrames function.
// GetLastFrame is suitable for frames, that can be only one in whole tag.
// For example, for text frames.
func (tag *Tag) GetLastFrame(id string) Framer {
	// Avoid an allocation of slice in GetFrames,
	// if there is anyway one frame.
	if f, exists := tag.frames[id]; exists {
		return f
	}

	fs := tag.GetFrames(id)
	if len(fs) == 0 {
		return nil
	}
	return fs[len(fs)-1]
}

// GetTextFrame returns text frame with corresponding id.
func (tag *Tag) GetTextFrame(id string) TextFrame {
	f := tag.GetLastFrame(id)
	if f == nil {
		return TextFrame{}
	}
	tf := f.(TextFrame)
	return tf
}

// DefaultEncoding returns default encoding of tag.
// Default encoding is used in methods (e.g. SetArtist, SetAlbum ...) for
// setting text frames without the explicit providing of encoding.
func (tag *Tag) DefaultEncoding() Encoding {
	return tag.defaultEncoding
}

// SetDefaultEncoding sets default encoding for tag.
// Default encoding is used in methods (e.g. SetArtist, SetAlbum ...) for
// setting text frames without explicit providing encoding.
func (tag *Tag) SetDefaultEncoding(encoding Encoding) {
	tag.defaultEncoding = encoding
}

func (tag *Tag) setDefaultEncodingBasedOnVersion(version byte) {
	if version == 4 {
		tag.SetDefaultEncoding(EncodingUTF8)
	} else {
		tag.SetDefaultEncoding(EncodingISO)
	}
}

// Count returns the number of frames in tag.
func (tag *Tag) Count() int {
	n := len(tag.frames)
	for _, s := range tag.sequences {
		n += s.Count()
	}
	return n
}

// HasFrames checks if there is at least one frame in tag.
// It's much faster than tag.Count() > 0.
func (tag *Tag) HasFrames() bool {
	return len(tag.frames) > 0 || len(tag.sequences) > 0
}

func (tag *Tag) Title() string {
	return tag.GetTextFrame(tag.CommonID("Title")).Text
}

func (tag *Tag) SetTitle(title string) {
	tag.AddTextFrame(tag.CommonID("Title"), tag.DefaultEncoding(), title)
}

func (tag *Tag) Artist() string {
	return tag.GetTextFrame(tag.CommonID("Artist")).Text
}

func (tag *Tag) SetArtist(artist string) {
	tag.AddTextFrame(tag.CommonID("Artist"), tag.DefaultEncoding(), artist)
}

func (tag *Tag) Album() string {
	return tag.GetTextFrame(tag.CommonID("Album/Movie/Show title")).Text
}

func (tag *Tag) SetAlbum(album string) {
	tag.AddTextFrame(tag.CommonID("Album/Movie/Show title"), tag.DefaultEncoding(), album)
}

func (tag *Tag) Year() string {
	return tag.GetTextFrame(tag.CommonID("Year")).Text
}

func (tag *Tag) SetYear(year string) {
	tag.AddTextFrame(tag.CommonID("Year"), tag.DefaultEncoding(), year)
}

func (tag *Tag) Genre() string {
	return tag.GetTextFrame(tag.CommonID("Content type")).Text
}

func (tag *Tag) SetGenre(genre string) {
	tag.AddTextFrame(tag.CommonID("Content type"), tag.DefaultEncoding(), genre)
}

// iterateOverAllFrames iterates over every single frame in tag and calls
// f for them. It consumps no memory at all, unlike the tag.AllFrames().
// It returns error only if f returns error.
func (tag *Tag) iterateOverAllFrames(f func(id string, frame Framer) error) error {
	for id, frame := range tag.frames {
		if err := f(id, frame); err != nil {
			return err
		}
	}
	for id, sequence := range tag.sequences {
		for _, frame := range sequence.Frames() {
			if err := f(id, frame); err != nil {
				return err
			}
		}
	}
	return nil
}

// Size returns the size of tag (tag header + size of all frames) in bytes.
func (tag *Tag) Size() int {
	if !tag.HasFrames() {
		return 0
	}

	var n int
	n += tagHeaderSize // Add the size of tag header
	tag.iterateOverAllFrames(func(id string, f Framer) error {
		n += frameHeaderSize + f.Size() // Add the whole frame size
		return nil
	})

	return n
}

// Version returns current ID3v2 version of tag.
func (tag *Tag) Version() byte {
	return tag.version
}

// SetVersion sets given ID3v2 version to tag.
// If version is less than 3 or greater than 4, then this method will do nothing.
// If tag has some frames, which are deprecated or changed in given version,
// then to your notice you can delete, change or just stay them.
func (tag *Tag) SetVersion(version byte) {
	if version < 3 || version > 4 {
		return
	}
	tag.version = version
	tag.setDefaultEncodingBasedOnVersion(version)
}

// Save writes tag to the file, if tag was opened with a file.
// If there are no frames in tag, Save will write
// only music part without any ID3v2 information.
// If tag was initiliazed not with file, it returns ErrNoFile.
func (tag *Tag) Save() error {
	file, ok := tag.reader.(*os.File)
	if !ok {
		return ErrNoFile
	}

	// Get original file mode.
	originalFile := file
	originalStat, err := originalFile.Stat()
	if err != nil {
		return err
	}

	// Create a temp file for mp3 file, which will contain new tag.
	name := file.Name() + "-id3v2"
	newFile, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE, originalStat.Mode())
	if err != nil {
		return err
	}

	// Make sure we clean up the temp file if it's still around.
	// tempfileShouldBeRemoved created only for performance
	// improvement to prevent calling redundant Remove syscalls if file is moved
	// and is not need to be removed.
	tempfileShouldBeRemoved := true
	defer func() {
		if tempfileShouldBeRemoved {
			os.Remove(newFile.Name())
		}
	}()

	// Write tag in new file.
	tagSize, err := tag.WriteTo(newFile)
	if err != nil {
		return err
	}

	// Seek to a music part of original file.
	if _, err = originalFile.Seek(tag.originalSize, os.SEEK_SET); err != nil {
		return err
	}

	// Write to new file the music part.
	buf := getByteSlice(128 * 1024)
	defer putByteSlice(buf)
	if _, err = io.CopyBuffer(newFile, originalFile, buf); err != nil {
		return err
	}

	// Close files to allow replacing.
	newFile.Close()
	originalFile.Close()

	// Replace original file with new file.
	if err = os.Rename(newFile.Name(), originalFile.Name()); err != nil {
		return err
	}
	tempfileShouldBeRemoved = false

	// Set tag.reader to new file with original name.
	tag.reader, err = os.Open(originalFile.Name())
	if err != nil {
		return err
	}

	// Set tag.originalSize to new frames size.
	tag.originalSize = tagSize

	return nil
}

// WriteTo writes whole tag in w if there is at least one frame.
// It returns the number of bytes written and error during the write.
// It returns nil as error if the write was successful.
func (tag *Tag) WriteTo(w io.Writer) (n int64, err error) {
	if w == nil {
		return 0, errors.New("w is nil")
	}

	// Count size of frames.
	framesSize := tag.Size() - tagHeaderSize
	if framesSize <= 0 {
		return 0, nil
	}

	// Write tag header.
	bw := getBufWriter(w)
	defer putBufWriter(bw)
	writeTagHeader(bw, uint(framesSize), tag.version)

	// Write frames.
	synchSafe := tag.Version() == 4
	err = tag.iterateOverAllFrames(func(id string, f Framer) error {
		return writeFrame(bw, id, f, synchSafe)
	})
	if err != nil {
		bw.Flush()
		return int64(bw.Written()), err
	}

	return int64(bw.Written()), bw.Flush()
}

func writeTagHeader(bw *bufWriter, framesSize uint, version byte) {
	bw.Write(id3Identifier)
	bw.WriteByte(version)
	bw.WriteByte(0) // Revision
	bw.WriteByte(0) // Flags
	bw.WriteBytesSize(framesSize, true)
}

func writeFrame(bw *bufWriter, id string, frame Framer, synchSafe bool) error {
	writeFrameHeader(bw, id, uint(frame.Size()), synchSafe)
	_, err := frame.WriteTo(bw)
	return err
}

func writeFrameHeader(bw *bufWriter, id string, frameSize uint, synchSafe bool) {
	bw.WriteString(id)
	bw.WriteBytesSize(frameSize, synchSafe)
	bw.Write([]byte{0, 0}) // Flags
}

// Close closes tag's file, if tag was opened with a file.
// If tag was initiliazed not with file, it returns ErrNoFile.
func (tag *Tag) Close() error {
	file, ok := tag.reader.(*os.File)
	if !ok {
		return ErrNoFile
	}
	return file.Close()
}
