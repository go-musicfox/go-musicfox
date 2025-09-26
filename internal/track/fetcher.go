package track

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/utils/netease"
	"github.com/go-musicfox/netease-music/service"
	"github.com/pkg/errors"
)

type Fetcher interface {
	FetchPlayableInfo(ctx context.Context, songID int64) (*netease.PlayableInfo, error)
	FetchStream(ctx context.Context, source PlayableSource) (io.ReadCloser, error)
	FetchLyric(ctx context.Context, songID int64) (structs.LRCData, error)
}

type fetcher struct {
	httpClient *http.Client
	quality    service.SongQualityLevel
}

type Option func(*fetcher)

func NewFetcher(opts ...Option) Fetcher {
	f := &fetcher{
		httpClient: &http.Client{Timeout: 60 * time.Second},
		quality:    service.Standard,
	}

	for _, opt := range opts {
		opt(f)
	}

	return f
}

func WithFetcherHTTPClient(client *http.Client) Option {
	return func(f *fetcher) {
		f.httpClient = client
	}
}

func WithFetcherSongQuality(quality service.SongQualityLevel) Option {
	return func(f *fetcher) {
		f.quality = quality
	}
}

func (f *fetcher) FetchPlayableInfo(ctx context.Context, songID int64) (*netease.PlayableInfo, error) {
	info, err := netease.FetchPlayableInfo(songID, f.quality)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch playable info for song %d: %w", songID, err)
	}
	return &info, nil
}

func (f *fetcher) FetchStream(ctx context.Context, source PlayableSource) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, source.Info.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create http request for song %d: %w", source.Id, err)
	}

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http get failed for song %d (%s): %w", source.Id, source.Info.URL, err)
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("invalid http status for song %d: %s", source.Id, resp.Status)
	}
	return resp.Body, nil
}

func (f *fetcher) FetchLyric(ctx context.Context, songID int64) (structs.LRCData, error) {
	lrcData, err := netease.FetchLyric(songID)
	if err != nil {
		return structs.LRCData{}, errors.Wrapf(err, "failed to fetch lyric for song %d", songID)
	}
	return lrcData, nil
}
