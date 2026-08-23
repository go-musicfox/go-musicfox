// Package plugins is the compile-time plugin aggregator. Each shipped plugin
// package is blank-imported here so its init() registration runs whenever the
// binary links this package; cmd/musicfox.go blank-imports the aggregator.
package plugins

import (
	_ "github.com/go-musicfox/go-musicfox/internal/plugins/checkupdate"
)
