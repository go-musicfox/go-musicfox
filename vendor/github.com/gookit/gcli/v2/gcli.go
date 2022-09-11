// Package gcli is a simple to use command line application and tool library.
//
// Contains: cli app, flags parse, interact, progress, data show tools.
//
// Source code and other details for the project are available at GitHub:
// 		https://github.com/gookit/gcli
//
// Usage please refer examples and see README
package gcli

import (
	"os"
	"regexp"
	"runtime"
	"strings"

	"github.com/gookit/goutil/envutil"
)

const (
	// OK success exit code
	OK = 0
	// ERR error exit code
	ERR = 2
	// GOON prepare run successful, goon run command
	GOON = -1
	// HelpCommand name
	HelpCommand = "help"
	// match an good option, argument name
	regGoodName = `^[a-zA-Z][\w-]*$`
	// match an good command name
	regGoodCmdName = `^[a-zA-Z][\w:-]*$`
)

var (
	// stdApp store default application instance
	stdApp *App
	// an empty argument
	emptyArg = &Argument{}
	// good name for option and argument
	goodName = regexp.MustCompile(regGoodName)
	// match an good command name
	goodCmdName = regexp.MustCompile(regGoodCmdName)

	// global options
	gOpts = &GlobalOpts{
		strictMode: true,
		verbose:    VerbError,
	}

	// CLI create an default instance
	CLI = &cmdLine{
		pid: os.Getpid(),
		// more info
		osName:  runtime.GOOS,
		binName: os.Args[0],
		argLine: strings.Join(os.Args[1:], " "),
	}
)

// init
func init() {
	// don't display date on print log
	// log.SetFlags(0)
	workDir, _ := os.Getwd()
	CLI.workDir = workDir

	// binName will contains work dir path on windows
	if envutil.IsWin() {
		CLI.binName = strings.Replace(CLI.binName, workDir+"\\", "", 1)
	}
}

// InitStdApp create the default cli app.
func InitStdApp(fn ...func(a *App)) *App {
	stdApp = NewApp(fn...)
	return stdApp
}

// StdApp get the default std app
func StdApp() *App {
	return stdApp
}

// GOpts get the global options
func GOpts() *GlobalOpts {
	return gOpts
}

// Verbose returns verbose level
func Verbose() uint {
	return gOpts.verbose
}

// SetDebugMode level
func SetDebugMode() {
	gOpts.verbose = VerbDebug
}

// SetQuietMode level
func SetQuietMode() {
	gOpts.verbose = VerbQuiet
}

// SetVerbose level
func SetVerbose(verbose uint) {
	gOpts.verbose = verbose
}

// StrictMode get is strict mode
func StrictMode() bool {
	return gOpts.strictMode
}

// SetStrictMode for parse flags
func SetStrictMode(strict bool) {
	gOpts.strictMode = strict
}

func bindingCommonGOpts(fs *Flags) {
	// binding global options
	fs.UintOpt(&gOpts.verbose, "verbose", "", gOpts.verbose, "Set error reporting level(quiet 0 - 4 debug)")
	fs.BoolOpt(&gOpts.showHelp, "help", "h", false, "Display the help information")
	fs.BoolOpt(&gOpts.NoColor, "no-color", "", gOpts.NoColor, "Disable color when outputting message")
	fs.BoolOpt(&gOpts.noProgress, "no-progress", "", gOpts.noProgress, "Disable display progress message")
	fs.BoolOpt(&gOpts.noInteractive, "no-interactive", "", gOpts.noInteractive, "Disable interactive confirmation operations")
}
