package envutil

import (
	"io"
	"os"
	"runtime"
	"strings"

	"github.com/gookit/goutil/internal/comfunc"
	"github.com/gookit/goutil/sysutil"
	"golang.org/x/term"
)

// IsWin system. linux windows darwin
func IsWin() bool {
	return runtime.GOOS == "windows"
}

// IsWindows system. alias of IsWin
func IsWindows() bool {
	return runtime.GOOS == "windows"
}

// IsMac system
func IsMac() bool {
	return runtime.GOOS == "darwin"
}

// IsLinux system
func IsLinux() bool {
	return runtime.GOOS == "linux"
}

// IsMSys msys(MINGW64) env. alias of the sysutil.IsMSys()
func IsMSys() bool {
	return sysutil.IsMSys()
}

var detectedWSL bool
var detectedWSLContents string

// IsWSL system env
// https://github.com/Microsoft/WSL/issues/423#issuecomment-221627364
func IsWSL() bool {
	if !detectedWSL {
		b := make([]byte, 1024)
		// `cat /proc/version`
		// on mac:
		// 	!not the file!
		// on linux(debian,ubuntu,alpine):
		//	Linux version 4.19.121-linuxkit (root@18b3f92ade35) (gcc version 9.2.0 (Alpine 9.2.0)) #1 SMP Thu Jan 21 15:36:34 UTC 2021
		// on win git bash, conEmu:
		// 	MINGW64_NT-10.0-19042 version 3.1.7-340.x86_64 (@WIN-N0G619FD3UK) (gcc version 9.3.0 (GCC) ) 2020-10-23 13:08 UTC
		// on WSL:
		//  Linux version 4.4.0-19041-Microsoft (Microsoft@Microsoft.com) (gcc version 5.4.0 (GCC) ) #488-Microsoft Mon Sep 01 13:43:00 PST 2020
		f, err := os.Open("/proc/version")
		if err == nil {
			_, _ = f.Read(b) // ignore error
			f.Close()
			detectedWSLContents = string(b)
		}
		detectedWSL = true
	}
	return strings.Contains(detectedWSLContents, "Microsoft")
}

// IsTerminal isatty check
//
// Usage:
//
//	envutil.IsTerminal(os.Stdout.Fd())
func IsTerminal(fd uintptr) bool {
	// return isatty.IsTerminal(fd) // "github.com/mattn/go-isatty"
	return term.IsTerminal(int(fd))
}

// StdIsTerminal os.Stdout is terminal
func StdIsTerminal() bool {
	return IsTerminal(os.Stdout.Fd())
}

// IsConsole check out is console env. alias of the sysutil.IsConsole()
func IsConsole(out io.Writer) bool {
	return sysutil.IsConsole(out)
}

// HasShellEnv has shell env check.
//
// Usage:
//
//	HasShellEnv("sh")
//	HasShellEnv("bash")
func HasShellEnv(shell string) bool {
	return comfunc.HasShellEnv(shell)
}

// Support color:
//
//	"TERM=xterm"
//	"TERM=xterm-vt220"
//	"TERM=xterm-256color"
//	"TERM=screen-256color"
//	"TERM=tmux-256color"
//	"TERM=rxvt-unicode-256color"
//
// Don't support color:
//
//	"TERM=cygwin"
var specialColorTerms = map[string]bool{
	"alacritty": true,
}

// IsSupportColor check current console is support color.
//
// Supported:
//
//	linux, mac, or windows's ConEmu, Cmder, putty, git-bash.exe
//
// Not support:
//
//	windows cmd.exe, powerShell.exe
func IsSupportColor() bool {
	envTerm := os.Getenv("TERM")
	if strings.Contains(envTerm, "xterm") {
		return true
	}

	// it's special color term
	if _, ok := specialColorTerms[envTerm]; ok {
		return true
	}

	// like on ConEmu software, e.g "ConEmuANSI=ON"
	if os.Getenv("ConEmuANSI") == "ON" {
		return true
	}

	// like on ConEmu software, e.g "ANSICON=189x2000 (189x43)"
	if os.Getenv("ANSICON") != "" {
		return true
	}

	// up: if support 256-color, can also support basic color.
	return IsSupport256Color()
}

// IsSupport256Color render
func IsSupport256Color() bool {
	// "TERM=xterm-256color"
	// "TERM=screen-256color"
	// "TERM=tmux-256color"
	// "TERM=rxvt-unicode-256color"
	supported := strings.Contains(os.Getenv("TERM"), "256color")
	if !supported {
		// up: if support true-color, can also support 256-color.
		supported = IsSupportTrueColor()
	}

	return supported
}

// IsSupportTrueColor render. IsSupportRGBColor
func IsSupportTrueColor() bool {
	// "COLORTERM=truecolor"
	return strings.Contains(os.Getenv("COLORTERM"), "truecolor")
}

// IsGithubActions env
func IsGithubActions() bool {
	return os.Getenv("GITHUB_ACTIONS") == "true"
}
