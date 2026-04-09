//go:build !darwin || macapp

package filex

import (
	"embed"
)

//go:embed embed
var embedDir embed.FS
