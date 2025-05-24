//go:build !darwin

package filex

import (
	"embed"
)

//go:embed embed
var embedDir embed.FS
