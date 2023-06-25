package ini

import "github.com/gookit/ini/v2/parser"

// Options for config
type Options struct {
	// Readonly set to read-only mode. default False
	Readonly bool
	// TagName for binding struct
	TagName string
	// ParseEnv parse ENV var name. default True
	ParseEnv bool
	// ParseVar parse variable reference "%(varName)s". default False
	ParseVar bool
	// ReplaceNl replace the "\n" to newline
	ReplaceNl bool

	// VarOpen var left open char. default "%("
	VarOpen string
	// VarClose var right close char. default ")s"
	VarClose string

	// IgnoreCase ignore key name case. default False
	IgnoreCase bool
	// DefSection default section name. default "__default", it's allow empty string.
	DefSection string
	// SectionSep sep char for split key path. default ".", use like "section.subKey"
	SectionSep string
}

// newDefaultOptions create a new default Options
//
// Notice:
//
//	Cannot use package var instead it. That will allow multiple instances to use the same Options
func newDefaultOptions() *Options {
	return &Options{
		ParseEnv: true,

		VarOpen:  "%(",
		VarClose: ")s",
		TagName:  DefTagName,

		DefSection: parser.DefSection,
		SectionSep: SepSection,
	}
}

// Readonly setting
//
// Usage:
//
//	ini.NewWithOptions(ini.Readonly)
func Readonly(opts *Options) {
	opts.Readonly = true
}

// ParseVar on get value
//
// Usage:
//
//	ini.NewWithOptions(ini.ParseVar)
func ParseVar(opts *Options) {
	opts.ParseVar = true
}

// ParseEnv will parse ENV key on get value
//
// Usage:
//
//	ini.NewWithOptions(ini.ParseEnv)
func ParseEnv(opts *Options) {
	opts.ParseEnv = true
}

// IgnoreCase for get/set value by key
func IgnoreCase(opts *Options) {
	opts.IgnoreCase = true
}

// ReplaceNl for parse
func ReplaceNl(opts *Options) {
	opts.ReplaceNl = true
}
