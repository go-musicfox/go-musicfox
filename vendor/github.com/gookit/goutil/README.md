# Go Util

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/gookit/goutil?style=flat-square)
[![GitHub tag (latest SemVer)](https://img.shields.io/github/tag/gookit/goutil)](https://github.com/gookit/goutil)
[![GoDoc](https://godoc.org/github.com/gookit/goutil?status.svg)](https://pkg.go.dev/github.com/gookit/goutil)
[![Go Report Card](https://goreportcard.com/badge/github.com/gookit/goutil)](https://goreportcard.com/report/github.com/gookit/goutil)
[![Unit-Tests](https://github.com/gookit/goutil/workflows/Unit-Tests/badge.svg)](https://github.com/gookit/goutil/actions)
[![Coverage Status](https://coveralls.io/repos/github/gookit/goutil/badge.svg?branch=master)](https://coveralls.io/github/gookit/goutil?branch=master)

💪 Useful utils for the Go: int, string, array/slice, map, error, time, format, CLI, ENV, filesystem, system, testing and more.

- [`arrutil`](./arrutil): Array/Slice util functions. eg: check, convert
- [`dump`](./dump):  Simple variable printing tool, printing slice, map will automatically wrap each element and display the call location
- `cliutil` Command-line util functions. eg: read input, exec command, cmdline parse/build
- [`errorx`](./errorx) Provide an enhanced error implements for go, allow with stacktraces and wrap another error.
- `envutil` ENV util for current runtime env information. eg: get one, get info, parse var
- `fmtutil` Format data util functions
- [`fsutil`](./fsutil) Filesystem util functions, quick create, read and write file. eg: file and dir check, operate
- `jsonutil` some util functions for quick read, write, encode, decode JSON data.
- [`maputil`](./maputil) Map data util functions. eg: convert, sub-value get, simple merge
- `mathutil`, `numutil` Math(int, number) util functions. eg: convert, math calc, random
- `netutil` Network util functions
  - `netutil/httpreq` An easier-to-use HTTP client that wraps http.Client
- `strutil` String util functions. eg: bytes, check, convert, encode, format and more
- `sysutil` System util functions. eg: sysenv, exec, user, process
- `testutil` Test help util functions. eg: http test, mock ENV value
- [`timex`](./timex) Provides an enhanced time.Time implementation. Add more commonly used functional methods
  - such as: DayStart(), DayAfter(), DayAgo(), DateFormat() and more.

> **[中文说明](README.zh-CN.md)**

## GoDoc

- [Godoc for github](https://pkg.go.dev/github.com/gookit/goutil)

## Packages

### Array/Slice

> Package `github.com/gookit/goutil/arrutil`

```go
// source at arrutil/arrutil.go
func Reverse(ss []string)
func StringsRemove(ss []string, s string) []string
func TrimStrings(ss []string, cutSet ...string) (ns []string)
func GetRandomOne(arr interface{}) interface{}
// source at arrutil/check.go
func IntsHas(ints []int, val int) bool
func Int64sHas(ints []int64, val int64) bool
func InStrings(elem string, ss []string) bool { return StringsHas(ss, elem) }
func StringsHas(ss []string, val string) bool
func HasValue(arr, val interface{}) bool
func Contains(arr, val interface{}) bool
func NotContains(arr, val interface{}) bool
// source at arrutil/convert.go
func JoinStrings(sep string, ss ...string) string
func StringsJoin(sep string, ss ...string) string
func StringsToInts(ss []string) (ints []int, err error)
func MustToStrings(arr interface{}) []string
func StringsToSlice(ss []string) []interface{}
func ToInt64s(arr interface{}) (ret []int64, err error)
func MustToInt64s(arr interface{}) []int64
func SliceToInt64s(arr []interface{}) []int64
func ToStrings(arr interface{}) (ret []string, err error)
func SliceToStrings(arr []interface{}) []string
func AnyToString(arr interface{}) string
func SliceToString(arr ...interface{}) string { return ToString(arr) }
func ToString(arr []interface{}) string
func JoinSlice(sep string, arr ...interface{}) string
```
#### ArrUtil Usage

**check value**:

```go
arrutil.IntsHas([]int{2, 4, 5}, 2) // True
arrutil.Int64sHas([]int64{2, 4, 5}, 2) // True
arrutil.StringsHas([]string{"a", "b"}, "a") // True

// list and val interface{}
arrutil.Contains(list, val)
arrutil.Contains([]uint32{9, 2, 3}, 9) // True
```

**convert**:

```go
ints, err := arrutil.ToInt64s([]string{"1", "2"}) // ints: []int64{1, 2} 
ss, err := arrutil.ToStrings([]int{1, 2}) // ss: []string{"1", "2"}
```

### CLI/Console

> Package `github.com/gookit/goutil/cliutil`

```go
// source at cliutil/cliutil.go
func LineBuild(binFile string, args []string) string
func BuildLine(binFile string, args []string) string
func String2OSArgs(line string) []string
func StringToOSArgs(line string) []string
func ParseLine(line string) []string
func QuickExec(cmdLine string, workDir ...string) (string, error)
func ExecLine(cmdLine string, workDir ...string) (string, error)
func ExecCmd(binName string, args []string, workDir ...string) (string, error)
func ExecCommand(binName string, args []string, workDir ...string) (string, error)
func ShellExec(cmdLine string, shells ...string) (string, error)
func CurrentShell(onlyName bool) (path string)
func HasShellEnv(shell string) bool
func Workdir() string
func BinDir() string
func BinFile() string
// source at cliutil/read.go
func ReadInput(question string) (string, error)
func ReadLine(question string) (string, error)
func ReadFirst(question string) (string, error)
func ReadFirstByte(question string) (byte, error)
func ReadFirstRune(question string) (rune, error)
// source at cliutil/read_nonwin.go
func ReadPassword(question ...string) string
```

#### CLI Util Usage

**helper functions:**

```go
cliutil.Workdir() // current workdir
cliutil.BinDir() // the program exe file dir

cliutil.ReadInput("Your name?")
cliutil.ReadPassword("Input password:")
ans, _ := cliutil.ReadFirstByte("continue?[y/n] ")
```

**cmdline parse:**

```go
package main

import (
	"fmt"

	"github.com/gookit/goutil/cliutil"
	"github.com/gookit/goutil/dump"
)

func main() {
	args := cliutil.ParseLine(`./app top sub --msg "has multi words"`)
	dump.P(args)

	s := cliutil.BuildLine("./myapp", []string{
		"-a", "val0",
		"-m", "this is message",
		"arg0",
	})
	fmt.Println("Build line:", s)
}
```

**output**:

```text
PRINT AT github.com/gookit/goutil/cliutil_test.TestParseLine(line_parser_test.go:30)
[]string [ #len=5
  string("./app"), #len=5
  string("top"), #len=3
  string("sub"), #len=3
  string("--msg"), #len=5
  string("has multi words"), #len=15
]

Build line: ./myapp -a val0 -m "this is message" arg0
```

### Dump

> Package `github.com/gookit/goutil/dump`

```go
// source at dump/dump.go
func Std() *Dumper
func Reset()
func Config(fn func(opts *Options))
func Print(vs ...interface{})
func Println(vs ...interface{})
func Fprint(w io.Writer, vs ...interface{})
func Format(vs ...interface{}) string
func NoLoc(vs ...interface{})
func Clear(vs ...interface{})
// source at dump/dumper.go
func NewDumper(out io.Writer, skip int) *Dumper
func NewWithOptions(fn func(opts *Options)) *Dumper
func NewDefaultOptions(out io.Writer, skip int) *Options
```
#### Examples

example code:

```go
package main

import "github.com/gookit/goutil/dump"

// rum demo:
// 	go run ./dump/_examples/demo1.go
func main() {
	otherFunc1()
}

func otherFunc1() {
	dump.P(
		23,
		[]string{"ab", "cd"},
		[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}, // len > 10
		map[string]interface{}{
			"key": "val", "sub": map[string]string{"k": "v"},
		},
		struct {
			ab string
			Cd int
		}{
			"ab", 23,
		},
	)
}
```

Preview:

![](dump/_examples/preview-demo1.png)

**nested struct**

> source code at `dump/dumper_test.TestStruct_WithNested`

![](dump/_examples/preview-nested-struct.png)

### ENV/Environment

> Package `github.com/gookit/goutil/envutil`

```go
// source at envutil/envutil.go
func VarReplace(s string) string
func VarParse(str string) string
func ParseEnvValue(str string) string
func ParseValue(val string) (newVal string)
// source at envutil/get.go
func Getenv(name string, def ...string) string
func GetInt(name string, def ...int) int
func GetBool(name string, def ...bool) bool
func Environ() map[string]string
// source at envutil/info.go
func IsWin() bool
func IsWindows() bool
func IsMac() bool
func IsLinux() bool
func IsMSys() bool
func IsWSL() bool
func IsTerminal(fd uintptr) bool
func StdIsTerminal() bool
func IsConsole(out io.Writer) bool
func HasShellEnv(shell string) bool
func IsSupportColor() bool
func IsSupport256Color() bool
func IsSupportTrueColor() bool
```
#### ENV Util Usage

**helper functions:**

```go
envutil.IsWin()
envutil.IsMac()
envutil.IsLinux()

// get ENV value by key, can with default value
envutil.Getenv("APP_ENV", "dev")
envutil.GetInt("LOG_LEVEL", 1)
envutil.GetBool("APP_DEBUG", true)

// parse ENV var value from input string, support default value.
envutil.ParseValue("${ENV_NAME | defValue}")
```

### Errorx

> Package `github.com/gookit/goutil/errorx`

Package errorx provide a enhanced error implements, allow with call stack and wrap another error.

```go
// source at errorx/errorx.go
func New(msg string) error
func Newf(tpl string, vars ...interface{}) error
func Errorf(tpl string, vars ...interface{}) error
func With(err error, msg string) error
func Withf(err error, tpl string, vars ...interface{}) error
func WithPrev(err error, msg string) error
func WithPrevf(err error, tpl string, vars ...interface{}) error
func WithStack(err error) error
func Traced(err error) error
func Stacked(err error) error
func WithOptions(msg string, fns ...func(opt *ErrStackOpt)) error
func Wrap(err error, msg string) error
func Wrapf(err error, tpl string, vars ...interface{}) error
// source at errorx/reply.go
func NewR(code int, msg string) ErrorR
func Fail(code int, msg string) ErrorR
func Suc(msg string) ErrorR
// source at errorx/stack.go
func FuncForPC(pc uintptr) *Func
func Config(fns ...func(opt *ErrStackOpt))
func SkipDepth(skipDepth int) func(opt *ErrStackOpt)
func TraceDepth(traceDepth int) func(opt *ErrStackOpt)
// source at errorx/util.go
func Raw(msg string) error
func Rawf(tpl string, vars ...interface{}) error
func Cause(err error) error
func Unwrap(err error) error
func Previous(err error) error { return Unwrap(err) }
func Has(err, target error) bool
func Is(err, target error) bool
func To(err error, target interface{}) bool
func As(err error, target interface{}) bool
```

#### Errorx Usage

**Create error with call stack info**

- use the `errorx.New` instead `errors.New`

```go
func doSomething() error {
    if false {
	    // return errors.New("a error happen")
	    return errorx.New("a error happen")
	}
}
```

- use the `errorx.Newf` or `errorx.Errorf` instead `fmt.Errorf`

```go
func doSomething() error {
    if false {
	    // return fmt.Errorf("a error %s", "happen")
	    return errorx.Newf("a error %s", "happen")
	}
}
```

**Wrap the previous error**

used like this before:

```go
    if err := SomeFunc(); err != nil {
	    return err
	}
```

can be replaced with:

```go
    if err := SomeFunc(); err != nil {
	    return errors.Stacked(err)
	}
```

**Print the errorx.New() error**

Examples for use `errorx` package, more please see [./errorx/README](errorx/README.md)

```go
    err := errorx.New("the error message")

    fmt.Println(err)
    // fmt.Printf("%v\n", err)
    // fmt.Printf("%#v\n", err)
```

> from the test: `errorx/errorx_test.TestNew()`

**Output**:

```text
the error message
STACK:
github.com/gookit/goutil/errorx_test.returnXErr()
  /Users/inhere/Workspace/godev/gookit/goutil/errorx/errorx_test.go:21
github.com/gookit/goutil/errorx_test.returnXErrL2()
  /Users/inhere/Workspace/godev/gookit/goutil/errorx/errorx_test.go:25
github.com/gookit/goutil/errorx_test.TestNew()
  /Users/inhere/Workspace/godev/gookit/goutil/errorx/errorx_test.go:29
testing.tRunner()
  /usr/local/Cellar/go/1.18/libexec/src/testing/testing.go:1439
runtime.goexit()
  /usr/local/Cellar/go/1.18/libexec/src/runtime/asm_amd64.s:1571
```

### Formatting

> Package `github.com/gookit/goutil/fmtutil`

```go
// source at fmtutil/format.go
func DataSize(bytes uint64) string
func PrettyJSON(v interface{}) (string, error)
func StringsToInts(ss []string) (ints []int, err error)
func ArgsWithSpaces(args []interface{}) (message string)
// source at fmtutil/time.go
func HowLongAgo(sec int64) string
```

### FileSystem

> Package `github.com/gookit/goutil/fsutil`

```go
// source at fsutil/check.go
func PathExists(path string) bool
func IsDir(path string) bool
func FileExists(path string) bool
func IsFile(path string) bool
func IsAbsPath(aPath string) bool
func IsImageFile(path string) bool
func IsZipFile(filepath string) bool
// source at fsutil/finder.go
func EmptyFinder() *FileFinder
func NewFinder(dirPaths []string, filePaths ...string) *FileFinder
func ExtFilterFunc(exts []string, include bool) FileFilterFunc
func SuffixFilterFunc(suffixes []string, include bool) FileFilterFunc
func PathNameFilterFunc(names []string, include bool) FileFilterFunc
func DotFileFilterFunc(include bool) FileFilterFunc
func ModTimeFilterFunc(limitSec int, op rune, include bool) FileFilterFunc
func GlobFilterFunc(patterns []string, include bool) FileFilterFunc
func RegexFilterFunc(pattern string, include bool) FileFilterFunc
func DotDirFilterFunc(include bool) DirFilterFunc
func DirNameFilterFunc(names []string, include bool) DirFilterFunc
// source at fsutil/fsutil.go
func DiscardReader(src io.Reader)
func OSTempFile(pattern string) (*os.File, error)
func TempFile(dir, pattern string) (*os.File, error)
func OSTempDir(pattern string) (string, error)
func TempDir(dir, pattern string) (string, error)
func MimeType(path string) (mime string)
func ReaderMimeType(r io.Reader) (mime string)
func GlobWithFunc(pattern string, fn func(filePath string) error) (err error)
// source at fsutil/info.go
func Dir(fpath string) string
func PathName(fpath string) string
func Name(fpath string) string
func FileExt(fpath string) string
func Suffix(fpath string) string
func ExpandPath(path string) string
func Realpath(pathStr string) string
// source at fsutil/operate.go
func Mkdir(dirPath string, perm os.FileMode) error
func MkParentDir(fpath string) error
func MustReadFile(filePath string) []byte
func MustReadReader(r io.Reader) []byte
func ReadExistFile(filePath string) []byte
func OpenFile(filepath string, flag int, perm os.FileMode) (*os.File, error)
func QuickOpenFile(filepath string) (*os.File, error)
func CreateFile(fpath string, filePerm, dirPerm os.FileMode) (*os.File, error)
func MustCreateFile(filePath string, filePerm, dirPerm os.FileMode) *os.File
func CopyFile(src string, dst string) error
func MustCopyFile(src string, dst string)
func Remove(fpath string) error
func MustRemove(fpath string)
func QuietRemove(fpath string)
func DeleteIfExist(fpath string) error
func DeleteIfFileExist(fpath string) error
func Unzip(archive, targetDir string) (err error)
```

#### FsUtil Usage

**files finder:**

```go
package main

import (
	"fmt"
	"os"

	"github.com/gookit/goutil/fsutil"
)

func main() {
	f := fsutil.EmptyFinder()

	f.
		AddDir("./testdata").
		AddFile("finder.go").
		NoDotFile().
		// NoDotDir().
		Find().
		Each(func(filePath string) {
			fmt.Println(filePath)
		})

	fsutil.NewFinder([]string{"./testdata"}).
		AddFile("finder.go").
		NoDotDir().
		EachStat(func(fi os.FileInfo, filePath string) {
			fmt.Println(filePath, "=>", fi.ModTime())
		})
}
```

### JSON

> Package `github.com/gookit/goutil/jsonutil`

```go
// source at jsonutil/jsonutil.go
func WriteFile(filePath string, data interface{}) error
func ReadFile(filePath string, v interface{}) error
func Pretty(v interface{}) (string, error)
func Encode(v interface{}) ([]byte, error)
func EncodePretty(v interface{}) ([]byte, error)
func EncodeToWriter(v interface{}, w io.Writer) error
func EncodeUnescapeHTML(v interface{}) ([]byte, error)
func Decode(bts []byte, ptr interface{}) error
func DecodeString(str string, ptr interface{}) error
func DecodeReader(r io.Reader, ptr interface{}) error
func StripComments(src string) string
```

### Map

> Package `github.com/gookit/goutil/maputil`

```go
// source at maputil/convert.go
func KeyToLower(src map[string]string) map[string]string
func ToStringMap(src map[string]interface{}) map[string]string
func HttpQueryString(data map[string]interface{}) string
func ToString(mp map[string]interface{}) string
// source at maputil/maputil.go
func MergeStringMap(src, dst map[string]string, ignoreCase bool) map[string]string
func GetByPath(key string, mp map[string]interface{}) (val interface{}, ok bool)
func Keys(mp interface{}) (keys []string)
func Values(mp interface{}) (values []interface{})
```

### Math/Number

> Package `github.com/gookit/goutil/mathutil`

```go
// source at mathutil/convert.go
func Int(in interface{}) (int, error)
func MustInt(in interface{}) int
func IntOrPanic(in interface{}) int
func ToInt(in interface{}) (iVal int, err error)
func Uint(in interface{}) (uint64, error)
func MustUint(in interface{}) uint64
func ToUint(in interface{}) (u64 uint64, err error)
func Int64(in interface{}) (int64, error)
func MustInt64(in interface{}) int64
func ToInt64(in interface{}) (i64 int64, err error)
func Float(in interface{}) (float64, error)
func ToFloat(in interface{}) (f64 float64, err error)
func FloatOrPanic(in interface{}) float64
func MustFloat(in interface{}) float64
func TryToString(val interface{}, defaultAsErr bool) (str string, err error)
func StringOrPanic(val interface{}) string
func MustString(val interface{}) string
func ToString(val interface{}) (string, error)
func String(val interface{}) string
// source at mathutil/number.go
func IsNumeric(c byte) bool
func Percent(val, total int) float64
func ElapsedTime(startTime time.Time) string
func DataSize(size uint64) string
func HowLongAgo(sec int64) string
// source at mathutil/random.go
func RandomInt(min, max int) int
func RandInt(min, max int) int { return RandomInt(min, max) }
func RandIntWithSeed(min, max int, seed int64) int
func RandomIntWithSeed(min, max int, seed int64) int
```

### Standard

> Package `github.com/gookit/goutil/stdutil`

```go
// source at stdutil/chan.go
func WaitCloseSignals(closer io.Closer) error
func Go(f func() error) error
// source at stdutil/check.go
func ValueIsEmpty(v reflect.Value) bool
func ValueLen(v reflect.Value) int
// source at stdutil/convert.go
func ToString(v interface{}) string
func MustString(v interface{}) string
func TryString(v interface{}) (string, error)
func BaseTypeVal2(v reflect.Value) (value interface{}, err error)
func BaseTypeVal(val interface{}) (value interface{}, err error)
// source at stdutil/gofunc.go
func FuncName(fn interface{}) string
func CutFuncName(fullFcName string) (pkgPath, shortFnName string)
func PkgName(fullFcName string) string
// source at stdutil/stack.go
func GetCallStacks(all bool) []byte
func GetCallerInfo(skip int) string
func SimpleCallersInfo(skip, num int) []string
func GetCallersInfo(skip, max int) []string
// source at stdutil/stdutil.go
func PanicIfErr(err error)
func PanicIf(err error)
func Panicf(format string, v ...interface{})
```

### Struct

> Package `github.com/gookit/goutil/structs`

```go
// source at structs/alias.go
func NewAliases(checker func(alias string)) *Aliases
// source at structs/data.go
func NewMapData() *MapDataStore
// source at structs/structs.go
func ToMap(st interface{}) map[string]interface{}
func TryToMap(st interface{}) (map[string]interface{}, error)
func MustToMap(st interface{}) map[string]interface{}
// source at structs/tags.go
func ParseTags(v interface{}) error
func ParseReflectTags(v reflect.Value) error
func ParseTagValue(str string) maputil.SMap
func ParseTagValueINI(field, str string) (mp maputil.SMap, err error)
```

### String

> Package `github.com/gookit/goutil/strutil`

```go
// source at strutil/bytes.go
func NewByteChanPool(maxSize int, width int, capWidth int) *ByteChanPool
// source at strutil/check.go
func IsNumeric(c byte) bool
func IsAlphabet(char uint8) bool
func IsAlphaNum(c uint8) bool
func StrPos(s, sub string) int
func BytePos(s string, bt byte) int
func RunePos(s string, ru rune) int
func HasOneSub(s string, subs []string) bool
func HasAllSubs(s string, subs []string) bool
func IsStartsOf(s string, subs []string) bool
func HasOnePrefix(s string, subs []string) bool
func IsStartOf(s, sub string) bool
func IsEndOf(s, sub string) bool
func Len(s string) int
func Utf8len(s string) int
func ValidUtf8String(s string) bool
func IsSpace(c byte) bool
func IsSpaceRune(r rune) bool
func IsEmpty(s string) bool
func IsBlank(s string) bool
func IsNotBlank(s string) bool
func IsBlankBytes(bs []byte) bool
func IsSymbol(r rune) bool
// source at strutil/convert.go
func Join(sep string, ss ...string) string
func Implode(sep string, ss ...string) string
func String(val interface{}) (string, error)
func MustString(in interface{}) string
func ToString(val interface{}) (string, error)
func AnyToString(val interface{}, defaultAsErr bool) (str string, err error)
func Byte2str(b []byte) string
func Byte2string(b []byte) string
func ToBytes(s string) (b []byte)
func ToBool(s string) (bool, error)
func MustBool(s string) bool
func Bool(s string) (bool, error)
func Int(s string) (int, error)
func ToInt(s string) (int, error)
func MustInt(s string) int
func IntOrPanic(s string) int
func Ints(s string, sep ...string) []int
func ToInts(s string, sep ...string) ([]int, error)
func ToIntSlice(s string, sep ...string) (ints []int, err error)
func ToArray(s string, sep ...string) []string
func Strings(s string, sep ...string) []string
func ToStrings(s string, sep ...string) []string
func ToSlice(s string, sep ...string) []string
func ToOSArgs(s string) []string
func MustToTime(s string, layouts ...string) time.Time
func ToTime(s string, layouts ...string) (t time.Time, err error)
// source at strutil/encode.go
func Base64(str string) string
func B64Encode(str string) string
func URLEncode(s string) string
func URLDecode(s string) string
// source at strutil/filter.go
func Trim(s string, cutSet ...string) string
func Ltrim(s string, cutSet ...string) string { return TrimLeft(s, cutSet...) }
func LTrim(s string, cutSet ...string) string { return TrimLeft(s, cutSet...) }
func TrimLeft(s string, cutSet ...string) string
func Rtrim(s string, cutSet ...string) string { return TrimRight(s, cutSet...) }
func RTrim(s string, cutSet ...string) string { return TrimRight(s, cutSet...) }
func TrimRight(s string, cutSet ...string) string
func FilterEmail(s string) string
// source at strutil/format.go
func Lower(s string) string { return strings.ToLower(s) }
func Lowercase(s string) string { return strings.ToLower(s) }
func Upper(s string) string { return strings.ToUpper(s) }
func Uppercase(s string) string { return strings.ToUpper(s) }
func UpperWord(s string) string
func LowerFirst(s string) string
func UpperFirst(s string) string
func Snake(s string, sep ...string) string
func SnakeCase(s string, sep ...string) string
func Camel(s string, sep ...string) string
func CamelCase(s string, sep ...string) string
// source at strutil/id.go
func MicroTimeID() string
func MicroTimeHexID() string
// source at strutil/random.go
func Md5(src interface{}) string
func GenMd5(src interface{}) string
func RandomChars(ln int) string
func RandomCharsV2(ln int) string
func RandomCharsV3(ln int) string
func RandomBytes(length int) ([]byte, error)
func RandomString(length int) (string, error)
// source at strutil/similar_find.go
func NewComparator(src, dst string) *SimilarComparator
func Similarity(s, t string, rate float32) (float32, bool)
// source at strutil/split.go
func Cut(s, sep string) (before string, after string, found bool)
func MustCut(s, sep string) (before string, after string)
func SplitValid(s, sep string) (ss []string) { return Split(s, sep) }
func Split(s, sep string) (ss []string)
func SplitNValid(s, sep string, n int) (ss []string) { return SplitN(s, sep, n) }
func SplitN(s, sep string, n int) (ss []string)
func SplitTrimmed(s, sep string) (ss []string)
func SplitNTrimmed(s, sep string, n int) (ss []string)
func Substr(s string, pos, length int) string
// source at strutil/strutil.go
func Padding(s, pad string, length int, pos uint8) string
func PadLeft(s, pad string, length int) string
func PadRight(s, pad string, length int) string
func Repeat(s string, times int) string
func RepeatRune(char rune, times int) (chars []rune)
func RepeatBytes(char byte, times int) (chars []byte)
func Replaces(str string, pairs map[string]string) string
func PrettyJSON(v interface{}) (string, error)
func RenderTemplate(input string, data interface{}, fns template.FuncMap, isFile ...bool) string
func RenderText(input string, data interface{}, fns template.FuncMap, isFile ...bool) string
```

### System

> Package `github.com/gookit/goutil/sysutil`

```go
// source at sysutil/exec.go
func QuickExec(cmdLine string, workDir ...string) (string, error)
func ExecLine(cmdLine string, workDir ...string) (string, error)
func ExecCmd(binName string, args []string, workDir ...string) (string, error)
func ShellExec(cmdLine string, shells ...string) (string, error)
func FindExecutable(binName string) (string, error)
func Executable(binName string) (string, error)
func HasExecutable(binName string) bool
// source at sysutil/sysenv.go
func Hostname() string
func IsWin() bool
func IsWindows() bool
func IsMac() bool
func IsLinux() bool
func IsMSys() bool
func IsConsole(out io.Writer) bool
func IsTerminal(fd uintptr) bool
func StdIsTerminal() bool
func CurrentShell(onlyName bool) (path string)
func HasShellEnv(shell string) bool
func IsShellSpecialVar(c uint8) bool
// source at sysutil/sysutil.go
func Workdir() string
func BinDir() string
func BinFile() string
// source at sysutil/sysutil_nonwin.go
func Kill(pid int, signal syscall.Signal) error
func ProcessExists(pid int) bool
// source at sysutil/user.go
func MustFindUser(uname string) *user.User
func LoginUser() *user.User
func CurrentUser() *user.User
func UserHomeDir() string
func UHomeDir() string
func HomeDir() string
func UserDir(subPath string) string
func UserCacheDir(subPath string) string
func UserConfigDir(subPath string) string
func ExpandPath(path string) string
// source at sysutil/user_nonwin.go
func ChangeUserByName(newUname string) (err error)
func ChangeUserUidGid(newUid int, newGid int) (err error)
```

### Testing

> Package `github.com/gookit/goutil/testutil`

```go
// source at testutil/httpmock.go
func NewHttpRequest(method, path string, data *MD) *http.Request
func MockRequest(h http.Handler, method, path string, data *MD) *httptest.ResponseRecorder
// source at testutil/testutil.go
func DiscardStdout() error
func ReadOutput() (s string)
func RewriteStdout()
func RestoreStdout() (s string)
func RewriteStderr()
func RestoreStderr() (s string)
func MockEnvValue(key, val string, fn func(nv string))
func MockEnvValues(kvMap map[string]string, fn func())
func MockOsEnvByText(envText string, fn func())
func MockOsEnv(mp map[string]string, fn func())
```

### Timex

> Package `github.com/gookit/goutil/timex`

```go
// source at timex/timex.go
func Now() *TimeX
func New(t time.Time) *TimeX
func Wrap(t time.Time) *TimeX
func FromTime(t time.Time) *TimeX
func Local() *TimeX
func FromUnix(sec int64) *TimeX
func FromDate(s string, template ...string) (*TimeX, error)
func FromString(s string, layouts ...string) (*TimeX, error)
func LocalByName(tzName string) *TimeX
func SetLocalByName(tzName string) error
// source at timex/util.go
func NowUnix() int64
func Format(t time.Time) string
func FormatBy(t time.Time, layout string) string
func Date(t time.Time, template string) string
func DateFormat(t time.Time, template string) string
func FormatByTpl(t time.Time, template string) string
func FormatUnix(sec int64) string
func FormatUnixBy(sec int64, layout string) string
func FormatUnixByTpl(sec int64, template string) string
func NowAddDay(day int) time.Time
func NowAddHour(hour int) time.Time
func NowAddMinutes(minutes int) time.Time
func NowAddSeconds(seconds int) time.Time
func AddDay(t time.Time, day int) time.Time
func AddHour(t time.Time, hour int) time.Time
func AddMinutes(t time.Time, minutes int) time.Time
func AddSeconds(t time.Time, seconds int) time.Time
func HourStart(t time.Time) time.Time
func HourEnd(t time.Time) time.Time
func DayStart(t time.Time) time.Time
func DayEnd(t time.Time) time.Time
func NowHourStart() time.Time
func TodayStart() time.Time
func TodayEnd() time.Time
func HowLongAgo(sec int64) string
func ToLayout(template string) string
```
#### Timex Usage

**Create timex instance**

```go
now := timex.Now()

// from time.Time
tx := timex.New(time.Now())
tx := timex.FromTime(time.Now())

// from time unix
tx := timex.FromUnix(1647411580)
```

Create from datetime string:

```go
// auto match layout by datetime
tx, err  := timex.FromString("2022-04-20 19:40:34")
// custom set the datetime layout
tx, err  := timex.FromString("2022-04-20 19:40:34", "2006-01-02 15:04:05")
// use date template as layout
tx, err  := timex.FromDate("2022-04-20 19:40:34", "Y-M-D H:I:S")
```

**Use timex instance**

```go
tx := timex.Now()
```

change time:

```go
tx.Yesterday()
tx.Tomorrow()

tx.DayStart() // get time at Y-M-D 00:00:00
tx.DayEnd() // get time at Y-M-D 23:59:59
tx.HourStart() // get time at Y-M-D H:00:00
tx.HourEnd() // get time at Y-M-D H:59:59

tx.AddDay(2)
tx.AddHour(1)
tx.AddMinutes(15)
tx.AddSeconds(120)
```

compare time:

```go
// before compare
tx.IsBefore(u time.Time)
tx.IsBeforeUnix(1647411580)
// after compare
tx.IsAfter(u time.Time)
tx.IsAfterUnix(1647411580)
```

**Helper functions**

```go
ts := timex.NowUnix() // current unix timestamp

t := NowAddDay(1) // from now add 1 day
t := NowAddHour(1) // from now add 1 hour
t := NowAddMinutes(3) // from now add 3 minutes
t := NowAddSeconds(180) // from now add 180 seconds
```

**Convert time to date by template**

```text
Template Vars:
 Y,y - year
  Y - year 2006
  y - year 06
 M,m - month 01
 D,d - day 02
 H,h - hour 15
 I,i - minute 04
 S,s - second 05
```

Examples, use timex:

```go
now := timex.Now()
date := now.DateFormat("Y-M-D H:i:s") // Output: 2022-04-20 19:40:34
date = now.DateFormat("y-M-D H:i:s") // Output: 22-04-20 19:40:34
```

Format time.Time:

```go
now := time.Now()
date := timex.DateFormat(now, "Y-M-D H:i:s") // Output: 2022-04-20 19:40:34
```

More usage:

```go
ts := timex.NowUnix() // current unix timestamp

date := FormatUnix(ts, "2006-01-02 15:04:05") // Get: 2022-04-20 19:40:34
date := FormatUnixByTpl(ts, "Y-M-D H:I:S") // Get: 2022-04-20 19:40:34
```

## Code Check & Testing

```bash
gofmt -w -l ./
golint ./...
go test ./...
```

## Gookit packages

  - [gookit/ini](https://github.com/gookit/ini) Go config management, use INI files
  - [gookit/rux](https://github.com/gookit/rux) Simple and fast request router for golang HTTP
  - [gookit/gcli](https://github.com/gookit/gcli) Build CLI application, tool library, running CLI commands
  - [gookit/slog](https://github.com/gookit/slog) Lightweight, easy to extend, configurable logging library written in Go
  - [gookit/color](https://github.com/gookit/color) A command-line color library with true color support, universal API methods and Windows support
  - [gookit/event](https://github.com/gookit/event) Lightweight event manager and dispatcher implements by Go
  - [gookit/cache](https://github.com/gookit/cache) Generic cache use and cache manager for golang. support File, Memory, Redis, Memcached.
  - [gookit/config](https://github.com/gookit/config) Go config management. support JSON, YAML, TOML, INI, HCL, ENV and Flags
  - [gookit/filter](https://github.com/gookit/filter) Provide filtering, sanitizing, and conversion of golang data
  - [gookit/validate](https://github.com/gookit/validate) Use for data validation and filtering. support Map, Struct, Form data
  - [gookit/goutil](https://github.com/gookit/goutil) Some utils for the Go: string, array/slice, map, format, cli, env, filesystem, test and more
  - More, please see https://github.com/gookit

## License

[MIT](LICENSE)
