package gcli

import (
	"flag"
	"fmt"
	"strings"
	"text/template"

	"github.com/gookit/color"
	"github.com/gookit/gcli/v2/helper"
	"github.com/gookit/goutil/strutil"
)

// parseGlobalOpts parse global options
func (app *App) parseGlobalOpts() (ok bool) {
	Logf(VerbDebug, "will begin parse global options")
	// global options flag
	// gf := flag.NewFlagSet(app.Args[0], flag.ContinueOnError)
	gf := app.GlobalFlags()

	// parse global options
	err := gf.Parse(app.Args[1:])
	if err != nil {
		color.Error.Tips(err.Error())
		return
	}

	// check global options
	if gOpts.showHelp {
		app.showApplicationHelp()
		return
	}

	if gOpts.showVer {
		app.showVersionInfo()
		return
	}

	// disable color
	if gOpts.NoColor {
		color.Enable = false
	}

	app.rawFlagArgs = gf.FSet().Args()
	Logf(VerbDebug, "console debug is enabled, verbose level is <mgb>%d</>", gOpts.verbose)

	// TODO show auto-completion for bash/zsh
	if gOpts.inCompletion {
		app.showAutoCompletion(app.rawFlagArgs)
		return
	}

	return true
}

// prepare to running, parse args, get command name and command args
func (app *App) prepareRun() (code int) {
	args := app.rawFlagArgs
	// if no input command
	if len(args) == 0 {
		// will try run defaultCommand
		defCmd := app.defaultCommand
		if len(defCmd) == 0 {
			app.showApplicationHelp()
			return
		}

		if !app.IsCommand(defCmd) {
			Logf(VerbError, "The default command '%s' is not exists", defCmd)
			app.showApplicationHelp()
			return
		}

		args = []string{defCmd}
	} else if args[0] == "help" { // is help command
		if len(args) == 1 { // like 'help'
			app.showApplicationHelp()
			return
		}

		// like 'help COMMAND'
		return app.showCommandHelp(args[1:])
	}

	app.rawName = args[0]
	app.cleanArgs = args[1:]
	return GOON
}

// Run running application
func (app *App) Run() (code int) {
	Logf(VerbDebug, "will begin run cli application")

	// ensure application initialized
	if !app.initialized {
		app.initialize()
	}

	// parse global flags
	if !app.parseGlobalOpts() {
		return app.exitIfExitOnEnd(code)
	}

	if code = app.prepareRun(); code != GOON {
		return app.exitIfExitOnEnd(code)
	}

	// trigger event
	app.fireEvent(EvtAppPrepareAfter, app)

	Logf(VerbCrazy, "begin run console application, process ID: %d", app.PID())

	args := app.cleanArgs
	name := app.ResolveName(app.rawName)

	Logf(VerbDebug, "input command: '<cyan>%s</>', real command: '<mga>%s</>', flags: %v", app.rawName, name, args)

	// display unknown input command and similar commands tips
	if !app.IsCommand(name) {
		Logf(VerbDebug, "input the command is not an registered: %s", name)
		app.showCommandTips(name)
		return
	}

	// do run input command
	code = app.doRun(name, args)

	Logf(VerbDebug, "command '%s' run complete, exit with code: %d", name, code)
	return app.exitIfExitOnEnd(code)
}

func (app *App) exitIfExitOnEnd(code int) int {
	if app.ExitOnEnd {
		app.Exit(code)
	}
	return code
}

func (app *App) doRun(name string, args []string) (code int) {
	var err error
	cmd := app.commands[name]

	app.commandName = name
	app.fireEvent(EvtAppBefore, cmd.Copy())

	Logf(VerbDebug, "command '%s' raw flags: %v", name, args)

	// if Command.CustomFlags=true, will not run Flags.Parse()
	if !cmd.CustomFlags {
		// contains keywords "-h" OR "--help" on end
		if CLI.hasHelpKeywords() {
			Logf(VerbDebug, "contains help keywords in flags, render command help message")
			cmd.ShowHelp()
			return
		}

		// parse options, don't contains command name.
		args, err = cmd.parseFlags(args)
		if err != nil {
			// if is flag.ErrHelp error
			if err == flag.ErrHelp {
				return
			}

			color.Error.Tips("Flags parse error - %s", err.Error())
			return ERR
		}
	}

	Logf(VerbDebug, "args on parse end: %v", args)

	// do execute command
	if err := cmd.execute(args); err != nil {
		code = ERR
		app.fireEvent(EvtAppError, err)
	} else {
		app.fireEvent(EvtAppAfter, nil)
	}
	return
}

// Exec running other command in current command
func (app *App) Exec(name string, args []string) (err error) {
	if !app.IsCommand(name) {
		return fmt.Errorf("exec unknown command name '%s'", name)
	}

	cmd := app.commands[name]

	// if Command.CustomFlags=true, will not run Flags.Parse()
	if !cmd.CustomFlags {
		// parse command flags
		args, err = cmd.parseFlags(args)
		if err != nil {
			// ignore flag.ErrHelp error
			// if err != flag.ErrHelp {
			// 	return
			// }
			return
		}
	}

	// do execute command
	return cmd.execute(args)
}

// CommandName get current command name
func (app *App) CommandName() string {
	return app.commandName
}

// CommandNames get all command names
func (app *App) CommandNames() []string {
	var ss []string
	for n := range app.names {
		ss = append(ss, n)
	}

	return ss
}

/*************************************************************
 * display app help
 *************************************************************/

// AppHelpTemplate help template for app(all commands)
var AppHelpTemplate = `{{.Description}} (Version: <info>{{.Version}}</>)
<comment>Usage:</>
  {$binName} [Global Options...] <info>{command}</> [--option ...] [argument ...]

<comment>Global Options:</>
{{.GOpts}}
<comment>Available Commands:</>{{range $module, $cs := .Cs}}{{if $module}}
<comment> {{ $module }}</>{{end}}{{ range $cs }}
  <info>{{.Name | paddingName }}</> {{.UseFor}}{{if .Aliases}} (alias: <cyan>{{ join .Aliases ","}}</>){{end}}{{end}}{{end}}

  <info>{{ paddingName "help" }}</> Display help information

Use "<cyan>{$binName} {COMMAND} -h</>" for more information about a command
`

// display app version info
func (app *App) showVersionInfo() {
	color.Printf(
		"%s\n\nVersion: <cyan>%s</>\n",
		strutil.UpperFirst(app.Description),
		app.Version,
	)

	if app.Logo.Text != "" {
		color.Printf("%s\n", color.WrapTag(app.Logo.Text, app.Logo.Style))
	}
}

// display unknown input command and similar commands tips
func (app *App) showCommandTips(name string) {
	color.Error.Tips(`unknown input command "<mga>%s</>"`, name)
	if ns := app.findSimilarCmd(name); len(ns) > 0 {
		color.Printf("\nMaybe you mean:\n  <green>%s</>\n", strings.Join(ns, ", "))
	}

	color.Printf("\nUse <cyan>%s --help</> to see available commands\n", app.binName)
}

// display app help and list all commands
func (app *App) showApplicationHelp() {
	// cmdHelpTemplate = color.ReplaceTag(cmdHelpTemplate)
	// render help text template
	s := helper.RenderText(AppHelpTemplate, map[string]interface{}{
		"Cs":    app.moduleCommands,
		"GOpts": app.gFlags.String(),
		// app version
		"Version": app.Version,
		// always upper first char
		"Description": strutil.UpperFirst(app.Description),
	}, template.FuncMap{
		"paddingName": func(n string) string {
			return strutil.PadRight(n, " ", app.nameMaxLen)
		},
	})

	// parse help vars and render color tags
	color.Print(app.ReplaceVars(s))
}

// showCommandHelp display help for an command
func (app *App) showCommandHelp(list []string) (code int) {
	binName := app.binName
	if len(list) != 1 {
		color.Error.Tips("Too many arguments given.\n\nUsage: %s help {COMMAND}", binName)
		return ERR
	}

	// get real name
	name := app.ResolveName(list[0])
	if name == HelpCommand || name == "-h" {
		color.Println("Display help message for application or command.\n")
		color.Printf("Usage:\n <cyan>%s {COMMAND} --help</> OR <cyan>%s help {COMMAND}</>\n", binName, binName)
		return
	}

	cmd, exist := app.commands[name]
	if !exist {
		color.Error.Prompt("Unknown command name '%s'. Run '<cyan>%s -h</>' see all commands", name, binName)
		return ERR
	}

	// show help for the command.
	cmd.ShowHelp()
	return
}

// show bash/zsh completion
func (app *App) showAutoCompletion(_ []string) {
	// TODO ...
}

// findSimilarCmd find similar cmd by input string
func (app *App) findSimilarCmd(input string) []string {
	var ss []string
	// ins := strings.Split(input, "")
	// fmt.Print(input, ins)
	ln := len(input)

	names := app.Names()
	names["help"] = 4 // add 'help' command

	// find from command names
	for name := range names {
		cln := len(name)
		if cln > ln && strings.Contains(name, input) {
			ss = append(ss, name)
		} else if ln > cln && strings.Contains(input, name) {
			// sns := strings.Split(str, "")
			ss = append(ss, name)
		}

		// max find 5 items
		if len(ss) == 5 {
			break
		}
	}

	// find from aliases
	for alias := range app.aliases {
		// max find 5 items
		if len(ss) >= 5 {
			break
		}

		cln := len(alias)
		if cln > ln && strings.Contains(alias, input) {
			ss = append(ss, alias)
		} else if ln > cln && strings.Contains(input, alias) {
			ss = append(ss, alias)
		}
	}

	return ss
}
