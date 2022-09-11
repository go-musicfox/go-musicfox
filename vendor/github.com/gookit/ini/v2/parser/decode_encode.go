package parser

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"sort"
)

// TagName default tag-name of mapping data to struct
var TagName = "ini"

// Decode INI content to golang data
func Decode(blob []byte, ptr interface{}) error {
	rv := reflect.ValueOf(ptr)
	if rv.Kind() != reflect.Ptr {
		return fmt.Errorf("ini: Decode of non-pointer %s", reflect.TypeOf(ptr))
	}

	// if rv.IsNil() {
	// 	return fmt.Errorf("ini: Decode of nil %s", reflect.TypeOf(v))
	// }

	p, err := Parse(string(blob), ModeFull, NoDefSection)
	if err != nil {
		return err
	}

	return p.MapStruct(ptr)
}

// Encode golang data to INI
func Encode(v interface{}, defSection ...string) (out []byte, err error) {
	switch vd := v.(type) {
	case map[string]interface{}: // from full mode
		return EncodeFull(vd, defSection...)
	case map[string]map[string]string: // from simple mode
		return EncodeSimple(vd, defSection...)
	default:
		err = errors.New("ini: invalid data to encode as ini")
	}
	return
}

// EncodeFull full mode data to INI
func EncodeFull(data map[string]interface{}, defSection ...string) (out []byte, err error) {
	if len(data) == 0 {
		return
	}

	defSecName := ""
	if len(defSection) > 0 {
		defSecName = defSection[0]
	}

	// sort data
	counter := 0
	sections := make([]string, len(data))
	for section := range data {
		sections[counter] = section
		counter++
	}
	sort.Strings(sections)

	defBuf := &bytes.Buffer{}
	secBuf := &bytes.Buffer{}

	for _, key := range sections {
		item := data[key]
		switch tpData := item.(type) {
		case float32, float64, int, int32, int64, string, bool: // k-v of the default section
			_, err = defBuf.WriteString(fmt.Sprintf("%s = %v\n", key, tpData))
			if err != nil {
				return
			}
		case []int:
		case []string: // array of the default section
			for _, v := range tpData {
				_, err = defBuf.WriteString(fmt.Sprintf("%s[] = %v\n", key, v))
				if err != nil {
					return
				}
			}
		// case map[string]string: // is section
		case map[string]interface{}: // is section
			if key != defSecName {
				secBuf.WriteString("[" + key + "]\n")
				err = buildSectionBuffer(tpData, secBuf)
			} else {
				err = buildSectionBuffer(tpData, defBuf)
			}

			if err != nil {
				return
			}
		}
	}

	defBuf.WriteString(secBuf.String())
	out = defBuf.Bytes()
	secBuf = nil

	return
}

func buildSectionBuffer(data map[string]interface{}, buf *bytes.Buffer) (err error) {
	for key, item := range data {
		switch tpData := item.(type) {
		case float32, float64, int, int32, int64, string, bool: // k-v of the section
			_, err = buf.WriteString(fmt.Sprintf("%s = %v\n", key, tpData))
			if err != nil {
				return
			}
		case []int:
		case []string: // array of the default section
			for _, v := range tpData {
				_, err = buf.WriteString(fmt.Sprintf("%s[] = %v\n", key, v))
				if err != nil {
					return
				}
			}
		default: // skip invalid data
			continue
		}
	}
	return
}

// EncodeSimple data to INI
func EncodeSimple(data map[string]map[string]string, defSection ...string) (out []byte, err error) {
	if len(data) == 0 {
		return
	}

	var n int64
	buf := &bytes.Buffer{}
	counter := 0
	thisWrite := 0
	defSecName := ""
	orderedSections := make([]string, len(data))

	if len(defSection) > 0 {
		defSecName = defSection[0]
	}

	for section := range data {
		orderedSections[counter] = section
		counter++
	}

	sort.Strings(orderedSections)
	for _, section := range orderedSections {
		// don't add section title for DefSection
		if section != defSecName {
			thisWrite, err = fmt.Fprintln(buf, "["+section+"]")
			n += int64(thisWrite)
			if err != nil {
				return
			}
		}

		counter = 0
		items := data[section]
		orderedStringKeys := make([]string, len(items))

		for key := range items {
			orderedStringKeys[counter] = key
			counter++
		}

		sort.Strings(orderedStringKeys)
		for _, key := range orderedStringKeys {
			thisWrite, err = fmt.Fprintln(buf, key, "=", items[key])
			n += int64(thisWrite)
			if err != nil {
				return
			}
		}

		thisWrite, err = fmt.Fprintln(buf)
		n += int64(thisWrite)
		if err != nil {
			return
		}
	}

	out = buf.Bytes()
	return
}
