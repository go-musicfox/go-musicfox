package track

import (
	"bytes"
	"image"
	"image/jpeg"
	"io"
	"net/http"
	"testing"

	songtag "github.com/frolovo22/tag"
	"github.com/go-musicfox/go-musicfox/internal/structs"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func TestSetFlacCoverAcceptsImageJPGAlias(t *testing.T) {
	var cover bytes.Buffer
	if err := jpeg.Encode(&cover, image.NewRGBA(image.Rect(0, 0, 1, 1)), nil); err != nil {
		t.Fatalf("encode test cover: %v", err)
	}

	flacMetadata := &songtag.FLAC{}
	tagger := metadataTagger{httpClient: &http.Client{
		Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
			return &http.Response{
				Status:     "200 OK",
				StatusCode: http.StatusOK,
				Header: http.Header{
					"Content-Type": []string{"image/jpg"},
				},
				Body: io.NopCloser(bytes.NewReader(cover.Bytes())),
			}, nil
		}),
	}}
	tagger.setFlacCover(flacMetadata, structs.Song{
		Id: 1,
		Album: structs.Album{
			PicUrl: "https://example.com/cover.jpg",
		},
	})

	for _, block := range flacMetadata.Blocks {
		if block.Type == songtag.FlacPicture {
			return
		}
	}
	t.Fatal("expected a FLAC picture metadata block")
}
