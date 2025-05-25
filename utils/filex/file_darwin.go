//go:build darwin

package filex

import (
	"embed"
)

//go:embed embed embed_darwin
var embedDir embed.FS
