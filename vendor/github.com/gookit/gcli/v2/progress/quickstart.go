package progress

import (
	"math/rand"
	"time"
)

// some built in chars
const (
	CharStar    rune = '*'
	CharPlus    rune = '+'
	CharWell    rune = '#'
	CharEqual   rune = '='
	CharEqual1  rune = '═'
	CharSpace   rune = ' '
	CharCenter  rune = '●'
	CharSquare  rune = '■'
	CharSquare1 rune = '▇'
	CharSquare2 rune = '▉'
	CharSquare3 rune = '░'
	CharSquare4 rune = '▒'
	CharSquare5 rune = '▢'
	// Hyphen Minus
	CharHyphen      rune = '-'
	CharCNHyphen    rune = '—'
	CharUnderline   rune = '_'
	CharLeftArrow   rune = '<'
	CharRightArrow  rune = '>'
	CharRightArrow1 rune = '▶'
)

// internal format for text progress
const (
	MinFormat  = "{@message}{@current}"
	TxtFormat  = "{@message}{@percent:4s}%({@current}/{@max})"
	DefFormat  = "{@message}{@percent:4s}%({@current}/{@max})"
	FullFormat = "{@percent:4s}%({@current}/{@max}) {@elapsed:7s}/{@estimated:-7s} {@memory:6s}"
)

// Txt progress bar create.
func Txt(maxSteps ...int) *Progress {
	return New(maxSteps...).Config(func(p *Progress) {
		p.Format = TxtFormat
	})
}

// Full text progress bar create.
func Full(maxSteps ...int) *Progress {
	return NewWithConfig(func(p *Progress) {
		p.Format = FullFormat
	}, maxSteps...)
}

// Counter progress bar create
func Counter(maxSteps ...int) *Progress {
	return NewWithConfig(func(p *Progress) {
		p.Format = MinFormat
	}, maxSteps...)
}

// DynamicText progress bar create
func DynamicText(messages map[int]string, maxSteps ...int) *Progress {
	return NewWithConfig(func(p *Progress) {
		p.Format = "{@percent:4s}%({@current}/{@max}){@message}"
		p.AddWidget("message", DynamicTextWidget(messages))
	}, maxSteps...)
}

/*************************************************************
 * Generic progress bar
 *************************************************************/

// internal format for ProgressBar
const (
	// BarWidth default bar width
	BarWidth  = 40
	BarFormat = "{@bar} {@percent:4s}%({@current}/{@max}){@message}"

	// MdlBarFormat more format
	MdlBarFormat  = "{@bar} {@percent:4s}%({@current}/{@max}) {@elapsed:7s}/{@estimated:-7s}"
	FullBarFormat = "{@bar} {@percent:4s}%({@current}/{@max}) {@elapsed:7s}/{@estimated:-7s} {@memory:6s}"
)

// BarChars setting for a progress bar. default {'#', '>', ' '}
type BarChars struct {
	Completed, Processing, Remaining rune
}

// BarStyles some built in BarChars style collection
var BarStyles = []BarChars{
	{'=', '>', ' '},
	{'=', '>', '-'},
	{'#', '>', ' '},
	{'#', '>', '-'},
	{'*', '>', '-'},
	{'▉', '▉', '░'},
	{'■', '■', ' '},
	{'■', '■', '▢'},
	{'■', '▶', ' '},
}

// Bar create a default progress bar.
// Preview:
// 		1 [->--------------------------]
// 		3 [■■■>------------------------]
// 	25/50 [==============>-------------]  50%
func Bar(maxSteps ...int) *Progress {
	return CustomBar(BarWidth, BarStyles[0], maxSteps...)
}

// Tape create new tape progress bar. is alias of Bar()
func Tape(maxSteps ...int) *Progress {
	return Bar(maxSteps...)
}

// CustomBar create a custom progress bar.
func CustomBar(width int, cs BarChars, maxSteps ...int) *Progress {
	return New(maxSteps...).
		Config(func(p *Progress) {
			p.Format = BarFormat
		}).
		AddWidget("bar", BarWidget(width, cs))
}

// RandomBarStyle get random bar style
func RandomBarStyle() BarChars {
	rand.Seed(time.Now().UnixNano())
	return BarStyles[rand.Intn(len(BarStyles)-1)]
}

/*************************************************************
 * RoundTrip progress bar: `[ ====   ] Pending ...`
 *************************************************************/

// RoundTripBar alias of RoundTrip()
func RoundTripBar(char rune, charNumAndBoxWidth ...int) *Progress {
	return RoundTrip(char, charNumAndBoxWidth...)
}

// RoundTrip create a RoundTrip progress bar.
// Usage:
// 	p := RoundTrip(CharEqual)
// 	// p := RoundTrip('*') // custom char
// 	p.Start()
// 	....
// 	p.Finish()
func RoundTrip(char rune, charNumAndBoxWidth ...int) *Progress {
	charNum := 4
	boxWidth := 12
	if ln := len(charNumAndBoxWidth); ln > 0 {
		charNum = charNumAndBoxWidth[0]
		if ln > 1 {
			boxWidth = charNumAndBoxWidth[1]
		}
	}

	return New().
		AddWidget("rtBar", RoundTripWidget(char, charNum, boxWidth)).
		Config(func(p *Progress) {
			p.Format = "[{@rtBar}] {@percent:4s}% ({@current}/{@max}){@message}"
		})
}

/*************************************************************
 * Loading bar
 *************************************************************/

// LoadingBar alias of load bar LoadBar()
func LoadingBar(chars []rune, maxSteps ...int) *Progress {
	return LoadBar(chars, maxSteps...)
}

// SpinnerBar alias of load bar LoadBar()
func SpinnerBar(chars []rune, maxSteps ...int) *Progress {
	return LoadBar(chars, maxSteps...)
}

// LoadBar create a loading progress bar
func LoadBar(chars []rune, maxSteps ...int) *Progress {
	return New(maxSteps...).Config(func(p *Progress) {
		p.Format = "{@loading} ({@current}/{@max}){@message}"
		p.AddWidget("loading", LoadingWidget(chars))
	})
}
