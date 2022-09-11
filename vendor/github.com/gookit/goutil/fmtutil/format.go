package fmtutil

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// data size
const (
	OneKByte = 1024
	OneMByte = 1024 * 1024
	OneGByte = 1024 * 1024
)

// DataSize format bytes number friendly.
//
// Usage:
// 	file, err := os.Open(path)
// 	fl, err := file.Stat()
// 	fmtSize := DataSize(fl.Size())
func DataSize(bytes uint64) string {
	switch {
	case bytes < 1024:
		return fmt.Sprintf("%dB", bytes)
	case bytes < 1024*1024:
		return fmt.Sprintf("%.2fK", float64(bytes)/1024)
	case bytes < 1024*1024*1024:
		return fmt.Sprintf("%.2fM", float64(bytes)/1024/1024)
	default:
		return fmt.Sprintf("%.2fG", float64(bytes)/1024/1024/1024)
	}
}

// PrettyJSON get pretty Json string
func PrettyJSON(v interface{}) (string, error) {
	out, err := json.MarshalIndent(v, "", "    ")
	return string(out), err
}

// StringsToInts string slice to int slice. alias of the arrutil.StringsToInts()
func StringsToInts(ss []string) (ints []int, err error) {
	for _, str := range ss {
		iVal, err := strconv.Atoi(str)
		if err != nil {
			return []int{}, err
		}

		ints = append(ints, iVal)
	}
	return
}

// ArgsWithSpaces it like Println, will add spaces for each argument
func ArgsWithSpaces(args []interface{}) (message string) {
	if ln := len(args); ln == 0 {
		message = ""
	} else if ln == 1 {
		message = fmt.Sprint(args[0])
	} else {
		message = fmt.Sprintln(args...)
		// clear last "\n"
		message = message[:len(message)-1]
	}
	return
}
