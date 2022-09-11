package gcli

import (
	"fmt"
	"os"
	"strings"
)

// core definition
type core struct {
	*cmdLine
	HelpVars
	Hooks // allowed hooks: "init", "before", "after", "error"
	// global options flag set
	gFlags *Flags
	// GOptsBinder you can custom binding global options
	GOptsBinder func(gf *Flags)
}

// GlobalFlags get the app GlobalFlags
func (c core) GlobalFlags() *Flags {
	return c.gFlags
}

// common basic help vars
func (c core) innerHelpVars() map[string]string {
	return map[string]string{
		"pid":     fmt.Sprint(CLI.pid),
		"workDir": CLI.workDir,
		"binName": CLI.binName,
	}
}

/*************************************************************
 * simple events manage
 *************************************************************/

// Hooks struct
type Hooks struct {
	// Hooks can setting some hooks func on running.
	hooks map[string]HookFunc
}

// On register event hook by name
func (h *Hooks) On(name string, handler HookFunc) {
	if handler != nil {
		if h.hooks == nil {
			h.hooks = make(map[string]HookFunc)
		}

		h.hooks[name] = handler
	}
}

// AddOn register on not exists hook.
func (h *Hooks) AddOn(name string, handler HookFunc) {
	if _, ok := h.hooks[name]; !ok {
		h.On(name, handler)
	}
}

// Fire event by name, allow with event data
func (h *Hooks) Fire(event string, data ...interface{}) {
	if handler, ok := h.hooks[event]; ok {
		handler(data...)
	}
}

// ClearHooks clear hooks data
func (h *Hooks) ClearHooks() {
	h.hooks = nil
}

/*************************************************************
 * Command Line: command data
 *************************************************************/

// cmdLine store common data for CLI
type cmdLine struct {
	// pid for current application
	pid int
	// os name.
	osName string
	// the CLI app work dir path. by `os.Getwd()`
	workDir string
	// bin script name, by `os.Args[0]`. eg "./cliapp"
	binName string
	// os.Args to string, but no binName.
	argLine string
}

// PID get PID
func (c *cmdLine) PID() int {
	return c.pid
}

// OsName is equals to `runtime.GOOS`
func (c *cmdLine) OsName() string {
	return c.osName
}

// OsArgs is equals to `os.Args`
func (c *cmdLine) OsArgs() []string {
	return os.Args
}

// BinName get bin script name
func (c *cmdLine) BinName() string {
	return c.binName
}

// WorkDir get work dir
func (c *cmdLine) WorkDir() string {
	return c.workDir
}

// ArgLine os.Args to string, but no binName.
func (c *cmdLine) ArgLine() string {
	return c.argLine
}

func (c *cmdLine) hasHelpKeywords() bool {
	return strings.HasSuffix(c.argLine, " -h") || strings.HasSuffix(c.argLine, " --help")
}

/*************************************************************
 * app/cmd help vars
 *************************************************************/

// HelpVarFormat allow var replace on render help info.
// Default support:
// 	"{$binName}" "{$cmd}" "{$fullCmd}" "{$workDir}"
const HelpVarFormat = "{$%s}"

// HelpVars struct. provide string var function for render help template.
type HelpVars struct {
	// varLeft, varRight string
	// varFormat string
	// Vars you can add some vars map for render help info
	Vars map[string]string
}

// AddVar get command name
func (hv *HelpVars) AddVar(name, value string) {
	if hv.Vars == nil {
		hv.Vars = make(map[string]string)
	}

	hv.Vars[name] = value
}

// AddVars add multi tpl vars
func (hv *HelpVars) AddVars(vars map[string]string) {
	for n, v := range vars {
		hv.AddVar(n, v)
	}
}

// GetVar get a help var by name
func (hv *HelpVars) GetVar(name string) string {
	return hv.Vars[name]
}

// GetVars get all tpl vars
func (hv *HelpVars) GetVars() map[string]string {
	return hv.Vars
}

// ReplaceVars replace vars in the input string.
func (hv *HelpVars) ReplaceVars(input string) string {
	// if not use var
	if !strings.Contains(input, "{$") {
		return input
	}

	var ss []string
	for n, v := range hv.Vars {
		ss = append(ss, fmt.Sprintf(HelpVarFormat, n), v)
	}

	return strings.NewReplacer(ss...).Replace(input)
}
