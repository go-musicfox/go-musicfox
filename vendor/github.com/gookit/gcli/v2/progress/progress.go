// Package progress provide terminal progress bar display.
// Such as: `Txt`, `Bar`, `Loading`, `RoundTrip`, `DynamicText` ...
package progress

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/gookit/color"
)

// use for match like "{@bar}" "{@percent:3s}"
var widgetMatch = regexp.MustCompile(`{@([\w]+)(?::([\w-]+))?}`)

// WidgetFunc handler func for progress widget
type WidgetFunc func(p *Progress) string

// Progresser progress interface
type Progresser interface {
	Start(maxSteps ...int)
	Advance(steps ...uint)
	AdvanceTo(step uint)
	Finish(msg ...string)
	Bound() interface{}
}

// Progress definition
// Refer:
// 	https://github.com/inhere/php-console/blob/master/src/utils/ProgressBar.php
type Progress struct {
	// Format string the bar format
	Format string
	// Newline render progress on newline
	Newline bool
	// MaxSteps maximal steps.
	MaxSteps uint
	// StepWidth the width for display "{@current}". default is 2
	StepWidth uint8
	// Overwrite prev output. default is True
	Overwrite bool
	// RedrawFreq redraw freq. default is 1
	RedrawFreq uint8
	// Widgets for build the progress bar
	Widgets map[string]WidgetFunc
	// Messages named messages for build progress bar
	// Example:
	// 	{"msg": "downloading ..."}
	// 	"{@percent}% {@msg}" -> "83% downloading ..."
	Messages map[string]string
	// current step value
	step uint
	// bound user custom data.
	bound interface{}
	// mark start status
	started bool
	// completed percent. eg: "83.8"
	percent float32
	// mark is first running
	firstRun bool
	// time consumption record
	startedAt  time.Time
	finishedAt time.Time
}

/*************************************************************
 * quick use
 *************************************************************/

// New Progress instance
func New(maxSteps ...int) *Progress {
	var max uint
	if len(maxSteps) > 0 {
		max = uint(maxSteps[0])
	}

	return &Progress{
		Format:    DefFormat,
		MaxSteps:  max,
		StepWidth: 3,
		Overwrite: true,
		// init widgets
		Widgets: make(map[string]WidgetFunc),
		// add a default message
		Messages: map[string]string{"message": ""},
	}
}

// NewWithConfig create new Progress
func NewWithConfig(fn func(p *Progress), maxSteps ...int) *Progress {
	return New(maxSteps...).Config(fn)
}

/*************************************************************
 * Some quick config methods
 *************************************************************/

// RenderFormat set rendered format option
func RenderFormat(f string) func(p *Progress) {
	return func(p *Progress) {
		p.Format = f
	}
}

// MaxSteps setting max steps
func MaxSteps(maxStep int) func(p *Progress) {
	return func(p *Progress) {
		p.MaxSteps = uint(maxStep)
	}
}

/*************************************************************
 * Config progress
 *************************************************************/

// Config the progress instance
func (p *Progress) Config(fn func(p *Progress)) *Progress {
	fn(p)
	return p
}

// WithOptions add more option at once for the progress instance
func (p *Progress) WithOptions(fns ...func(p *Progress)) *Progress {
	if len(fns) > 0 {
		for _, fn := range fns {
			fn(p)
		}
	}
	return p
}

// WithMaxSteps setting max steps
func (p *Progress) WithMaxSteps(maxSteps ...int) *Progress {
	if len(maxSteps) > 0 {
		p.MaxSteps = uint(maxSteps[0])
	}
	return p
}

// Binding user custom data to instance
func (p *Progress) Binding(data interface{}) *Progress {
	p.bound = data
	return p
}

// Bound get bound sub struct instance
func (p *Progress) Bound() interface{} {
	return p.bound
}

// AddMessage to progress instance
func (p *Progress) AddMessage(name, message string) {
	p.Messages[name] = message
}

// AddMessages to progress instance
func (p *Progress) AddMessages(msgMap map[string]string) {
	if p.Messages == nil {
		p.Messages = make(map[string]string)
	}

	for name, message := range msgMap {
		p.Messages[name] = message
	}
}

// AddWidget to progress instance
func (p *Progress) AddWidget(name string, handler WidgetFunc) *Progress {
	if _, ok := p.Widgets[name]; !ok {
		p.Widgets[name] = handler
	}
	return p
}

// SetWidget to progress instance
func (p *Progress) SetWidget(name string, handler WidgetFunc) *Progress {
	p.Widgets[name] = handler
	return p
}

// AddWidgets to progress instance
func (p *Progress) AddWidgets(widgets map[string]WidgetFunc) {
	if p.Widgets == nil {
		p.Widgets = make(map[string]WidgetFunc)
	}

	for name, handler := range widgets {
		p.AddWidget(name, handler)
	}
}

/*************************************************************
 * running
 *************************************************************/

// Start the progress bar
func (p *Progress) Start(maxSteps ...int) {
	if p.started {
		panic("Progress bar already started")
	}

	// init
	p.init(maxSteps...)

	// render
	p.Display()
}

func (p *Progress) init(maxSteps ...int) {
	p.step = 0
	p.percent = 0.0
	p.started = true
	p.firstRun = true
	p.startedAt = time.Now()

	if p.RedrawFreq == 0 {
		p.RedrawFreq = 1
	}

	if len(maxSteps) > 0 {
		p.MaxSteps = uint(maxSteps[0])
	}

	// use MaxSteps len as StepWidth. eg: MaxSteps=1000 -> StepWidth=4
	if p.MaxSteps > 0 {
		maxStepsLen := len(fmt.Sprint(p.MaxSteps))
		p.StepWidth = uint8(maxStepsLen)
	}

	if p.StepWidth == 0 {
		p.StepWidth = 3
	}

	// load default widgets
	p.AddWidgets(builtinWidgets)
}

// Advance specified step size. default is 1
func (p *Progress) Advance(steps ...uint) {
	p.checkStart()

	var step uint = 1
	if len(steps) > 0 && steps[0] > 0 {
		step = steps[0]
	}

	p.AdvanceTo(p.step + step)
}

// AdvanceTo specified number of steps
func (p *Progress) AdvanceTo(step uint) {
	p.checkStart()

	// check arg
	if p.MaxSteps > 0 && step > p.MaxSteps {
		p.MaxSteps = step
	}

	freq := uint(p.RedrawFreq)
	prevPeriod := int(p.step / freq)
	currPeriod := int(step / freq)

	p.step = step
	if p.MaxSteps > 0 {
		p.percent = float32(p.step) / float32(p.MaxSteps)
	}

	if prevPeriod != currPeriod || p.MaxSteps == step {
		p.Display()
	}
}

// Finish the progress output.
// if provide finish message, will delete progress bar then print the message.
func (p *Progress) Finish(message ...string) {
	p.checkStart()
	p.finishedAt = time.Now()

	if p.MaxSteps == 0 {
		p.MaxSteps = p.step
	}

	// prevent double 100% output
	if p.step == p.MaxSteps && !p.Overwrite {
		return
	}

	p.AdvanceTo(p.MaxSteps)

	if len(message) > 0 {
		p.render(message[0])
	}

	fmt.Println() // new line
}

// Display outputs the current progress string.
func (p *Progress) Display() {
	if p.Format == "" {
		p.Format = DefFormat
	}

	p.render(p.buildLine())
}

// Destroy removes the progress bar from the current line.
//
// This is useful if you wish to write some output while a progress bar is running.
// Call display() to show the progress bar again.
func (p *Progress) Destroy() {
	if p.Overwrite {
		p.render("")
	}
}

/*************************************************************
 * helper methods
 *************************************************************/

// render progress bar to terminal
func (p *Progress) render(text string) {
	if p.Overwrite {
		// first run. create new line
		if p.firstRun && p.Newline {
			fmt.Println()
			p.firstRun = false

			// delete prev rendered line.
		} else {
			// \x0D - Move the cursor to the beginning of the line
			// \x1B[2K - Erase(Delete) the line
			fmt.Print("\x0D\x1B[2K")
		}

		color.Print(text)
	} else if p.step > 0 {
		color.Println(text)
	}
}

func (p *Progress) checkStart() {
	if !p.started {
		panic("Progress bar has not yet been started.")
	}
}

// build widgets form Format string.
func (p *Progress) buildLine() string {
	return widgetMatch.ReplaceAllStringFunc(p.Format, func(s string) string {
		var text string
		// {@current} -> current
		// {@percent:3s} -> percent:3s
		name := strings.Trim(s, "{@}")
		fmtArg := ""

		// percent:3s
		if pos := strings.IndexRune(name, ':'); pos > 0 {
			fmtArg = name[pos+1:]
			name = name[0:pos]
		}

		if handler, ok := p.Widgets[name]; ok {
			text = handler(p)
		} else if msg, ok := p.Messages[name]; ok {
			text = msg
		} else {
			return s
		}

		// like {@percent:3s} "7%" -> "  7%"
		if fmtArg != "" {
			text = fmt.Sprintf("%"+fmtArg, text)
		}
		// fmt.Println("info:", arg, name, ", text:", text)
		return text
	})
}

// Handler get widget handler by widget name
func (p *Progress) Handler(name string) WidgetFunc {
	if handler, ok := p.Widgets[name]; ok {
		return handler
	}

	return nil
}

/*************************************************************
 * getter methods
 *************************************************************/

// Percent gets the current percent
func (p *Progress) Percent() float32 {
	return p.percent
}

// Step gets the current step position.
func (p *Progress) Step() uint {
	return p.step
}

// Progress alias of the Step()
func (p *Progress) Progress() uint {
	return p.step
}

// StartedAt time get
func (p *Progress) StartedAt() time.Time {
	return p.startedAt
}

// FinishedAt time get
func (p *Progress) FinishedAt() time.Time {
	return p.finishedAt
}
