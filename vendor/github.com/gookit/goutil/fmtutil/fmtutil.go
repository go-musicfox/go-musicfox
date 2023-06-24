// Package fmtutil provide some format util functions.
package fmtutil

import (
	"encoding/json"

	"github.com/gookit/goutil/strutil"
)

// StringOrJSON encode pretty JSON data to json bytes.
func StringOrJSON(v any) ([]byte, error) {
	s, err := strutil.StringOrErr(v)
	if err != nil {
		return json.MarshalIndent(v, "", "    ")
	}
	return []byte(s), nil
}
