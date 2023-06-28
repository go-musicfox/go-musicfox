package structs

import (
	"errors"
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

// Init struct default value by field "default" tag.
func Init(ptr any, optFns ...InitOptFunc) error {
	return InitDefaults(ptr, optFns...)
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
		sf := rt.Field(i)
		// skip don't exported field
		if IsUnexported(sf.Name) {
			continue
		}

		val, hasTag := sf.Tag.Lookup(opt.TagName)
		if !hasTag || val == "-" {
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
			// special: handle for pointer struct field
			if fv.Kind() == reflect.Pointer {
				fv = fv.Elem()
				if fv.Kind() == reflect.Struct {
					if err := initDefaults(fv, opt); err != nil {
						return err
					}
				}
			} else if fv.Kind() == reflect.Slice {
				el := sf.Type.Elem()
				if el.Kind() == reflect.Pointer {
					el = el.Elem()
				}

				// init sub struct in slice. like `[]SubStruct` or `[]*SubStruct`
				if el.Kind() == reflect.Struct && fv.Len() > 0 {
					for i := 0; i < fv.Len(); i++ {
						subFv := reflect.Indirect(fv.Index(i))
						if err := initDefaults(subFv, opt); err != nil {
							return err
						}
					}
				}
			}
			continue
		}

		// handle for pointer field
		if fv.Kind() == reflect.Pointer {
			if fv.IsNil() {
				fv.Set(reflect.New(fv.Type().Elem()))
			}

			fv = fv.Elem()
			if fv.Kind() == reflect.Struct {
				if err := initDefaults(fv, opt); err != nil {
					return err
				}
				continue
			}
		} else if fv.Kind() == reflect.Slice {
			el := sf.Type.Elem()
			isPtr := el.Kind() == reflect.Pointer
			if isPtr {
				el = el.Elem()
			}

			// init sub struct in slice. like `[]SubStruct` or `[]*SubStruct`
			if el.Kind() == reflect.Struct {
				// make sub-struct and init. like: `SubStruct`
				subFv := reflect.New(el)
				subFvE := subFv.Elem()
				if err := initDefaults(subFvE, opt); err != nil {
					return err
				}

				// make new slice and set value.
				newFv := reflect.MakeSlice(reflect.SliceOf(sf.Type.Elem()), 0, 1)
				if isPtr {
					newFv = reflect.Append(newFv, subFv)
				} else {
					newFv = reflect.Append(newFv, subFvE)
				}
				fv.Set(newFv)
				continue
			}
		}

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

	// simple slice: convert simple kind(string,intX,uintX,...) to slice. eg: "1,2,3" => []int{1,2,3}
	if reflects.IsArrayOrSlice(fv.Kind()) && reflects.IsSimpleKind(reflects.SliceElemKind(fv.Type())) {
		ss := strutil.SplitTrimmed(val, ",")
		valRv, err := reflects.ConvSlice(reflect.ValueOf(ss), fv.Type().Elem())
		if err == nil {
			reflects.SetRValue(fv, valRv)
		}
		return err
	}

	// set value
	return reflects.SetValue(fv, anyVal)
}
