package lyric

import (
	"context"

	"github.com/go-musicfox/go-musicfox/internal/structs"
)

// Fetcher defines the behavior for retrieving structured lyric data for a song.
type Fetcher interface {
	GetLyric(ctx context.Context, songID int64) (structs.LRCData, error)
}
