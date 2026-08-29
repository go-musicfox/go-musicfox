package webui

import (
	"embed"
	"io/fs"
)

// staticFS embeds the WebUI static assets (the placeholder page for T2; the
// full implementation replaces it in a later ticket). assets.go must stay in
// the same directory as static/.
//
//go:embed static
var staticFS embed.FS

// staticRoot is the embedded filesystem rooted at the contents of static/ so
// http.FileServer can serve index.html at "/" (the raw embed root would expose
// a "static/" directory entry instead).
var staticRoot = func() fs.FS {
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		// static/ is embedded next to this file; unreachable in practice.
		panic(err)
	}
	return sub
}()
