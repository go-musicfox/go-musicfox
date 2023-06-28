package parser

import (
	"bytes"
	"errors"
	"fmt"
	"sort"

	"github.com/gookit/goutil/timex"
)

// Encode golang data(map, struct) to INI string.
func Encode(v any) ([]byte, error) { return EncodeWithDefName(v) }

// EncodeWithDefName golang data(map, struct) to INI, can set default section name
func EncodeWithDefName(v any, defSection ...string) (out []byte, err error) {
	switch vd := v.(type) {
	case map[string]any: // from full mode
		return EncodeFull(vd, defSection...)
	case map[string]map[string]string: // from lite mode
		return EncodeSimple(vd, defSection...)
	default:
		err = errors.New("ini: invalid data to encode as INI")
	}
	return
}

// EncodeFull full mode data to INI, can set default section name
func EncodeFull(data map[string]any, defSection ...string) (out []byte, err error) {
	ln := len(data)
	if ln == 0 {
		return
	}

	defSecName := ""
	if len(defSection) > 0 {
		defSecName = defSection[0]
	}

	sortedGroups := make([]string, 0, ln)
	for section := range data {
		sortedGroups = append(sortedGroups, section)
	}

	buf := &bytes.Buffer{}
	buf.Grow(ln * 4)
	buf.WriteString("; exported at " + timex.Now().Datetime() + "\n\n")
	secBuf := &bytes.Buffer{}

	sort.Strings(sortedGroups)
	max := len(sortedGroups) - 1
	for idx, section := range sortedGroups {
		item := data[section]
		switch tpData := item.(type) {
		case []int:
		case []string: // array of the default section
			for _, v := range tpData {
				buf.WriteString(fmt.Sprintf("%s[] = %v\n", section, v))
			}
		// case map[string]string: // is section
		case map[string]any: // is section
			if section != defSecName {
				secBuf.WriteString("[" + section + "]\n")
				writeAnyMap(secBuf, tpData)
			} else {
				writeAnyMap(buf, tpData)
			}

			if idx < max {
				secBuf.WriteByte('\n')
			}
		default: // k-v of the default section
			buf.WriteString(fmt.Sprintf("%s = %v\n", section, tpData))
		}
	}

	buf.WriteByte('\n')
	buf.Write(secBuf.Bytes())
	out = buf.Bytes()
	secBuf = nil
	return
}

func writeAnyMap(buf *bytes.Buffer, data map[string]any) {
	for key, item := range data {
		switch tpData := item.(type) {
		case []int:
		case []string: // array of the default section
			for _, v := range tpData {
				buf.WriteString(key + "[] = ")
				buf.WriteString(fmt.Sprint(v))
				buf.WriteByte('\n')
			}
		default: // k-v of the section
			buf.WriteString(key + " = ")
			buf.WriteString(fmt.Sprint(tpData))
			buf.WriteByte('\n')
		}
	}
}

// EncodeSimple data to INI
func EncodeSimple(data map[string]map[string]string, defSection ...string) ([]byte, error) {
	return EncodeLite(data, defSection...)
}

// EncodeLite data to INI
func EncodeLite(data map[string]map[string]string, defSection ...string) (out []byte, err error) {
	ln := len(data)
	if ln == 0 {
		return
	}

	defSecName := ""
	if len(defSection) > 0 {
		defSecName = defSection[0]
	}

	sortedGroups := make([]string, 0, ln)
	for section := range data {
		// don't add section title for default section
		if section != defSecName {
			sortedGroups = append(sortedGroups, section)
		}
	}

	buf := &bytes.Buffer{}
	buf.Grow(ln * 4)
	buf.WriteString("; exported at " + timex.Now().Datetime() + "\n\n")

	// first, write default section values
	if defSec, ok := data[defSecName]; ok {
		writeStringMap(buf, defSec)
		buf.WriteByte('\n')
	}

	sort.Strings(sortedGroups)
	max := len(sortedGroups) - 1
	for idx, section := range sortedGroups {
		buf.WriteString("[" + section + "]\n")
		writeStringMap(buf, data[section])

		if idx < max {
			buf.WriteByte('\n')
		}
	}

	out = buf.Bytes()
	return
}

func writeStringMap(buf *bytes.Buffer, strMap map[string]string) {
	sortedKeys := make([]string, 0, len(strMap))
	for key := range strMap {
		sortedKeys = append(sortedKeys, key)
	}

	sort.Strings(sortedKeys)
	for _, key := range sortedKeys {
		buf.WriteString(key + " = " + strMap[key] + "\n")
	}
}
