package gcli

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/gookit/color"
	"github.com/gookit/gcli/v2/helper"
	"github.com/gookit/goutil/strutil"
)

// Command a CLI command structure
type Command struct {
	// core is internal use
	core
	// cmdLine is internal use
	// *cmdLine
	// HelpVars
	// // Hooks can allow setting some hooks func on running.
	// Hooks // allowed hooks: "init", "before", "after", "error"

	// Name is the full command name.
	Name string
	// module is the name for grouped commands
	// subName is the name for grouped commands
	// eg: "sys:info" -> module: "sys", subName: "info"
	module, subName string
	// UseFor is the command description message.
	UseFor string
	// Aliases is the command name's alias names
	Aliases []string
	// Config func, will call on `initialize`.
	// - you can config options and other init works
	Config func(c *Command)
	// Flags(command options) is a set of flags specific to this command.
	// Flags flag.FlagSet
	// Examples some usage example display
	Examples string
	// Func is the command handler func. Func Runner
	Func CmdFunc
	// Help is the long help message text
	Help string
	// HelpRender custom render cmd help message
	HelpRender func(c *Command)

	// CustomFlags indicates that the command will do its own flag parsing.
	CustomFlags bool
	// Arguments for the command
	Arguments
	// Flags options for the command.
	Flags

	// application
	app *App
	// mark is alone running.
	alone bool
	// mark is disabled. if true will skip register to cli-app.
	disabled bool
	// all option names of the command
	// optNames map[string]string
}

// NewCommand create a new command instance.
// Usage:
// 	cmd := NewCommand("my-cmd", "description")
//	// OR with an config func
// 	cmd := NewCommand("my-cmd", "description", func(c *Command) { ... })
// 	app.Add(cmd) // OR cmd.AttachTo(app)
func NewCommand(name, useFor string, fn ...func(c *Command)) *Command {
	c := &Command{
		Name:   name,
		UseFor: useFor,
	}

	// has config func
	if len(fn) > 0 {
		c.Config = fn[0]
	}

	// set name
	c.Arguments.SetName(name)
	return c
}

// SetFunc Settings command handler func
func (c *Command) SetFunc(fn CmdFunc) *Command {
	c.Func = fn
	return c
}

// AttachTo attach the command to CLI application
func (c *Command) AttachTo(app *App) {
	app.AddCommand(c)
}

// Disable set cmd is disabled
func (c *Command) Disable() {
	c.disabled = true
}

// IsDisabled get cmd is disabled
func (c *Command) IsDisabled() bool {
	return c.disabled
}

// Runnable reports whether the command can be run; otherwise
// it is a documentation pseudo-command such as import path.
func (c *Command) Runnable() bool {
	return c.Func != nil
}

// initialize works for the command
func (c *Command) initialize() *Command {
	c.core.cmdLine = CLI

	// init for cmd Arguments
	c.Arguments.SetName(c.Name)
	c.Arguments.SetValidateNum(!c.alone && gOpts.strictMode)

	// init for cmd Flags
	c.Flags.InitFlagSet(c.Name)
	c.Flags.FSet().SetOutput(c.Flags.out)
	c.Flags.FSet().Usage = func() { // call on exists "-h" "--help"
		Logf(VerbDebug, "render help message on exists '-h|--help' or has unknown flag")
		c.ShowHelp()
	}

	// format description
	if len(c.UseFor) > 0 {
		c.UseFor = strutil.UpperFirst(c.UseFor)

		// contains help var "{$cmd}". replace on here is for 'app help'
		if strings.Contains(c.UseFor, "{$cmd}") {
			c.UseFor = strings.Replace(c.UseFor, "{$cmd}", c.Name, -1)
		}
	}

	// call Config func
	if c.Config != nil {
		c.Config(c)
	}

	// set help vars
	// c.Vars = c.app.vars // Error: var is map, map is ref addr
	c.AddVars(c.core.innerHelpVars())
	c.AddVars(map[string]string{
		"cmd": c.Name,
		// full command
		"fullCmd": c.binName + " " + c.Name,
	})

	c.Fire(EvtCmdInit, nil)

	return c
}

// IsAlone running
func (c *Command) IsAlone() bool {
	return c.alone
}

// NotAlone running
func (c *Command) NotAlone() bool {
	return !c.alone
}

// Module name of the grouped command
func (c *Command) Module() string {
	return c.module
}

// SubName name of the grouped command
func (c *Command) SubName() string {
	return c.subName
}

// ID get command ID name.
func (c *Command) goodName() string {
	name := strings.Trim(strings.TrimSpace(c.Name), ": ")
	if name == "" {
		panicf("the command name can not be empty")
	}

	if !goodCmdName.MatchString(name) {
		panicf("the command name '%s' is invalid, must match: %s", name, regGoodCmdName)
	}

	// update name
	c.Name = name
	return name
}

/*************************************************************
 * command run
 *************************************************************/

// do parse option flags, remaining is cmd args
func (c *Command) parseFlags(args []string) (ss []string, err error) {
	// strict format options
	if gOpts.strictMode && len(args) > 0 {
		args = strictFormatArgs(args)
	}

	// fix and compatible
	args = moveArgumentsToEnd(args)

	Logf(VerbDebug, "flags on after format: %v", args)

	// NOTICE: disable output internal error message on parse flags
	// c.FSet().SetOutput(ioutil.Discard)

	// parse options, don't contains command name.
	if err = c.Parse(args); err != nil {
		return
	}

	return c.Flags.RawArgs(), nil
}

// prepare: before execute the command
func (c *Command) prepare(_ []string) (status int, err error) {
	return
}

// do execute the command
func (c *Command) execute(args []string) (err error) {
	c.Fire(EvtCmdBefore, args)

	// collect and binding named args
	if err := c.ParseArgs(args); err != nil {
		c.Fire(EvtCmdError, err)
		return err
	}

	// call command handler func
	if c.Func == nil {
		Logf(VerbWarn, "the command '%s' no handler func to running", c.Name)
	} else {
		// err := c.Func.Run(c, args)
		err = c.Func(c, args)
	}

	if err != nil {
		c.Fire(EvtCmdError, err)
	} else {
		c.Fire(EvtCmdAfter, nil)
	}
	return
}

// Fire event handler by name
func (c *Command) Fire(event string, data interface{}) {
	Logf(VerbDebug, "command '%s' trigger the event: <mga>%s</>", c.Name, event)

	c.Hooks.Fire(event, c, data)
}

// On add hook handler for a hook event
func (c *Command) On(name string, handler HookFunc) {
	Logf(VerbDebug, "command '%s' add hook: %s", c.Name, name)

	c.Hooks.On(name, handler)
}

// Copy a new command for current
func (c *Command) Copy() *Command {
	nc := *c
	// reset some fields
	nc.Func = nil
	nc.Hooks.ClearHooks()
	// nc.Flags = flag.FlagSet{}

	return &nc
}

/*************************************************************
 * alone running
 *************************************************************/

var errCallRun = errors.New("this method can only be called in standalone mode")

// MustRun Alone the current command, will panic on error
func (c *Command) MustRun(inArgs []string) {
	if err := c.Run(inArgs); err != nil {
		color.Error.Println("Run command error: %s", err.Error())
		panic(err)
	}
}

// Run Alone the current command
func (c *Command) Run(inArgs []string) (err error) {
	// - Running in application.
	if c.app != nil {
		return errCallRun
	}

	// - Alone running command

	// mark is alone
	c.alone = true
	// only init global flags on alone run.
	c.core.gFlags = NewFlags(c.Name + ".GlobalOpts").WithOption(FlagsOption{
		Alignment: AlignLeft,
	})

	// binding global options
	bindingCommonGOpts(c.gFlags)

	// TODO parse global options

	// init the command
	c.initialize()

	// add default error handler.
	c.Hooks.AddOn(EvtCmdError, defaultErrHandler)

	// check input args
	if len(inArgs) == 0 {
		inArgs = os.Args[1:]
	}

	// if Command.CustomFlags=true, will not run Flags.Parse()
	if !c.CustomFlags {
		// contains keywords "-h" OR "--help" on end
		if c.hasHelpKeywords() {
			c.ShowHelp()
			return
		}

		// if CustomFlags=true, will not run Flags.Parse()
		inArgs, err = c.parseFlags(inArgs)
		if err != nil {
			// ignore flag.ErrHelp error
			if err == flag.ErrHelp {
				err = nil
			}
			return
		}
	}

	return c.execute(inArgs)
}

/*************************************************************
 * command help
 *************************************************************/

// CmdHelpTemplate help template for a command
var CmdHelpTemplate = `{{.UseFor}}
{{if .Cmd.NotAlone}}
<comment>Name:</> {{.Cmd.Name}}{{if .Cmd.Aliases}} (alias: <info>{{.Cmd.AliasesString}}</>){{end}}{{end}}
<comment>Usage:</> {$binName} [Global Options...] {{if .Cmd.NotAlone}}<info>{{.Cmd.Name}}</> {{end}}[--option ...] [arguments ...]

<comment>Global Options:</>
{{.GOpts}}{{if .Options}}
<comment>Options:</>
{{.Options}}{{end}}{{if .Cmd.Args}}
<comment>Arguments:</>{{range $a := .Cmd.Args}}
  <info>{{$a.HelpName | printf "%-12s"}}</>{{$a.Desc | ucFirst}}{{if $a.Required}}<red>*</>{{end}}{{end}}
{{end}}{{if .Cmd.Examples}}
<comment>Examples:</>
{{.Cmd.Examples}}{{end}}{{if .Cmd.Help}}
<comment>Help:</>
{{.Cmd.Help}}{{end}}`

// ShowHelp show command help info
func (c *Command) ShowHelp() {
	// custom help render func
	if c.HelpRender != nil {
		c.HelpRender(c)
		return
	}

	// clear space and empty new line
	if c.Examples != "" {
		c.Examples = strings.Trim(c.Examples, "\n") + "\n"
	}

	// clear space and empty new line
	if c.Help != "" {
		c.Help = strings.Join([]string{strings.TrimSpace(c.Help), "\n"}, "")
	}

	// render help message
	s := helper.RenderText(CmdHelpTemplate, map[string]interface{}{
		"Cmd": c,
		// global options
		"GOpts": c.gFlags.String(),
		// parse options to string
		"Options": c.Flags.String(),
		// always upper first char
		"UseFor": c.UseFor,
	}, nil)

	// parse help vars then print help
	color.Print(c.ReplaceVars(s))
	// fmt.Printf("%#v\n", s)
}

/*************************************************************
 * helper methods
 *************************************************************/

// App returns the CLI application
func (c *Command) App() *App {
	return c.app
}

// Errorf format message and add error to the command
func (c *Command) Errorf(format string, v ...interface{}) error {
	return fmt.Errorf(format, v...)
}

// AliasesString returns aliases string
func (c *Command) AliasesString(sep ...string) string {
	s := ","
	if len(sep) == 1 {
		s = sep[0]
	}

	return strings.Join(c.Aliases, s)
}

// Logf print log message
func (c *Command) Logf(level uint, format string, v ...interface{}) {
	Logf(level, format, v...)
}
