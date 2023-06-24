// Package cmdline provide quick build and parse cmd line string.
package cmdline

// LineBuild build command line string by given args.
func LineBuild(binFile string, args []string) string {
	return NewBuilder(binFile, args...).String()
}

// ParseLine input command line text. alias of the StringToOSArgs()
func ParseLine(line string) []string {
	return NewParser(line).Parse()
}
