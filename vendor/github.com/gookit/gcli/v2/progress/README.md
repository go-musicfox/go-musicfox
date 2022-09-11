# Progress Display

Package progress provide terminal progress bar display. Such as: `Txt`, `Bar`, `Loading`, `RoundTrip`, `DynamicText` ...

- progress bar
- text progress bar
- pending/loading progress bar
- counter
- dynamic Text

## GoDoc

please see https://godoc.org/github.com/gookit/gcli/progress

## Install

```bash
go get -u github.com/gookit/gcli/v2
```

## Usage

Examples:

```go
package main

import (
	"time"

	"github.com/gookit/gcli/v2/progress"
)

func main()  {
	speed := 100
	maxSteps := 110
	p := progress.Bar(maxSteps)
	p.Start()

	for i := 0; i < maxSteps; i++ {
		time.Sleep(time.Duration(speed) * time.Millisecond)
		p.Advance()
	}

	p.Finish()
}
```

> More demos please see [progress_demo.go](../_examples/cmd/progress_demo.go)

run demos:

```bash
go run ./_examples/cliapp.go prog txt
go run ./_examples/cliapp.go prog bar
go run ./_examples/cliapp.go prog roundTrip
```

## Progress Bar

### Internal Widgets

Widget Name | Usage example | Description
------------|----------------|----------------
`max`  | `{@max}` | Display max steps for progress bar
`current`  | `{@current}` | Display current steps for progress bar
`percent`  | `{@percent:4s}` | Display percent for progress run
`elapsed`  | `{@elapsed:7s}` | Display has elapsed time for progress run
`remaining`  | `{@remaining:7s}` | Display remaining time
`estimated`  | `{@estimated:-7s}` | Display estimated time
`memory`   | `{@memory:6s}` | Display memory consumption size

### Custom Progress Bar

Allow you custom progress bar render format. There are internal format for Progress

```go
// txt bar
MinFormat  = "{@message}{@current}"
TxtFormat  = "{@message}{@percent:4s}%({@current}/{@max})"
DefFormat  = "{@message}{@percent:4s}%({@current}/{@max})"
FullFormat = "{@percent:4s}%({@current}/{@max}) {@elapsed:7s}/{@estimated:-7s} {@memory:6s}"

// bar

DefBarFormat  = "{@bar} {@percent:4s}%({@current}/{@max}){@message}"
FullBarFormat = "{@bar} {@percent:4s}%({@current}/{@max}) {@elapsed:7s}/{@estimated:-7s} {@memory:6s}"
```

Examples:

```go
package main
import "github.com/gookit/gcli/v2/progress"

// CustomBar create a custom progress bar
func main() {
    maxSteps := 100
	// use special bar style: [==============>-------------]
	// barStyle := progress.BarStyles[0]
	// use random bar style
	barStyle := progress.RandomBarStyle()

	p: = progress.New(maxSteps).
	Config(func(p *Progress) {
		p.Format = progress.DefBarFormat
	}).
	AddWidget("bar", progress.BarWidget(60, barStyle))

	p.Start()

	for i := 0; i < maxStep; i++ {
		time.Sleep(80 * time.Millisecond)
		p.Advance()
	}

	p.Finish()
}
```

### Progress Functions

Quick create progress bar:

```text
func Bar(maxSteps ...int) *Progress
func Counter(maxSteps ...int) *Progress
func CustomBar(width int, cs BarChars, maxSteps ...int) *Progress
func DynamicText(messages map[int]string, maxSteps ...int) *Progress
func Full(maxSteps ...int) *Progress
func LoadBar(chars []rune, maxSteps ...int) *Progress
func LoadingBar(chars []rune, maxSteps ...int) *Progress
func New(maxSteps ...int) *Progress
func NewWithConfig(fn func(p *Progress), maxSteps ...int) *Progress
func RoundTrip(char rune, charNumAndBoxWidth ...int) *Progress
func RoundTripBar(char rune, charNumAndBoxWidth ...int) *Progress
func SpinnerBar(chars []rune, maxSteps ...int) *Progress
func Tape(maxSteps ...int) *Progress
func Txt(maxSteps ...int) *Progress
```

## Spinner Bar

### Spinner Functions

Quick create progress spinner:

```text
func LoadingSpinner(chars []rune, speed time.Duration) *SpinnerFactory
func RoundTripLoading(char rune, speed time.Duration, charNumAndBoxWidth ...int) *SpinnerFactory
func RoundTripSpinner(char rune, speed time.Duration, charNumAndBoxWidth ...int) *SpinnerFactory
func Spinner(speed time.Duration) *SpinnerFactory
```

## Related

- https://github.com/vbauerster/mpb
- https://github.com/schollz/progressbar
- https://github.com/gosuri/uiprogress