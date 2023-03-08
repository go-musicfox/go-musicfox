package structs

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/gookit/goutil/arrutil"
	"github.com/gookit/goutil/maputil"
	"github.com/gookit/goutil/strutil"
)

// ErrNotAnStruct error
// var emptyStringMap = make(maputil.SMap)
var ErrNotAnStruct = errors.New("must input an struct value")

// ParseTags for parse struct tags.
func ParseTags(st any, tagNames []string) (map[string]maputil.SMap, error) {
	p := NewTagParser(tagNames...)

	if err := p.Parse(st); err != nil {
		return nil, err
	}
	return p.Tags(), nil
}

// ParseReflectTags parse struct tags info.
func ParseReflectTags(rt reflect.Type, tagNames []string) (map[string]maputil.SMap, error) {
	p := NewTagParser(tagNames...)

	if err := p.ParseType(rt); err != nil {
		return nil, err
	}
	return p.Tags(), nil
}

// TagValFunc handle func
type TagValFunc func(field, tagVal string) (maputil.SMap, error)

// TagParser struct
type TagParser struct {
	// TagNames want parsed tag names.
	TagNames []string
	// ValueFunc tag value parse func.
	ValueFunc TagValFunc

	// key: field name
	// value: tag map {tag-name: value string.}
	tags map[string]maputil.SMap
}

// Tags map data for struct fields
func (p *TagParser) Tags() map[string]maputil.SMap {
	return p.tags
}

// NewTagParser instance
func NewTagParser(tagNames ...string) *TagParser {
	return &TagParser{
		TagNames:  tagNames,
		ValueFunc: ParseTagValueDefault,
	}
}

// Parse an struct value
func (p *TagParser) Parse(st any) error {
	rv := reflect.ValueOf(st)
	if rv.Kind() == reflect.Ptr && !rv.IsNil() {
		rv = rv.Elem()
	}

	return p.ParseType(rv.Type())
}

// ParseType parse a struct type value
func (p *TagParser) ParseType(rt reflect.Type) error {
	if rt.Kind() != reflect.Struct {
		return ErrNotAnStruct
	}

	// key is field name.
	p.tags = make(map[string]maputil.SMap)
	return p.parseType(rt, "")
}

func (p *TagParser) parseType(rt reflect.Type, parent string) error {
	for i := 0; i < rt.NumField(); i++ {
		sf := rt.Field(i)

		// skip don't exported field
		name := sf.Name
		if name[0] >= 'a' && name[0] <= 'z' {
			continue
		}

		smp := make(maputil.SMap)
		for _, tagName := range p.TagNames {
			// eg: `json:"age"`
			// eg: "name=int0;shorts=i;required=true;desc=int option message"
			tagVal := sf.Tag.Get(tagName)
			if tagVal == "" {
				continue
			}

			smp[tagName] = tagVal
		}

		pathKey := name
		if parent != "" {
			pathKey = parent + "." + name
		}

		p.tags[pathKey] = smp

		ft := sf.Type
		if ft.Kind() == reflect.Ptr {
			ft = ft.Elem()
		}

		// field is struct.
		if ft.Kind() == reflect.Struct {
			err := p.parseType(ft, pathKey)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Info parse the give field, returns tag value info.
//
//	info, err := p.Info("Name", "json")
//	exportField := info.Get("name")
func (p *TagParser) Info(field, tag string) (maputil.SMap, error) {
	field = strutil.UpperFirst(field)
	fTags, ok := p.tags[field]
	if !ok {
		return nil, fmt.Errorf("field %q not found", field)
	}

	val, ok := fTags.Value(tag)
	if !ok {
		return make(maputil.SMap), nil
	}

	// parse tag value
	return p.ValueFunc(field, val)
}

/*************************************************************
 * some built in tag value parse func
 *************************************************************/

// ParseTagValueDefault parse like json tag value.
//
// see json.Marshal():
//
//	// JSON as key "myName", skipped if empty.
//	Field int `json:"myName,omitempty"`
//
//	// Field appears in JSON as key "Field" (the default), but skipped if empty.
//	Field int `json:",omitempty"`
//
//	// Field is ignored by this package.
//	Field int `json:"-"`
//
//	// Field appears in JSON as key "-".
//	Field int `json:"-,"`
//
//	Int64String int64 `json:",string"`
//
// Returns:
//
//	{
//		"name": "myName", // maybe is empty, on tag value is "-"
//		"omitempty": "true",
//		"string": "true",
//		// ... more custom bool settings.
//	}
func ParseTagValueDefault(field, tagVal string) (mp maputil.SMap, err error) {
	ss := strutil.SplitTrimmed(tagVal, ",")
	ln := len(ss)
	if ln == 0 || tagVal == "," {
		return maputil.SMap{"name": field}, nil
	}

	mp = make(maputil.SMap, ln)
	if ln == 1 {
		// valid field name
		if ss[0] != "-" {
			mp["name"] = ss[0]
		}
		return
	}

	// ln > 1
	mp["name"] = ss[0]
	// other settings: omitempty, string
	for _, key := range ss[1:] {
		mp[key] = "true"
	}
	return
}

// ParseTagValueQuick quick parse tag value string by sep(;)
func ParseTagValueQuick(tagVal string, defines []string) maputil.SMap {
	parseFn := ParseTagValueDefine(";", defines)

	mp, _ := parseFn("", tagVal)
	return mp
}

// ParseTagValueDefine parse tag value string by given defines.
//
// Examples:
//
//	eg: "desc;required;default;shorts"
//	type MyStruct {
//		Age int `flag:"int option message;;a,b"`
//	}
//	sepStr := ";"
//	defines := []string{"desc", "required", "default", "shorts"}
func ParseTagValueDefine(sep string, defines []string) TagValFunc {
	defNum := len(defines)

	return func(field, tagVal string) (maputil.SMap, error) {
		ss := strutil.SplitNTrimmed(tagVal, sep, defNum)
		ln := len(ss)
		mp := make(maputil.SMap, ln)
		if ln == 0 {
			return mp, nil
		}

		for i, val := range ss {
			key := defines[i]
			mp[key] = val
		}
		return mp, nil
	}
}

// ParseTagValueNamed parse k-v tag value string. it's like INI format contents.
//
// Examples:
//
//	eg: "name=val0;shorts=i;required=true;desc=a message"
//	=>
//	{name: val0, shorts: i, required: true, desc: a message}
func ParseTagValueNamed(field, tagVal string, keys ...string) (mp maputil.SMap, err error) {
	ss := strutil.Split(tagVal, ";")
	ln := len(ss)
	if ln == 0 {
		return
	}

	mp = make(maputil.SMap, ln)
	for _, s := range ss {
		if !strings.ContainsRune(s, '=') {
			err = fmt.Errorf("parse tag error on field '%s': must match `KEY=VAL`", field)
			return
		}

		key, val := strutil.TrimCut(s, "=")
		if len(keys) > 0 && !arrutil.StringsHas(keys, key) {
			err = fmt.Errorf("parse tag error on field '%s': invalid key name '%s'", field, key)
			return
		}

		mp[key] = val
	}
	return
}
