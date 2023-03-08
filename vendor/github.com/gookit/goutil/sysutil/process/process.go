// Package process Provide some process handle util functions
package process

import "os"

// PID get process ID
func PID() int {
	return os.Getpid()
}
