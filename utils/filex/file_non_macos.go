//go:build !darwin

package filex

import (
	"embed"
)

//go:embed embed/go-musicfox.ini embed/logo.png
var embedDir embed.FS
