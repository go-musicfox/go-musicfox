// Package registration is the compile-time frontend aggregator. Each shipped
// frontend package (ui, headless, ...) is blank-imported here so its init()
// registration runs whenever the binary links this package; cmd/musicfox.go
// blank-imports the aggregator.
//
// It must live in a subpackage rather than in frontend itself: ui imports
// frontend, so placing this aggregator in frontend would create the import
// cycle ui → frontend → ui.
package registration

import (
	_ "github.com/go-musicfox/go-musicfox/internal/frontend/gui"
	_ "github.com/go-musicfox/go-musicfox/internal/headless"
	_ "github.com/go-musicfox/go-musicfox/internal/ui"
	_ "github.com/go-musicfox/go-musicfox/internal/webui"
)
