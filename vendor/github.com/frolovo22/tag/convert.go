package tag

import (
	"fmt"
)

// nolint:gocyclo
func GetMap(metadata Metadata) map[string]interface{} {
	var tags = map[string]interface{}{}
	var val interface{}
	var err error

	if val, err = metadata.GetTitle(); err == nil {
		tags["title"] = val
	}
	if val, err = metadata.GetArtist(); err == nil {
		tags["artist"] = val
	}
	if val, err = metadata.GetAlbum(); err == nil {
		tags["album"] = val
	}
	if val, err = metadata.GetYear(); err == nil {
		tags["year"] = val
	}
	if val, err = metadata.GetComment(); err == nil {
		tags["comment"] = val
	}
	if val, err = metadata.GetGenre(); err == nil {
		tags["genre"] = val
	}
	if val, err = metadata.GetAlbumArtist(); err == nil {
		tags["album artist"] = val
	}
	if val, err = metadata.GetDate(); err == nil {
		tags["date"] = val
	}
	if val, err = metadata.GetArranger(); err == nil {
		tags["arranger"] = val
	}
	if val, err = metadata.GetAuthor(); err == nil {
		tags["author"] = val
	}
	if val, err = metadata.GetBPM(); err == nil {
		tags["bmp"] = val
	}
	if val, err = metadata.GetCatalogNumber(); err == nil {
		tags["catalog number"] = val
	}
	if val, err = metadata.GetCompilation(); err == nil {
		tags["compilation"] = val
	}
	if val, err = metadata.GetComposer(); err == nil {
		tags["composer"] = val
	}
	if val, err = metadata.GetConductor(); err == nil {
		tags["conductor"] = val
	}
	if val, err = metadata.GetCopyright(); err == nil {
		tags["copyright"] = val
	}
	if val, err = metadata.GetDescription(); err == nil {
		tags["description"] = val
	}
	if number, total, errGetDiskNumber := metadata.GetDiscNumber(); errGetDiskNumber == nil {
		tags["disc number"] = fmt.Sprintf("%d/%d", number, total)
	}
	if val, err = metadata.GetEncodedBy(); err == nil {
		tags["encoded by"] = val
	}
	if number, total, err := metadata.GetTrackNumber(); err == nil {
		tags["track number"] = fmt.Sprintf("%d/%d", number, total)
	}

	return tags
}
