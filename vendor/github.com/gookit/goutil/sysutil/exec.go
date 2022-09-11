package sysutil

import (
	"bytes"
	"os/exec"

	"github.com/gookit/goutil/cliutil/cmdline"
)

// QuickExec quick exec an simple command line
func QuickExec(cmdLine string, workDir ...string) (string, error) {
	return ExecLine(cmdLine, workDir...)
}

// ExecLine quick exec an command line string
func ExecLine(cmdLine string, workDir ...string) (string, error) {
	p := cmdline.NewParser(cmdLine)

	// create a new Cmd instance
	cmd := p.NewExecCmd()
	if len(workDir) > 0 {
		cmd.Dir = workDir[0]
	}

	bs, err := cmd.Output()
	return string(bs), err
}

// ExecCmd an command and return output.
// Usage:
// 	ExecCmd("ls", []string{"-al"})
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
// cmdStr eg. "ls -al"
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

// FindExecutable in the system
//
// Usage:
// 	FindExecutable("bash")
func FindExecutable(binName string) (string, error) {
	return exec.LookPath(binName)
}

// Executable find in the system
//
// Usage:
// 	Executable("bash")
func Executable(binName string) (string, error) {
	return exec.LookPath(binName)
}

// HasExecutable in the system
//
// Usage:
// 	HasExecutable("bash")
func HasExecutable(binName string) bool {
	_, err := exec.LookPath(binName)
	return err == nil
}
