//go:build darwin && !macapp

package filex

import (
	"embed"
)

//go:embed embed embed_darwin
var embedDir embed.FS
