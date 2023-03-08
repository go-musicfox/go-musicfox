package process

import (
	"os"

	"github.com/gookit/goutil/fsutil"
	"github.com/gookit/goutil/mathutil"
	"github.com/gookit/goutil/strutil"
)

// PidFile struct
type PidFile struct {
	pid  int
	file string
	// body string
}

// NewPidFile instance
func NewPidFile(file string) *PidFile {
	return &PidFile{
		file: file,
	}
}

// Exists of th pid file
func (pf *PidFile) Exists() bool {
	return fsutil.FileExist(pf.file)
}

// File path
func (pf *PidFile) File() string {
	return pf.file
}

// PID value
func (pf *PidFile) PID() int {
	if pf.pid > 0 {
		return pf.pid
	}

	if fsutil.FileExist(pf.file) {
		bts, err := os.ReadFile(pf.file)
		if err == nil {
			pf.pid = strutil.QuietInt(string(bts))
		}
	}

	return pf.pid
}

// String PID value string
func (pf *PidFile) String() string {
	return mathutil.String(pf.pid)
}

// SetPID value
func (pf *PidFile) SetPID(val int) int {
	pf.pid = val
	return pf.pid
}

// Save PID value to file
func (pf *PidFile) Save() error {
	if pf.pid < 1 {
		return nil
	}

	_, err := fsutil.PutContents(pf.file, pf.String())
	return err
}
