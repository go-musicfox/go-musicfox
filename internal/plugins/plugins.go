// Package plugins is the compile-time plugin aggregator. Each shipped plugin
// package is blank-imported here so its init() runs whenever the binary links
// this package; cmd/musicfox.go blank-imports the aggregator. Since P5 the
// init() only declares the plugin constructor in the framework registry
// (framework.RegisterPlugin) — the actual menu/page/main-menu registrations
// happen in each plugin's Start, which the frontend scope (internal/ui)
// drives for the enabled subset.
package plugins

import (
	_ "github.com/go-musicfox/go-musicfox/internal/plugins/album"
	_ "github.com/go-musicfox/go-musicfox/internal/plugins/artist"
	_ "github.com/go-musicfox/go-musicfox/internal/plugins/checkupdate"
	_ "github.com/go-musicfox/go-musicfox/internal/plugins/dj"
	_ "github.com/go-musicfox/go-musicfox/internal/plugins/lastfm"
	_ "github.com/go-musicfox/go-musicfox/internal/plugins/playlist"
	_ "github.com/go-musicfox/go-musicfox/internal/plugins/recommend"
	_ "github.com/go-musicfox/go-musicfox/internal/plugins/search"
	_ "github.com/go-musicfox/go-musicfox/internal/plugins/song"
)
