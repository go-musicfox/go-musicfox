// Package plugins is the compile-time plugin aggregator. Each shipped plugin
// package is blank-imported here so its init() registration runs whenever the
// binary links this package; cmd/musicfox.go blank-imports the aggregator.
package plugins

import (
	_ "github.com/go-musicfox/go-musicfox/internal/plugins/album"
	_ "github.com/go-musicfox/go-musicfox/internal/plugins/artist"
	_ "github.com/go-musicfox/go-musicfox/internal/plugins/checkupdate"
	_ "github.com/go-musicfox/go-musicfox/internal/plugins/dj"
	_ "github.com/go-musicfox/go-musicfox/internal/plugins/lastfm"
	_ "github.com/go-musicfox/go-musicfox/internal/plugins/playlist"
	_ "github.com/go-musicfox/go-musicfox/internal/plugins/recommend"
)
