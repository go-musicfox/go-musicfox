package track

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-flac/flacpicture"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/utils/app"

	"github.com/bogem/id3v2/v2"
	songtag "github.com/frolovo22/tag"
)

type Tagger interface {
	SetSongTag(filePath string, song structs.Song) error
}

type metadataTagger struct {
	httpClient *http.Client
}

func NewTagger() Tagger {
	return &metadataTagger{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (m *metadataTagger) SetSongTag(filePath string, song structs.Song) (err error) {
	file, err := os.OpenFile(filePath, os.O_RDWR, 0666)
	if err != nil {
		slog.Error("Failed to open file for setting tags", "file", filePath, "error", err)
		return err
	}

	version := songtag.CheckVersion(file)
	file.Close()

	switch version {
	case songtag.VersionID3v22, songtag.VersionID3v23, songtag.VersionID3v24:
		return m.setID3v2Tag(filePath, song)
	default:
		return m.setGenericTag(filePath, song)
	}
}

func (m *metadataTagger) setID3v2Tag(filePath string, song structs.Song) error {
	tag, err := id3v2.Open(filePath, id3v2.Options{Parse: true})
	if err != nil {
		return fmt.Errorf("failed to open and parse id3v2 file: %w", err)
	}
	defer tag.Close()

	tag.SetDefaultEncoding(id3v2.EncodingUTF8)
	tag.SetTitle(song.Name)
	tag.SetAlbum(song.Album.Name)
	tag.SetArtist(song.ArtistName())

	if coverData, mimeType, err := m.fetchCover(song.PicUrl); err == nil {
		picFrame := id3v2.PictureFrame{
			Encoding:    id3v2.EncodingUTF8,
			MimeType:    mimeType,
			PictureType: id3v2.PTOther,
			Picture:     coverData,
		}
		tag.AddAttachedPicture(picFrame)
	} else {
		slog.Warn("Failed to fetch or set cover image", "songId", song.Id, "error", err)
	}

	return tag.Save()
}

func (m *metadataTagger) setGenericTag(filePath string, song structs.Song) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file for generic tagging: %w", err)
	}

	metadata, err := songtag.Read(file)
	if err != nil {
		return fmt.Errorf("failed to read generic tag: %w", err)
	}
	defer metadata.Close()

	_ = metadata.SetAlbum(song.Album.Name)
	_ = metadata.SetArtist(song.ArtistName())
	_ = metadata.SetAlbumArtist(song.Album.ArtistName())
	_ = metadata.SetTitle(song.Name)
	if flacMeta, ok := metadata.(*songtag.FLAC); ok {
		m.setFlacCover(flacMeta, song)
	}

	originalPath := file.Name()
	tempPath := originalPath + ".tmp"

	if err = metadata.SaveFile(tempPath); err != nil {
		return fmt.Errorf("failed to save metadata to temp file: %w", err)
	}

	if err = file.Close(); err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("failed to close original file before renaming: %w", err)
	}

	if err = os.Rename(tempPath, originalPath); err != nil {
		return fmt.Errorf("failed to rename temp file to original: %w", err)
	}

	return nil
}

func (m *metadataTagger) fetchCover(picURL string) ([]byte, string, error) {
	resp, err := m.httpClient.Get(app.AddResizeParamForPicUrl(picURL, 1024))
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("bad status getting cover: %s", resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}
	return data, resp.Header.Get("Content-Type"), nil
}

func (m *metadataTagger) setFlacCover(flacMeta *songtag.FLAC, song structs.Song) {
	coverData, mimeType, err := m.fetchCover(song.PicUrl)
	if err != nil {
		slog.Warn("Failed to fetch cover image for flac", "songId", song.Id, "error", err)
		return
	}

	img, err := flacpicture.NewFromImageData(flacpicture.PictureTypeFrontCover, "cover", coverData, mimeType)
	if err != nil {
		slog.Warn("Failed to create flac picture from image data", "songId", song.Id, "error", err)
		return
	}

	if err = flacMeta.SetFlacPicture(img); err != nil {
		slog.Warn("Failed to set flac picture", "songId", song.Id, "error", err)
	}
}
