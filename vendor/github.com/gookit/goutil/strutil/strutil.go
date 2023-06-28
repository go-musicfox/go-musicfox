// Package strutil provide some string,char,byte util functions
package strutil

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"text/template"
)

// OrCond return s1 on cond is True, OR return s2.
// Like: cond ? s1 : s2
func OrCond(cond bool, s1, s2 string) string {
	if cond {
		return s1
	}
	return s2
}

// OrElse return s OR orVal(new-value) on s is empty
func OrElse(s, orVal string) string {
	if s != "" {
		return s
	}
	return orVal
}

// OrHandle return fn(s) on s is not empty.
func OrHandle(s string, fn func(s string) string) string {
	if s != "" {
		return fn(s)
	}
	return s
}

// Valid return first not empty element.
func Valid(ss ...string) string {
	for _, s := range ss {
		if s != "" {
			return s
		}
	}
	return ""
}

// Replaces replace multi strings
//
//	pairs: {old1: new1, old2: new2, ...}
//
// Can also use:
//
//	strings.NewReplacer("old1", "new1", "old2", "new2").Replace(str)
func Replaces(str string, pairs map[string]string) string {
	return NewReplacer(pairs).Replace(str)
}

// NewReplacer instance
func NewReplacer(pairs map[string]string) *strings.Replacer {
	ss := make([]string, len(pairs)*2)
	for old, newVal := range pairs {
		ss = append(ss, old, newVal)
	}
	return strings.NewReplacer(ss...)
}

// PrettyJSON get pretty Json string
// Deprecated: please use fmtutil.PrettyJSON() or jsonutil.Pretty() instead it
func PrettyJSON(v any) (string, error) {
	out, err := json.MarshalIndent(v, "", "    ")
	return string(out), err
}

// RenderTemplate render text template
func RenderTemplate(input string, data any, fns template.FuncMap, isFile ...bool) string {
	return RenderText(input, data, fns, isFile...)
}

// RenderText render text template
func RenderText(input string, data any, fns template.FuncMap, isFile ...bool) string {
	t := template.New("simple-text")
	t.Funcs(template.FuncMap{
		// don't escape content
		"raw": func(s string) string {
			return s
		},
		"trim": func(s string) string {
			return strings.TrimSpace(s)
		},
		// join strings
		"join": func(ss []string, sep string) string {
			return strings.Join(ss, sep)
		},
		// lower first char
		"lcFirst": func(s string) string {
			return LowerFirst(s)
		},
		// upper first char
		"upFirst": func(s string) string {
			return UpperFirst(s)
		},
	})

	// add custom template functions
	if len(fns) > 0 {
		t.Funcs(fns)
	}

	if len(isFile) > 0 && isFile[0] {
		template.Must(t.ParseFiles(input))
	} else {
		template.Must(t.Parse(input))
	}

	// use buffer receive rendered content
	buf := new(bytes.Buffer)
	if err := t.Execute(buf, data); err != nil {
		panic(err)
	}
	return buf.String()
}

// WrapTag for given string.
func WrapTag(s, tag string) string {
	if s == "" {
		return s
	}
	return fmt.Sprintf("<%s>%s</%s>", tag, s, tag)
}

// SubstrCount returns the number of times the substr substring occurs in the s string.
// Actually, it comes from strings.Count().
// s The string to search in
// substr The substring to search for
// params[0] The offset where to start counting.
// params[1] The maximum length after the specified offset to search for the substring.
func SubstrCount(s string, substr string, params ...uint64) (int, error) {
	larg := len(params)
	hasArgs := larg != 0
	if hasArgs && larg > 2 {
		return 0, errors.New("too many parameters")
	}
	if !hasArgs {
		return strings.Count(s, substr), nil
	}
	strlen := len(s)
	offset := 0
	end := strlen
	if hasArgs {
		offset = int(params[0])
		if larg == 2 {
			length := int(params[1])
			end = offset + length
		}
		if end > strlen {
			end = strlen
		}
	}
	s = string([]rune(s)[offset:end])
	return strings.Count(s, substr), nil
}
