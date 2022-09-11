package gcli

import (
	"fmt"
	"strconv"
)

// constants for error level 0 - 4
const (
	VerbQuiet uint = iota // don't report anything
	VerbError             // reporting on error
	VerbWarn
	VerbInfo
	VerbDebug
	VerbCrazy
)

// constants for hooks event, there are default allowed event names
const (
	EvtAppInit   = "app.init"
	EvtAppBefore = "app.run.before"
	EvtAppAfter  = "app.run.after"
	EvtAppError  = "app.run.error"

	EvtCmdInit   = "cmd.init"
	EvtCmdBefore = "cmd.run.before"
	EvtCmdAfter  = "cmd.run.after"
	EvtCmdError  = "cmd.run.error"

	EvtAppPrepareAfter = "app.prepare.after"
	// EvtStop   = "stop"
)

// GlobalOpts global flags
type GlobalOpts struct {
	verbose  uint // message report level
	NoColor  bool
	showVer  bool
	showHelp bool
	// dont display progress
	noProgress bool
	// close interactive confirm
	noInteractive bool
	// StrictMode use strict mode for parse flags
	// If True(default):
	// 	- short opt must be begin "-", long opt must be begin "--"
	//	- will convert like "-ab" to "-a -b"
	// 	- will check invalid arguments, like to many arguments
	strictMode bool
	// command auto completion mode.
	// eg "./cli --cmd-completion [COMMAND --OPT ARG]"
	inCompletion bool
}

// Runner interface
type Runner interface {
	// Config(c *Command)
	Run(c *Command, args []string) error
}

// CmdFunc definition
type CmdFunc func(c *Command, args []string) error

// Run implement the Runner interface
func (f CmdFunc) Run(c *Command, args []string) error {
	return f(c, args)
}

// Commander interface definition
type Commander interface {
	// Creator for create new command
	Creator() *Command
	// Prepare bind Flags or Arguments for the command
	Prepare(c *Command)
	// Execute(c *Command, args []string) error
	Run(c *Command, args []string) error
}

// Executor definition
type Executor interface {
}

// Executor definition
type RunningAble struct {
}

/*************************************************************************
 * options: some special flag vars
 * - implemented flag.Value interface
 *************************************************************************/

// Ints The int flag list, implemented flag.Value interface
type Ints []int

// String to string
func (s *Ints) String() string {
	return fmt.Sprintf("%v", *s)
}

// Set new value
func (s *Ints) Set(value string) error {
	intVal, err := strconv.Atoi(value)
	if err == nil {
		*s = append(*s, intVal)
	}

	return err
}

// Strings The string flag list, implemented flag.Value interface
type Strings []string

// String to string
func (s *Strings) String() string {
	return fmt.Sprintf("%v", *s)
}

// Set new value
func (s *Strings) Set(value string) error {
	*s = append(*s, value)
	return nil
}

// Booleans The bool flag list, implemented flag.Value interface
type Booleans []bool

// String to string
func (s *Booleans) String() string {
	return fmt.Sprintf("%v", *s)
}

// Set new value
func (s *Booleans) Set(value string) error {
	boolVal, err := strconv.ParseBool(value)
	if err == nil {
		*s = append(*s, boolVal)
	}

	return err
}
