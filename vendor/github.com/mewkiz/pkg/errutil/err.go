// Package errutil implements some error utility functions.
package errutil

import (
	"errors"
	"fmt"
	"path"
	"runtime"

	"github.com/mewkiz/pkg/term"
)

// UseColor indicates if error messages should use colors.
var UseColor = true

// ErrInfo is en error containing position information.
type ErrInfo struct {
	// err is the original error message.
	Err error
	// pos refers to the position of the original error message. A nil value
	// indicates that no position information should be displayed with the error
	// message.
	pos *position
}

// position includes information about file name, line number and callee.
type position struct {
	// base file name.
	file string
	// line number.
	line int
	// callee function name.
	callee string
}

func (pos *position) String() string {
	if pos == nil {
		return "<no position>"
	}
	filePos := fmt.Sprintf("(%s:%d):", pos.file, pos.line)

	if UseColor {
		// Use colors.
		filePosColor := term.WhiteBold(filePos)
		if pos.callee == "" {
			return filePosColor
		}
		return fmt.Sprintf("%s %s", term.MagentaBold(pos.callee), filePosColor)
	}

	// No colors.
	if pos.callee == "" {
		return filePos
	}
	return fmt.Sprintf("%s %s", pos.callee, filePos)
}

// New returns an error which contains position information from the callee.
func New(text string) (err error) {
	return backendErr(errors.New(text))
}

// Newf returns a formatted error which contains position information from the
// callee.
func Newf(format string, a ...interface{}) (err error) {
	return backendErr(fmt.Errorf(format, a...))
}

// NewNoPos returns an error which explicitly contains no position information.
// Further calls to Err will not embed any position information.
func NewNoPos(text string) (err error) {
	return &ErrInfo{Err: errors.New(text)}
}

// NewNoPosf returns a formatted error which explicitly contains no position information.
// Further calls to Err will not embed any position information.
func NewNoPosf(format string, a ...interface{}) (err error) {
	return &ErrInfo{Err: fmt.Errorf(format, a...)}
}

// ErrNoPos return an error which explicitly contains no position information.
func ErrNoPos(e error) (err error) {
	return &ErrInfo{Err: e}
}

// Err returns an error which contains position information from the callee. The
// original position information is left unaltered if available.
func Err(e error) (err error) {
	return backendErr(e)
}

func backendErr(e error) (err error) {
	_, ok := e.(*ErrInfo)
	if ok {
		return e
	}
	pc, file, line, ok := runtime.Caller(2)
	if !ok {
		return e
	}
	var callee string
	f := runtime.FuncForPC(pc)
	if f != nil {
		callee = f.Name()
	}
	err = &ErrInfo{
		Err: e,
		pos: &position{
			file:   path.Base(file),
			line:   line,
			callee: callee,
		},
	}
	return err
}

// Error returns an error string with position information.
//
// The error format is as follows:
//    pkg.func (file:line): error: text
func (e *ErrInfo) Error() string {
	text := "<nil>"
	if e.Err != nil {
		text = e.Err.Error()
	}

	if UseColor {
		// Use colors.
		prefix := term.RedBold("error:")
		if e.pos == nil {
			return fmt.Sprintf("%s %s", prefix, text)
		}
		return fmt.Sprintf("%s %s %s", e.pos, prefix, text)
	}

	// No colors.
	if e.pos == nil {
		return text
	}
	return fmt.Sprintf("%s %s", e.pos, text)

}
