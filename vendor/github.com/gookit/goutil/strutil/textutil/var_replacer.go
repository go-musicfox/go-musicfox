package textutil

import (
	"reflect"
	"regexp"
	"strings"

	"github.com/gookit/goutil/arrutil"
	"github.com/gookit/goutil/internal/comfunc"
	"github.com/gookit/goutil/maputil"
	"github.com/gookit/goutil/strutil"
)

const defaultVarFormat = "{{,}}"

// FallbackFn type
type FallbackFn = func(name string) (val string, ok bool)

// VarReplacer struct
type VarReplacer struct {
	init bool

	Left, Right string
	lLen, rLen  int

	varReg *regexp.Regexp
	// flatten sub map in vars
	flatSubs bool
	parseEnv bool
	// support parse default value. eg: {{ name | inhere }}
	parseDef bool
	// keepMissVars list. default False: will clear on each replace
	keepMissVars bool
	// missing vars list
	missVars []string
	// NotFound handler
	NotFound FallbackFn
}

// NewVarReplacer instance
func NewVarReplacer(format string, opFns ...func(vp *VarReplacer)) *VarReplacer {
	vp := &VarReplacer{flatSubs: true}
	for _, fn := range opFns {
		fn(vp)
	}
	return vp.WithFormat(format)
}

// NewFullReplacer instance
func NewFullReplacer(format string) *VarReplacer {
	return NewVarReplacer(format, func(vp *VarReplacer) {
		vp.WithParseEnv().WithParseDefault().KeepMissingVars()
	})
}

// DisableFlatten on the input vars map
func (r *VarReplacer) DisableFlatten() *VarReplacer {
	r.flatSubs = false
	return r
}

// KeepMissingVars on the replacement handle
func (r *VarReplacer) KeepMissingVars() *VarReplacer {
	r.keepMissVars = true
	return r
}

// WithParseDefault value on the input template contents
func (r *VarReplacer) WithParseDefault() *VarReplacer {
	r.parseDef = true
	return r
}

// WithParseEnv on the input vars value
func (r *VarReplacer) WithParseEnv() *VarReplacer {
	r.parseEnv = true
	return r
}

// OnNotFound var handle
func (r *VarReplacer) OnNotFound(fn FallbackFn) *VarReplacer {
	r.NotFound = fn
	return r
}

// WithFormat custom var template
func (r *VarReplacer) WithFormat(format string) *VarReplacer {
	r.Left, r.Right = strutil.QuietCut(strutil.OrElse(format, defaultVarFormat), ",")
	r.Init()
	return r
}

// Init var matcher
func (r *VarReplacer) Init() *VarReplacer {
	if !r.init {
		r.lLen, r.rLen = len(r.Left), len(r.Right)
		if r.Right != "" {
			r.varReg = regexp.MustCompile(regexp.QuoteMeta(r.Left) + `([\w\s\|.-]+)` + regexp.QuoteMeta(r.Right))
		} else {
			// no right tag. eg: $name, $user.age
			r.varReg = regexp.MustCompile(regexp.QuoteMeta(r.Left) + `(\w[\w-]*(?:\.[\w-]+)*)`)
		}
	}

	return r
}

// ParseVars the text contents and collect vars
func (r *VarReplacer) ParseVars(s string) []string {
	ss := arrutil.StringsMap(r.varReg.FindAllString(s, -1), func(val string) string {
		return strings.TrimSpace(val[r.lLen : len(val)-r.rLen])
	})

	return arrutil.Unique(ss)
}

// Render any-map vars in the text contents
func (r *VarReplacer) Render(s string, tplVars map[string]any) string {
	return r.Replace(s, tplVars)
}

// Replace any-map vars in the text contents
func (r *VarReplacer) Replace(s string, tplVars map[string]any) string {
	if !strings.Contains(s, r.Left) {
		return s
	}
	if !r.parseDef && len(tplVars) == 0 {
		return s
	}

	var varMap map[string]string

	if r.flatSubs {
		varMap = make(map[string]string, len(tplVars)*2)
		maputil.FlatWithFunc(tplVars, func(path string, val reflect.Value) {
			if val.Kind() == reflect.String {
				if r.parseEnv {
					varMap[path] = comfunc.ParseEnvVar(val.String(), nil)
				} else {
					varMap[path] = val.String()
				}
			} else {
				varMap[path] = strutil.QuietString(val.Interface())
			}
		})
	} else {
		varMap = maputil.ToStringMap(tplVars)
	}

	return r.Init().doReplace(s, varMap)
}

// ReplaceSMap string-map vars in the text contents
func (r *VarReplacer) ReplaceSMap(s string, varMap map[string]string) string {
	return r.RenderSimple(s, varMap)
}

// RenderSimple string-map vars in the text contents. alias of ReplaceSMap()
func (r *VarReplacer) RenderSimple(s string, varMap map[string]string) string {
	if len(varMap) == 0 || !strings.Contains(s, r.Left) {
		return s
	}

	if r.parseEnv {
		for name, val := range varMap {
			varMap[name] = comfunc.ParseEnvVar(val, nil)
		}
	}

	return r.Init().doReplace(s, varMap)
}

// MissVars list
func (r *VarReplacer) MissVars() []string {
	return r.missVars
}

// ResetMissVars list
func (r *VarReplacer) ResetMissVars() {
	r.missVars = make([]string, 0)
}

// Replace string-map vars in the text contents
func (r *VarReplacer) doReplace(s string, varMap map[string]string) string {
	if !r.keepMissVars {
		r.missVars = make([]string, 0) // clear on each replace
	}

	return r.varReg.ReplaceAllStringFunc(s, func(sub string) string {
		name := strings.TrimSpace(sub[r.lLen : len(sub)-r.rLen])

		var defVal string
		if r.parseDef && strings.ContainsRune(name, '|') {
			name, defVal = strutil.TrimCut(name, "|")
		}

		if val, ok := varMap[name]; ok {
			return val
		}

		if r.NotFound != nil {
			if val, ok := r.NotFound(name); ok {
				return val
			}
		}

		if len(defVal) > 0 {
			return defVal
		}
		r.missVars = append(r.missVars, name)
		return sub
	})
}
