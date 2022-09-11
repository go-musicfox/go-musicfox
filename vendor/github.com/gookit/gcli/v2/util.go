package gcli

import (
	"fmt"
	"path"
	"runtime"
	"strings"

	"github.com/gookit/color"
	"github.com/gookit/goutil/strutil"
)

/*************************************************************
 * console log
 *************************************************************/

var level2name = map[uint]string{
	VerbError: "ERROR",
	VerbWarn:  "WARN",
	VerbInfo:  "INFO",
	VerbDebug: "DEBUG",
	VerbCrazy: "CRAZY",
}

var level2color = map[uint]color.Color{
	VerbError: color.FgRed,
	VerbWarn:  color.FgYellow,
	VerbInfo:  color.FgGreen,
	VerbDebug: color.FgCyan,
	VerbCrazy: color.FgMagenta,
}

// Logf print log message
func Logf(level uint, format string, v ...interface{}) {
	if gOpts.verbose < level {
		return
	}

	var fnName string
	name, has := level2name[level]
	if !has {
		name, level = "CRAZY", VerbCrazy
	}

	pc, fName, line, ok := runtime.Caller(1)
	if !ok {
		fnName, fName, line = "UNKNOWN", "???.go", 0
	} else {
		fName = path.Base(fName)
		fnName = runtime.FuncForPC(pc).Name()
	}

	name = level2color[level].Render(name)
	color.Printf("GCli: [%s] [%s() At %s:%d] %s\n", name, fnName, fName, line, fmt.Sprintf(format, v...))
}

func defaultErrHandler(data ...interface{}) {
	if len(data) == 2 && data[1] != nil {
		if err, ok := data[1].(error); ok {
			color.Error.Tips(err.Error())
			// fmt.Println(color.Red.Render("ERROR:"), err.Error())
		}
	}
}

/*************************************************************
 * some helper methods
 *************************************************************/

// Print messages
func Print(args ...interface{}) {
	color.Print(args...)
}

// Println messages
func Println(args ...interface{}) {
	color.Println(args...)
}

// Printf messages
func Printf(format string, args ...interface{}) {
	color.Printf(format, args...)
}

func panicf(format string, v ...interface{}) {
	panic(fmt.Sprintf("GCli: "+format, v...))
}

// func exitWithMsg(format string, v ...interface{}) {
// 	fmt.Printf(format, v...)
// 	Exit(0)
// }

// strictFormatArgs '-ab' will split to '-a -b', '--o' -> '-o'
func strictFormatArgs(args []string) (fmtArgs []string) {
	if len(args) == 0 {
		return args
	}

	for _, arg := range args {
		// eg: --a ---name
		if strings.Index(arg, "--") == 0 {
			farg := strings.TrimLeft(arg, "-")
			if rl := len(farg); rl == 1 { // fix: "--a" -> "-a"
				arg = "-" + farg
			} else if rl > 1 { // fix: "---name" -> "--name"
				arg = "--" + farg
			}
			// TODO No change remain OR remove like "--" "---"
			// maybe ...

		} else if strings.IndexByte(arg, '-') == 0 {
			ln := len(arg)
			// fix: "-abc" -> "-a -b -c"
			if ln > 2 {
				chars := strings.Split(strings.Trim(arg, "-"), "")

				for _, s := range chars {
					fmtArgs = append(fmtArgs, "-"+s)
				}
				continue
			}
		}

		fmtArgs = append(fmtArgs, arg)
	}

	return fmtArgs
}

// flags parser is flag#FlagSet.Parse(), so:
// - if args like: "arg0 arg1 --opt", will parse fail
// - if args convert to: "--opt arg0 arg1", can correctly parse
func moveArgumentsToEnd(args []string) []string {
	if len(args) < 2 {
		return args
	}

	var argEnd int
	for i, arg := range args {
		// strop on the first option
		if strings.IndexByte(arg, '-') == 0 {
			argEnd = i
			break
		}
	}

	// the first is an option
	if argEnd == -1 {
		return args
	}

	return append(args[argEnd:], args[0:argEnd]...)
}

// split "ef" to ["e", "f"]
func splitShortStr(str string) (ss []string) {
	bs := []byte(str)

	for _, b := range bs {
		if strutil.IsAlphabet(b) {
			ss = append(ss, string(b))
		}
	}
	return
}

func shorts2str(ss []string) string {
	var newSs []string
	for _, s := range ss {
		newSs = append(newSs, "-"+s)
	}

	// eg: "-t, -o"
	return strings.Join(newSs, ", ")
}
