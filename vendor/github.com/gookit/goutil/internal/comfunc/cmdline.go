package comfunc

import (
	"fmt"
	"strings"
)

// Cmdline build
func Cmdline(args []string, binName ...string) string {
	b := new(strings.Builder)

	if len(binName) > 0 {
		b.WriteString(binName[0])
		b.WriteByte(' ')
	}

	for i, a := range args {
		if i > 0 {
			b.WriteByte(' ')
		}

		if strings.ContainsRune(a, '"') {
			b.WriteString(fmt.Sprintf(`'%s'`, a))
		} else if a == "" || strings.ContainsRune(a, '\'') || strings.ContainsRune(a, ' ') {
			b.WriteString(fmt.Sprintf(`"%s"`, a))
		} else {
			b.WriteString(a)
		}
	}
	return b.String()
}
