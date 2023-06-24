package comfunc

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Workdir get
func Workdir() string {
	dir, _ := os.Getwd()
	return dir
}

// ExpandHome will parse first `~` as user home dir path.
func ExpandHome(pathStr string) string {
	if len(pathStr) == 0 {
		return pathStr
	}

	if pathStr[0] != '~' {
		return pathStr
	}

	if len(pathStr) > 1 && pathStr[1] != '/' && pathStr[1] != '\\' {
		return pathStr
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return pathStr
	}
	return homeDir + pathStr[1:]
}

// ExecCmd an command and return output.
//
// Usage:
//
//	ExecCmd("ls", []string{"-al"})
func ExecCmd(binName string, args []string, workDir ...string) (string, error) {
	// create a new Cmd instance
	cmd := exec.Command(binName, args...)
	if len(workDir) > 0 {
		cmd.Dir = workDir[0]
	}

	bs, err := cmd.Output()
	return string(bs), err
}

// ShellExec exec command by shell
// cmdLine e.g. "ls -al"
func ShellExec(cmdLine string, shells ...string) (string, error) {
	// shell := "/bin/sh"
	shell := "sh"
	if len(shells) > 0 {
		shell = shells[0]
	}

	var out bytes.Buffer

	cmd := exec.Command(shell, "-c", cmdLine)
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return "", err
	}
	return out.String(), nil
}

// curShell cache
var curShell string

// CurrentShell get current used shell env file.
//
// eg "/bin/zsh" "/bin/bash".
// if onlyName=true, will return "zsh", "bash"
func CurrentShell(onlyName bool) (binPath string) {
	var err error
	if curShell == "" {
		binPath = os.Getenv("SHELL")
		if len(binPath) == 0 {
			binPath, err = ShellExec("echo $SHELL")
			if err != nil {
				return ""
			}
		}

		binPath = strings.TrimSpace(binPath)
		// cache result
		curShell = binPath
	} else {
		binPath = curShell
	}

	if onlyName && len(binPath) > 0 {
		binPath = filepath.Base(binPath)
	}
	return
}

// HasShellEnv has shell env check.
//
// Usage:
//
//	HasShellEnv("sh")
//	HasShellEnv("bash")
func HasShellEnv(shell string) bool {
	// can also use: "echo $0"
	out, err := ShellExec("echo OK", shell)
	if err != nil {
		return false
	}
	return strings.TrimSpace(out) == "OK"
}
