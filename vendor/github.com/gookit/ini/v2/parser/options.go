package parser

// mode of parse data
//
//	ModeFull   - will parse array value and inline array
//	ModeLite/ModeSimple - don't parse array value line
const (
	ModeFull   parseMode = 1
	ModeLite   parseMode = 2
	ModeSimple parseMode = 2 // alias of ModeLite
)

// DefSection default section key name
const DefSection = "__default"

type parseMode uint8

// Unit8 mode value to uint8
func (m parseMode) Unit8() uint8 {
	return uint8(m)
}

// TagName default tag-name of mapping data to struct
var TagName = "ini"

// OptFunc define
type OptFunc func(opt *Options)

// UserCollector custom data collector.
//
// Notice: in lite mode, isSlice always is false.
type UserCollector func(section, key, val string, isSlice bool)

// Options for parser
type Options struct {
	// TagName of mapping data to struct
	TagName string
	// ParseMode setting. default is ModeLite
	ParseMode parseMode
	// Ignore case for key name
	IgnoreCase bool
	// ReplaceNl replace the "\n" to newline
	ReplaceNl bool
	// default section name. default is "__default"
	DefSection string
	// NoDefSection setting. only for full parse mode
	NoDefSection bool
	// InlineComment support parse inline comments. default is false
	InlineComment bool
	// Collector allow you custom the value collector.
	//
	// Notice: in lite mode, isSlice always is false.
	Collector UserCollector
}

// NewOptions instance
func NewOptions(fns ...OptFunc) *Options {
	opt := &Options{
		TagName:    TagName,
		ParseMode:  ModeLite,
		DefSection: DefSection,
	}

	for _, fn := range fns {
		fn(opt)
	}
	return opt
}

// WithParseMode name for parse
func WithParseMode(mode parseMode) OptFunc {
	return func(opt *Options) {
		opt.ParseMode = mode
	}
}

// InlineComment for parse
func InlineComment(opt *Options) {
	opt.InlineComment = true
}

// WithReplaceNl for parse
func WithReplaceNl(opt *Options) {
	opt.ReplaceNl = true
}

// WithDefSection name for parse
func WithDefSection(name string) OptFunc {
	return func(opt *Options) {
		opt.DefSection = name
	}
}

// WithTagName for decode data
func WithTagName(name string) OptFunc {
	return func(opt *Options) {
		opt.TagName = name
	}
}
