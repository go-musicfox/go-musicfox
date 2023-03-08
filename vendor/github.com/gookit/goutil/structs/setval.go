package structs

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/gookit/goutil/internal/comfunc"
	"github.com/gookit/goutil/reflects"
	"github.com/gookit/goutil/strutil"
)

const defaultInitTag = "default"

// InitOptFunc define
type InitOptFunc func(opt *InitOptions)

// InitOptions struct
type InitOptions struct {
	// TagName default value tag name. tag: default
	TagName string
	// ParseEnv var name on default value. eg: `default:"${APP_ENV}"`
	//
	// default: false
	ParseEnv bool
	// ValueHook before set value hook TODO
	ValueHook func(val string) any
}

// InitDefaults init struct default value by field "default" tag.
//
// TIPS:
//
//	Support init field types: string, bool, intX, uintX, floatX, array, slice
//
// Example:
//
//	type User1 struct {
//		Name string `default:"inhere"`
//		Age  int32  `default:"30"`
//	}
//
//	u1 := &User1{}
//	err = structs.InitDefaults(u1)
//	fmt.Printf("%+v\n", u1) // Output: {Name:inhere Age:30}
func InitDefaults(ptr any, optFns ...InitOptFunc) error {
	rv := reflect.ValueOf(ptr)
	if rv.Kind() != reflect.Ptr {
		return errors.New("must be provider an pointer value")
	}

	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		return errors.New("must be provider an struct value")
	}

	opt := &InitOptions{TagName: defaultInitTag}
	for _, fn := range optFns {
		fn(opt)
	}

	return initDefaults(rv, opt)
}

func initDefaults(rv reflect.Value, opt *InitOptions) error {
	rt := rv.Type()

	for i := 0; i < rt.NumField(); i++ {
		ft := rt.Field(i)
		// skip don't exported field
		if ft.Name[0] >= 'a' && ft.Name[0] <= 'z' {
			continue
		}

		fv := rv.Field(i)
		if fv.Kind() == reflect.Struct {
			if err := initDefaults(fv, opt); err != nil {
				return err
			}
			continue
		}

		// skip on field has value
		if !fv.IsZero() {
			continue
		}

		val := ft.Tag.Get(opt.TagName)
		if err := initDefaultValue(fv, val, opt.ParseEnv); err != nil {
			return err
		}
	}

	return nil
}

func initDefaultValue(fv reflect.Value, val string, parseEnv bool) error {
	if val == "" || !fv.CanSet() {
		return nil
	}

	// parse env var
	if parseEnv {
		val = comfunc.ParseEnvVar(val, nil)
	}

	var anyVal any = val

	// convert string to slice
	if reflects.IsArrayOrSlice(fv.Kind()) {
		ss := strutil.SplitTrimmed(val, ",")
		valRv, err := reflects.ConvSlice(reflect.ValueOf(ss), fv.Type().Elem())
		if err != nil {
			return err
		}
		anyVal = valRv.Interface()
	}

	// set value
	return reflects.SetValue(fv, anyVal)
}

/*************************************************************
 * load values to a struct
 *************************************************************/

// SetOptFunc define
type SetOptFunc func(opt *SetOptions)

// SetOptions for set values to struct
type SetOptions struct {
	// FieldTagName get field name for read value. default tag: json
	FieldTagName string
	// ValueHook before set value hook TODO
	ValueHook func(val any) any

	// ParseDefault init default value by DefaultValTag tag value.
	// default: false
	//
	// see InitDefaults()
	ParseDefault bool
	// DefaultValTag name. tag: default
	DefaultValTag string
	// ParseDefaultEnv parse env var on default tag. eg: `default:"${APP_ENV}"`
	//
	// default: false
	ParseDefaultEnv bool
}

// SetValues set data values to struct ptr
//
// TIPS:
//
//	Only support set: string, bool, intX, uintX, floatX
func SetValues(ptr any, data map[string]any, optFns ...SetOptFunc) error {
	rv := reflect.ValueOf(ptr)
	if rv.Kind() != reflect.Ptr {
		return errors.New("must be provider an pointer value")
	}

	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		return errors.New("must be provider an struct value")
	}

	opt := &SetOptions{
		FieldTagName:  "json",
		DefaultValTag: defaultInitTag,
	}

	for _, fn := range optFns {
		fn(opt)
	}
	return setValues(rv, data, opt)
}

func setValues(rv reflect.Value, data map[string]any, opt *SetOptions) error {
	if len(data) == 0 {
		return nil
	}

	rt := rv.Type()

	for i := 0; i < rt.NumField(); i++ {
		ft := rt.Field(i)
		name := ft.Name
		// skip don't exported field
		if name[0] >= 'a' && name[0] <= 'z' {
			continue
		}

		// get field name
		tagVal, ok := ft.Tag.Lookup(opt.FieldTagName)
		if ok {
			info, err := ParseTagValueDefault(name, tagVal)
			if err != nil {
				return err
			}

			name = info.Get("name")
		}

		fv := rv.Field(i)
		val, ok := data[name]

		// set field value by default tag.
		if !ok && fv.IsZero() {
			defVal := ft.Tag.Get(opt.DefaultValTag)

			if err := initDefaultValue(fv, defVal, opt.ParseDefaultEnv); err != nil {
				return err
			}
			continue
		}

		// field is struct
		if fv.Kind() == reflect.Struct {
			asMp, ok := val.(map[string]any)
			if !ok {
				return fmt.Errorf("field is struct, must provide map data value")
			}

			if err := setValues(fv, asMp, opt); err != nil {
				return err
			}
			continue
		}

		// set field value
		if err := reflects.SetValue(fv, val); err != nil {
			return err
		}
	}

	return nil
}
