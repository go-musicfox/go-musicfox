package track

import (
	"bytes"
	"image"
	"image/jpeg"
	"image/png"
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

func TestSetFlacCoverDetectsImageTypeFromData(t *testing.T) {
	testCases := []struct {
		name     string
		mimeType string
		encode   func(*bytes.Buffer) error
	}{
		{
			name:     "JPEG",
			mimeType: "image/jpeg",
			encode: func(cover *bytes.Buffer) error {
				return jpeg.Encode(cover, image.NewRGBA(image.Rect(0, 0, 1, 1)), nil)
			},
		},
		{
			name:     "PNG with JPEG response header",
			mimeType: "image/png",
			encode: func(cover *bytes.Buffer) error {
				return png.Encode(cover, image.NewRGBA(image.Rect(0, 0, 1, 1)))
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var cover bytes.Buffer
			if err := testCase.encode(&cover); err != nil {
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

			picture, err := flacMetadata.GetMetadataBlockPicture()
			if err != nil {
				t.Fatalf("read FLAC picture metadata block: %v", err)
			}
			if picture.MIME != testCase.mimeType {
				t.Fatalf("expected %s MIME, got %q", testCase.mimeType, picture.MIME)
			}
		})
	}
}
