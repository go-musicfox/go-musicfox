# Go Util

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/gookit/goutil?style=flat-square)
[![GitHub tag (latest SemVer)](https://img.shields.io/github/tag/gookit/goutil)](https://github.com/gookit/goutil)
[![Go Report Card](https://goreportcard.com/badge/github.com/gookit/goutil)](https://goreportcard.com/report/github.com/gookit/goutil)
[![Unit-Tests](https://github.com/gookit/goutil/workflows/Unit-Tests/badge.svg)](https://github.com/gookit/goutil/actions)
[![Coverage Status](https://coveralls.io/repos/github/gookit/goutil/badge.svg?branch=master)](https://coveralls.io/github/gookit/goutil?branch=master)
[![Go Reference](https://pkg.go.dev/badge/github.com/gookit/goutil.svg)](https://pkg.go.dev/github.com/gookit/goutil)

💪 Useful utils package for the Go: int, string, array/slice, map, error, time, format, CLI, ENV, filesystem, system, testing and more.

> **[中文说明](README.zh-CN.md)**

**Basic packages:**

- [`arrutil`](./arrutil): Array/Slice util functions. eg: check, convert, formatting, enum, collections
- [`cliutil`](./cliutil) Command-line util functions. eg: colored print, read input, exec command
- [`envutil`](./envutil) ENV util for current runtime env information. eg: get one, get info, parse var
- [`fmtutil`](./fmtutil) Format data util functions. eg: data, size, time
- [`fsutil`](./fsutil) Filesystem util functions, quick create, read and write file. eg: file and dir check, operate
- [`jsonutil`](./jsonutil) some util functions for quick read, write, encode, decode JSON data.
- [`maputil`](./maputil) Map data util functions. eg: convert, sub-value get, simple merge
- [`mathutil`](./mathutil) Math(int, number) util functions. eg: convert, math calc, random
- `netutil` Network util functions
- [`reflects`](./reflects) Provide extends reflect util functions.
- [`stdutil`](./stdutil) Provide some commonly std util functions.
- [`structs`](./structs) Provide some extends util functions for struct. eg: tag parse, struct data init
- [`strutil`](./strutil) String util functions. eg: bytes, check, convert, encode, format and more
- [`sysutil`](./sysutil) System util functions. eg: sysenv, exec, user, process
  - [process](./sysutil/process) Provide some process handle util functions.

**Advance packages:**

- [`cflag`](./cflag):  Wraps and extends go `flag.FlagSet` to build simple command line applications
- cli util:
  - [cmdline](./cliutil/cmdline) Provide cmdline parse, args build to cmdline
- [`dump`](./dump): GO value printing tool. print slice, map will auto wrap each element and display the call location
- [`errorx`](./errorx) Provide an enhanced error implements for go, allow with stacktrace and wrap another error.
- net util:
  - [httpreq](netutil/httpreq) An easier-to-use HTTP client that wraps http.Client
- string util:
  - [textscan](strutil/textscan) Implemented a parser that quickly scans and analyzes text content. It can be used to parse INI, Properties and other formats
- sys util:
  - [clipboard](./sysutil/clipboard) Provide a simple clipboard read and write operations.
  - [cmdr](./sysutil/cmdr) Provide for quick build and run a cmd, batch run multi cmd tasks
- [`testutil`](./testutil) Test help util functions. eg: http test, mock ENV value
  - [assert](./testutil/assert) Asserts functions for help testing
- [`timex`](./timex) Provides an enhanced time.Time implementation. Add more commonly used functional methods
  - such as: DayStart(), DayAfter(), DayAgo(), DateFormat() and more.

## Go Doc

Please see [Go doc](https://pkg.go.dev/github.com/gookit/goutil)

## Install

```shell
go get github.com/gookit/goutil
```

## Usage

```go
// github.com/gookit/goutil
is.True(goutil.IsEmpty(nil))
is.False(goutil.IsEmpty("abc"))

is.True(goutil.IsEqual("a", "a"))
is.True(goutil.IsEqual([]string{"a"}, []string{"a"}))
is.True(goutil.IsEqual(23, 23))

is.True(goutil.Contains("abc", "a"))
is.True(goutil.Contains([]string{"abc", "def"}, "abc"))
is.True(goutil.Contains(map[int]string{2: "abc", 4: "def"}, 4))

// convert type
str = goutil.String(23) // "23"
iVal = goutil.Int("-2") // 2
i64Val = goutil.Int64("-2") // -2
u64Val = goutil.Uint("2") // 2
```

## Packages

### Array/Slice

> Package `github.com/gookit/goutil/arrutil`

```go
// source at arrutil/arrutil.go
func Reverse(ss []string)
func StringsRemove(ss []string, s string) []string
func StringsFilter(ss []string, filter ...func(s string) bool) []string
func StringsMap(ss []string, mapFn func(s string) string) []string
func TrimStrings(ss []string, cutSet ...string) []string
// source at arrutil/check.go
func IntsHas(ints []int, val int) bool
func Int64sHas(ints []int64, val int64) bool
func InStrings(elem string, ss []string) bool { return StringsHas(ss, elem) }
func StringsHas(ss []string, val string) bool
func HasValue(arr, val any) bool
func Contains(arr, val any) bool
func NotContains(arr, val any) bool
// source at arrutil/collection.go
func TwowaySearch(data any, item any, fn Comparer) (int, error)
func MakeEmptySlice(itemType reflect.Type) any
func CloneSlice(data any) any
func Excepts(first, second any, fn Comparer) any
func Intersects(first any, second any, fn Comparer) any
func Union(first, second any, fn Comparer) any
func Find(source any, fn Predicate) (any, error)
func FindOrDefault(source any, fn Predicate, defaultValue any) any
func TakeWhile(data any, fn Predicate) any
func ExceptWhile(data any, fn Predicate) any
// source at arrutil/convert.go
func JoinStrings(sep string, ss ...string) string
func StringsJoin(sep string, ss ...string) string
func StringsToInts(ss []string) (ints []int, err error)
func MustToStrings(arr any) []string
func StringsToSlice(ss []string) []any
func ToInt64s(arr any) (ret []int64, err error)
func MustToInt64s(arr any) []int64
func SliceToInt64s(arr []any) []int64
func ToStrings(arr any) (ret []string, err error)
func SliceToStrings(arr []any) []string
func AnyToString(arr any) string
func SliceToString(arr ...any) string { return ToString(arr) }
func ToString(arr []any) string
func JoinSlice(sep string, arr ...any) string
// source at arrutil/format.go
func NewFormatter(arr any) *ArrFormatter
func FormatIndent(arr any, indent string) string
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


### Cflag

> Package `github.com/gookit/goutil/cflag`

```go
// source at cflag/app.go
func NewApp(fns ...func(app *App)) *App
func NewCmd(name, desc string) *Cmd
// source at cflag/cflag.go
func SetDebug(open bool)
func New(fns ...func(c *CFlags)) *CFlags
func NewEmpty(fns ...func(c *CFlags)) *CFlags
func WithDesc(desc string) func(c *CFlags)
func WithVersion(version string) func(c *CFlags)
// source at cflag/optarg.go
func NewArg(name, desc string, required bool) *FlagArg
// source at cflag/util.go
func IsZeroValue(opt *flag.Flag, value string) (bool, bool)
func AddPrefix(name string) string
func AddPrefixes(name string, shorts []string) string
func AddPrefixes2(name string, shorts []string, nameAtEnd bool) string
func SplitShortcut(shortcut string) []string
func FilterNames(names []string) []string
func IsFlagHelpErr(err error) bool
func WrapColorForCode(s string) string
func ReplaceShorts(args []string, shortsMap map[string]string) []string
```
#### `cflag` Usage

`cflag` usage please see [cflag/README.md](cflag/README.md)


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
func BuildOptionHelpName(names []string) string
func ShellQuote(s string) string
func OutputLines(output string) []string
func FirstLine(output string) string
// source at cliutil/color_print.go
func Redp(a ...any) { color.Red.Print(a...) }
func Redf(format string, a ...any) { color.Red.Printf(format, a...) }
func Redln(a ...any) { color.Red.Println(a...) }
func Bluep(a ...any) { color.Blue.Print(a...) }
func Bluef(format string, a ...any) { color.Blue.Printf(format, a...) }
func Blueln(a ...any) { color.Blue.Println(a...) }
func Cyanp(a ...any) { color.Cyan.Print(a...) }
func Cyanf(format string, a ...any) { color.Cyan.Printf(format, a...) }
func Cyanln(a ...any) { color.Cyan.Println(a...) }
func Grayp(a ...any) { color.Gray.Print(a...) }
func Grayf(format string, a ...any) { color.Gray.Printf(format, a...) }
func Grayln(a ...any) { color.Gray.Println(a...) }
func Greenp(a ...any) { color.Green.Print(a...) }
func Greenf(format string, a ...any) { color.Green.Printf(format, a...) }
func Greenln(a ...any) { color.Green.Println(a...) }
func Yellowp(a ...any) { color.Yellow.Print(a...) }
func Yellowf(format string, a ...any) { color.Yellow.Printf(format, a...) }
func Yellowln(a ...any) { color.Yellow.Println(a...) }
func Magentap(a ...any) { color.Magenta.Print(a...) }
func Magentaf(format string, a ...any) { color.Magenta.Printf(format, a...) }
func Magentaln(a ...any) { color.Magenta.Println(a...) }
func Infop(a ...any) { color.Info.Print(a...) }
func Infof(format string, a ...any) { color.Info.Printf(format, a...) }
func Infoln(a ...any) { color.Info.Println(a...) }
func Successp(a ...any) { color.Success.Print(a...) }
func Successf(format string, a ...any) { color.Success.Printf(format, a...) }
func Successln(a ...any) { color.Success.Println(a...) }
func Errorp(a ...any) { color.Error.Print(a...) }
func Errorf(format string, a ...any) { color.Error.Printf(format, a...) }
func Errorln(a ...any) { color.Error.Println(a...) }
func Warnp(a ...any) { color.Warn.Print(a...) }
func Warnf(format string, a ...any) { color.Warn.Printf(format, a...) }
func Warnln(a ...any) { color.Warn.Println(a...) }
// source at cliutil/info.go
func Workdir() string
func BinDir() string
func BinFile() string
func BinName() string
func GetTermSize(refresh ...bool) (w int, h int)
// source at cliutil/read.go
func ReadInput(question string) (string, error)
func ReadLine(question string) (string, error)
func ReadFirst(question string) (string, error)
func ReadFirstByte(question string) (byte, error)
func ReadFirstRune(question string) (rune, error)
func ReadPassword(question ...string) string
func InputIsYes(ans string) bool
func ByteIsYes(ans byte) bool
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

```shell
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

> More, please see [./cliutil/README](cliutil/README.md)

### Dumper

> Package `github.com/gookit/goutil/dump`

```go
// source at dump/dump.go
func Std() *Dumper
func Reset()
func Config(fn func(opts *Options))
func Print(vs ...any)
func Println(vs ...any)
func Fprint(w io.Writer, vs ...any)
func Format(vs ...any) string
func NoLoc(vs ...any)
func Clear(vs ...any)
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
func VarReplace(s string) string { return os.ExpandEnv(s) }
func VarParse(val string) string
func ParseEnvValue(val string) string
func ParseValue(val string) (newVal string)
func SetEnvs(mp map[string]string)
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
func IsGithubActions() bool
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
// source at errorx/errors.go
func NewR(code int, msg string) ErrorR
func Fail(code int, msg string) ErrorR
func Suc(msg string) ErrorR
// source at errorx/errorx.go
func New(msg string) error
func Newf(tpl string, vars ...any) error
func Errorf(tpl string, vars ...any) error
func With(err error, msg string) error
func Withf(err error, tpl string, vars ...any) error
func WithPrev(err error, msg string) error
func WithPrevf(err error, tpl string, vars ...any) error
func WithStack(err error) error
func Traced(err error) error
func Stacked(err error) error
func WithOptions(msg string, fns ...func(opt *ErrStackOpt)) error
func Wrap(err error, msg string) error
func Wrapf(err error, tpl string, vars ...any) error
// source at errorx/stack.go
func FuncForPC(pc uintptr) *Func
func ResetStdOpt()
func Config(fns ...func(opt *ErrStackOpt))
func SkipDepth(skipDepth int) func(opt *ErrStackOpt)
func TraceDepth(traceDepth int) func(opt *ErrStackOpt)
// source at errorx/util.go
func Raw(msg string) error
func Rawf(tpl string, vars ...any) error
func Cause(err error) error
func Unwrap(err error) error
func Previous(err error) error { return Unwrap(err) }
func ToErrorX(err error) (ex *ErrorX, ok bool)
func Has(err, target error) bool
func Is(err, target error) bool
func To(err error, target any) bool
func As(err error, target any) bool
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
func DataSize(size uint64) string
func SizeToString(size uint64) string { return DataSize(size) }
func StringToByte(sizeStr string) uint64 { return ParseByte(sizeStr) }
func ParseByte(sizeStr string) uint64
func PrettyJSON(v any) (string, error)
func StringsToInts(ss []string) (ints []int, err error)
func ArgsWithSpaces(args []any) (message string)
// source at fmtutil/time.go
func HowLongAgo(sec int64) string
```

### File System

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
// source at fsutil/fsutil.go
func OSTempFile(pattern string) (*os.File, error)
func TempFile(dir, pattern string) (*os.File, error)
func OSTempDir(pattern string) (string, error)
func TempDir(dir, pattern string) (string, error)
func MimeType(path string) (mime string)
func ReaderMimeType(r io.Reader) (mime string)
// source at fsutil/info.go
func Dir(fpath string) string
func PathName(fpath string) string
func Name(fpath string) string
func FileExt(fpath string) string
func Suffix(fpath string) string
func Expand(pathStr string) string
func ExpandPath(pathStr string) string
func Realpath(pathStr string) string
func SplitPath(pathStr string) (dir, name string)
func GlobWithFunc(pattern string, fn func(filePath string) error) (err error)
func FindInDir(dir string, handleFn HandleFunc, filters ...FilterFunc) (e error)
// source at fsutil/operate.go
func Mkdir(dirPath string, perm os.FileMode) error
func MkParentDir(fpath string) error
func DiscardReader(src io.Reader)
func MustReadFile(filePath string) []byte
func MustReadReader(r io.Reader) []byte
func GetContents(in any) []byte
func ReadExistFile(filePath string) []byte
func OpenFile(filepath string, flag int, perm os.FileMode) (*os.File, error)
func QuickOpenFile(filepath string, fileFlag ...int) (*os.File, error)
func OpenReadFile(filepath string) (*os.File, error)
func CreateFile(fpath string, filePerm, dirPerm os.FileMode, fileFlag ...int) (*os.File, error)
func MustCreateFile(filePath string, filePerm, dirPerm os.FileMode) *os.File
func PutContents(filePath string, data any, fileFlag ...int) (int, error)
func WriteFile(filePath string, data any, perm os.FileMode, fileFlag ...int) error
func CopyFile(srcPath, dstPath string) error
func MustCopyFile(srcPath, dstPath string)
func Remove(fPath string) error
func MustRemove(fPath string)
func QuietRemove(fPath string) { _ = os.Remove(fPath) }
func RmIfExist(fPath string) error { return DeleteIfExist(fPath) }
func DeleteIfExist(fPath string) error
func RmFileIfExist(fPath string) error { return DeleteIfFileExist(fPath) }
func DeleteIfFileExist(fPath string) error
func Unzip(archive, targetDir string) (err error)
```

#### FsUtil Usage

**files finder:**

```go
package main

import (
	"fmt"
	"os"

	"github.com/gookit/goutil/fsutil/finder"
)

func main() {
	f := finder.EmptyFinder()

	f.
		AddDir("./testdata").
		AddFile("finder.go").
		NoDotFile().
		// NoDotDir().
		Find().
		Each(func(filePath string) {
			fmt.Println(filePath)
		})

	finder.NewFinder([]string{"./testdata"}).
		AddFile("finder.go").
		NoDotDir().
		EachStat(func(fi os.FileInfo, filePath string) {
			fmt.Println(filePath, "=>", fi.ModTime())
		})
}
```


### JSON Utils

> Package `github.com/gookit/goutil/jsonutil`

```go
// source at jsonutil/jsonutil.go
func WriteFile(filePath string, data any) error
func ReadFile(filePath string, v any) error
func Pretty(v any) (string, error)
func Encode(v any) ([]byte, error)
func EncodePretty(v any) ([]byte, error)
func EncodeToWriter(v any, w io.Writer) error
func EncodeUnescapeHTML(v any) ([]byte, error)
func Decode(bts []byte, ptr any) error
func DecodeString(str string, ptr any) error
func DecodeReader(r io.Reader, ptr any) error
func Mapping(src, dst any) error
func StripComments(src string) string
```

### Map

> Package `github.com/gookit/goutil/maputil`

```go
// source at maputil/check.go
func HasKey(mp, key any) (ok bool)
func HasAllKeys(mp any, keys ...any) (ok bool, noKey any)
// source at maputil/convert.go
func KeyToLower(src map[string]string) map[string]string
func ToStringMap(src map[string]any) map[string]string
func HttpQueryString(data map[string]any) string
func ToString(mp map[string]any) string
func ToString2(mp any) string
func FormatIndent(mp any, indent string) string
func Flatten(mp map[string]any) map[string]any
func FlatWithFunc(mp map[string]any, fn reflects.FlatFunc)
// source at maputil/format.go
func NewFormatter(mp any) *MapFormatter
// source at maputil/get.go
func DeepGet(mp map[string]any, path string) (val any)
func QuietGet(mp map[string]any, path string) (val any)
func GetByPath(path string, mp map[string]any) (val any, ok bool)
func Keys(mp any) (keys []string)
func Values(mp any) (values []any)
// source at maputil/maputil.go
func MergeSMap(src, dst map[string]string, ignoreCase bool) map[string]string
func MergeStringMap(src, dst map[string]string, ignoreCase bool) map[string]string
func MakeByPath(path string, val any) (mp map[string]any)
func MakeByKeys(keys []string, val any) (mp map[string]any)
// source at maputil/setval.go
func SetByPath(mp *map[string]any, path string, val any) error
func SetByKeys(mp *map[string]any, keys []string, val any) (err error)
```

### Math/Number

> Package `github.com/gookit/goutil/mathutil`

```go
// source at mathutil/check.go
func Compare(srcVal, dstVal any, op string) (ok bool)
func CompInt64(srcI64, dstI64 int64, op string) (ok bool)
func CompFloat(srcF64, dstF64 float64, op string) (ok bool)
// source at mathutil/convert.go
func Int(in any) (int, error)
func QuietInt(in any) int
func MustInt(in any) int
func IntOrPanic(in any) int
func IntOrErr(in any) (iVal int, err error)
func ToInt(in any) (iVal int, err error)
func StrInt(s string) int
func Uint(in any) (uint64, error)
func QuietUint(in any) uint64
func MustUint(in any) uint64
func UintOrErr(in any) (uint64, error)
func ToUint(in any) (u64 uint64, err error)
func Int64(in any) (int64, error)
func QuietInt64(in any) int64
func MustInt64(in any) int64
func Int64OrErr(in any) (int64, error)
func ToInt64(in any) (i64 int64, err error)
func QuietFloat(in any) float64
func FloatOrPanic(in any) float64
func MustFloat(in any) float64
func Float(in any) (float64, error)
func FloatOrErr(in any) (float64, error)
func ToFloat(in any) (f64 float64, err error)
func StringOrPanic(val any) string
func MustString(val any) string
func ToString(val any) (string, error)
func StringOrErr(val any) (string, error)
func QuietString(val any) string
func String(val any) string
func TryToString(val any, defaultAsErr bool) (str string, err error)
// source at mathutil/mathutil.go
func MaxFloat(x, y float64) float64
func MaxInt(x, y int) int
func SwapMaxInt(x, y int) (int, int)
func MaxI64(x, y int64) int64
func SwapMaxI64(x, y int64) (int64, int64)
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

### Reflects

> Package `github.com/gookit/goutil/reflects`

```go
// source at reflects/check.go
func HasChild(v reflect.Value) bool
func IsNil(v reflect.Value) bool
func IsFunc(val any) bool
func IsEqual(src, dst any) bool
func IsEmpty(v reflect.Value) bool
func IsEmptyValue(v reflect.Value) bool
// source at reflects/conv.go
func BaseTypeVal(v reflect.Value) (value any, err error)
func ValueByType(val any, typ reflect.Type) (rv reflect.Value, err error)
func ValueByKind(val any, kind reflect.Kind) (rv reflect.Value, err error)
func String(rv reflect.Value) string
func ToString(rv reflect.Value) (str string, err error)
func ValToString(rv reflect.Value, defaultAsErr bool) (str string, err error)
// source at reflects/type.go
func ToBaseKind(kind reflect.Kind) BKind
func ToBKind(kind reflect.Kind) BKind
func TypeOf(v any) Type
// source at reflects/util.go
func Elem(v reflect.Value) reflect.Value
func Indirect(v reflect.Value) reflect.Value
func Len(v reflect.Value) int
func SliceSubKind(typ reflect.Type) reflect.Kind
func SetValue(rv reflect.Value, val any) error
func FlatMap(rv reflect.Value, fn FlatFunc)
// source at reflects/value.go
func Wrap(rv reflect.Value) Value
func ValueOf(v any) Value
```

### Stdio

> Package `github.com/gookit/goutil/stdio`

```go
// source at stdio/ioutil.go
func QuietFprint(w io.Writer, ss ...string)
func QuietFprintf(w io.Writer, tpl string, vs ...any)
func QuietFprintln(w io.Writer, ss ...string)
func QuietWriteString(w io.Writer, ss ...string)
func DiscardReader(src io.Reader)
func MustReadReader(r io.Reader) []byte
// source at stdio/writer.go
func NewWriteWrapper(w io.Writer) *WriteWrapper
```

### Standard

> Package `github.com/gookit/goutil/stdutil`

```go
// source at stdutil/chan.go
func WaitCloseSignals(closer io.Closer) error
func Go(f func() error) error
func SignalHandler(ctx context.Context, signals ...os.Signal) (execute func() error, interrupt func(error))
// source at stdutil/check.go
func IsNil(v any) bool
func IsEmpty(v any) bool
func IsFunc(val any) bool
func IsEqual(src, dst any) bool
func Contains(data, elem any) bool
func IsContains(data, elem any) bool
func CheckContains(data, elem any) (valid, found bool)
func ValueIsEmpty(v reflect.Value) bool
func ValueLen(v reflect.Value) int
// source at stdutil/conv.go
func ToString(v any) string
func MustString(v any) string
func TryString(v any) (string, error)
func BaseTypeVal(val any) (value any, err error)
func BaseTypeVal2(v reflect.Value) (value any, err error)
// source at stdutil/gofunc.go
func FuncName(fn any) string
func CutFuncName(fullFcName string) (pkgPath, shortFnName string)
func PkgName(fullFcName string) string
// source at stdutil/stack.go
func GetCallStacks(all bool) []byte
func GetCallerInfo(skip int) string
func SimpleCallersInfo(skip, num int) []string
func GetCallersInfo(skip, max int) []string
// source at stdutil/stdutil.go
func DiscardE(_ error) {}
func PanicIfErr(err error)
func PanicIf(err error)
func Panicf(format string, v ...any)
func GoVersion() string
```

### Structs

> Package `github.com/gookit/goutil/structs`

```go
// source at structs/alias.go
func NewAliases(checker func(alias string)) *Aliases
// source at structs/convert.go
func ToMap(st any, optFns ...MapOptFunc) map[string]any
func MustToMap(st any, optFns ...MapOptFunc) map[string]any
func TryToMap(st any, optFns ...MapOptFunc) (map[string]any, error)
func StructToMap(st any, optFns ...MapOptFunc) (map[string]any, error)
// source at structs/data.go
func NewData() *Data
// source at structs/setval.go
func InitDefaults(ptr any, optFns ...InitOptFunc) error
func SetValues(ptr any, data map[string]any, optFns ...SetOptFunc) error
// source at structs/structs.go
func MapStruct(srcSt, dstSt any)
// source at structs/tags.go
func ParseTags(st any, tagNames []string) (map[string]maputil.SMap, error)
func ParseReflectTags(rt reflect.Type, tagNames []string) (map[string]maputil.SMap, error)
func NewTagParser(tagNames ...string) *TagParser
func ParseTagValueDefault(field, tagVal string) (mp maputil.SMap, err error)
func ParseTagValueDefine(sep string, defines []string) TagValFunc
func ParseTagValueNamed(field, tagVal string, keys ...string) (mp maputil.SMap, err error)
// source at structs/value.go
func NewValue(val any) *Value
```

### Strings

> Package `github.com/gookit/goutil/strutil`

```go
// source at strutil/bytes.go
func NewBuffer() *Buffer
func NewByteChanPool(maxSize int, width int, capWidth int) *ByteChanPool
// source at strutil/check.go
func NoCaseEq(s, t string) bool
func IsNumChar(c byte) bool
func IsNumeric(s string) bool
func IsAlphabet(char uint8) bool
func IsAlphaNum(c uint8) bool
func StrPos(s, sub string) int
func BytePos(s string, bt byte) int
func HasOneSub(s string, subs []string) bool
func HasAllSubs(s string, subs []string) bool
func IsStartsOf(s string, prefixes []string) bool
func HasOnePrefix(s string, prefixes []string) bool
func HasPrefix(s string, prefix string) bool { return strings.HasPrefix(s, prefix) }
func IsStartOf(s, prefix string) bool { return strings.HasPrefix(s, prefix) }
func HasSuffix(s string, suffix string) bool { return strings.HasSuffix(s, suffix) }
func IsEndOf(s, suffix string) bool { return strings.HasSuffix(s, suffix) }
func IsValidUtf8(s string) bool { return utf8.ValidString(s) }
func IsSpace(c byte) bool
func IsEmpty(s string) bool { return len(s) == 0 }
func IsBlank(s string) bool
func IsNotBlank(s string) bool
func IsBlankBytes(bs []byte) bool
func IsSymbol(r rune) bool
func IsVersion(s string) bool
func Compare(s1, s2, op string) bool
func VersionCompare(v1, v2, op string) bool
// source at strutil/convert.go
func Quote(s string) string { return strconv.Quote(s) }
func Unquote(s string) string
func Join(sep string, ss ...string) string { return strings.Join(ss, sep) }
func JoinList(sep string, ss []string) string { return strings.Join(ss, sep) }
func Implode(sep string, ss ...string) string { return strings.Join(ss, sep) }
func String(val any) (string, error)
func QuietString(in any) string
func MustString(in any) string
func StringOrErr(val any) (string, error)
func ToString(val any) (string, error)
func AnyToString(val any, defaultAsErr bool) (str string, err error)
func Byte2str(b []byte) string
func Byte2string(b []byte) string
func ToBytes(s string) (b []byte)
func ToBool(s string) (bool, error)
func QuietBool(s string) bool
func MustBool(s string) bool
func Bool(s string) (bool, error)
func Int(s string) (int, error)
func ToInt(s string) (int, error)
func QuietInt(s string) int
func MustInt(s string) int
func IntOrPanic(s string) int
func Int64(s string) int64
func QuietInt64(s string) int64
func ToInt64(s string) (int64, error)
func Int64OrErr(s string) (int64, error)
func MustInt64(s string) int64
func Int64OrPanic(s string) int64
func Ints(s string, sep ...string) []int
func ToInts(s string, sep ...string) ([]int, error) { return ToIntSlice(s, sep...) }
func ToIntSlice(s string, sep ...string) (ints []int, err error)
func ToArray(s string, sep ...string) []string { return ToSlice(s, sep...) }
func Strings(s string, sep ...string) []string { return ToSlice(s, sep...) }
func ToStrings(s string, sep ...string) []string { return ToSlice(s, sep...) }
func ToSlice(s string, sep ...string) []string
func ToOSArgs(s string) []string
func MustToTime(s string, layouts ...string) time.Time
func ToTime(s string, layouts ...string) (t time.Time, err error)
func ToDuration(s string) (time.Duration, error)
// source at strutil/encode.go
func EscapeJS(s string) string
func EscapeHTML(s string) string
func AddSlashes(s string) string
func StripSlashes(s string) string
func Md5(src any) string
func MD5(src any) string { return Md5(src) }
func GenMd5(src any) string { return Md5(src) }
func Md5Bytes(src any) []byte
func URLEncode(s string) string
func URLDecode(s string) string
func B32Encode(str string) string
func B32Decode(str string) string
func B64Encode(str string) string
func B64EncodeBytes(src []byte) []byte
func B64Decode(str string) string
func B64DecodeBytes(str string) []byte
func Encoding(base int, typ BaseType) BaseEncoder
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
func Title(s string) string { return strings.ToTitle(s) }
func Lower(s string) string { return strings.ToLower(s) }
func Lowercase(s string) string { return strings.ToLower(s) }
func Upper(s string) string { return strings.ToUpper(s) }
func Uppercase(s string) string { return strings.ToUpper(s) }
func UpperWord(s string) string
func LowerFirst(s string) string
func UpperFirst(s string) string
func SnakeCase(s string, sep ...string) string
func Camel(s string, sep ...string) string { return CamelCase(s, sep...) }
func CamelCase(s string, sep ...string) string
func Indent(s, prefix string) string
func IndentBytes(b, prefix []byte) []byte
// source at strutil/id.go
func MicroTimeID() string
func MicroTimeHexID() string
// source at strutil/padding.go
func Padding(s, pad string, length int, pos PosFlag) string
func PadLeft(s, pad string, length int) string
func PadRight(s, pad string, length int) string
func Resize(s string, length int, align PosFlag) string
func PadBytes(bs []byte, pad byte, length int, pos PosFlag) []byte
func PadBytesLeft(bs []byte, pad byte, length int) []byte
func PadBytesRight(bs []byte, pad byte, length int) []byte
func PadRunes(rs []rune, pad rune, length int, pos PosFlag) []rune
func PadRunesLeft(rs []rune, pad rune, length int) []rune
func PadRunesRight(rs []rune, pad rune, length int) []rune
func Repeat(s string, times int) string
func RepeatRune(char rune, times int) []rune
func RepeatBytes(char byte, times int) []byte
// source at strutil/random.go
func RandomChars(ln int) string
func RandomCharsV2(ln int) string
func RandomCharsV3(ln int) string
func RandomBytes(length int) ([]byte, error)
func RandomString(length int) (string, error)
// source at strutil/runes.go
func RuneIsWord(c rune) bool
func RuneIsLower(c rune) bool
func RuneIsUpper(c rune) bool
func RunePos(s string, ru rune) int { return strings.IndexRune(s, ru) }
func IsSpaceRune(r rune) bool
func Utf8Len(s string) int { return utf8.RuneCountInString(s) }
func Utf8len(s string) int { return utf8.RuneCountInString(s) }
func RuneCount(s string) int { return len([]rune(s)) }
func RuneWidth(r rune) int
func TextWidth(s string) int { return Utf8Width(s) }
func Utf8Width(s string) int { return RunesWidth([]rune(s)) }
func RunesWidth(rs []rune) (w int)
func TextTruncate(s string, w int, tail string) string { return Utf8Truncate(s, w, tail) }
func Utf8Truncate(s string, w int, tail string) string
func TextSplit(s string, w int) []string { return Utf8Split(s, w) }
func Utf8Split(s string, w int) []string
func TextWrap(s string, w int) string { return WidthWrap(s, w) }
func WidthWrap(s string, w int) string
func WordWrap(s string, w int) string
// source at strutil/similar_find.go
func NewComparator(src, dst string) *SimilarComparator
func Similarity(s, t string, rate float32) (float32, bool)
// source at strutil/split.go
func Cut(s, sep string) (before string, after string, found bool)
func MustCut(s, sep string) (before string, after string)
func TrimCut(s, sep string) (string, string)
func SplitValid(s, sep string) (ss []string) { return Split(s, sep) }
func Split(s, sep string) (ss []string)
func SplitNValid(s, sep string, n int) (ss []string) { return SplitN(s, sep, n) }
func SplitN(s, sep string, n int) (ss []string)
func SplitTrimmed(s, sep string) (ss []string)
func SplitNTrimmed(s, sep string, n int) (ss []string)
func Substr(s string, pos, length int) string
func SplitInlineComment(val string) (string, string)
// source at strutil/strutil.go
func Replaces(str string, pairs map[string]string) string
func PrettyJSON(v any) (string, error)
func RenderTemplate(input string, data any, fns template.FuncMap, isFile ...bool) string
func RenderText(input string, data any, fns template.FuncMap, isFile ...bool) string
func WrapTag(s, tag string) string
```

### System

> Package `github.com/gookit/goutil/sysutil`

```go
// source at sysutil/exec.go
func NewCmd(bin string, args ...string) *cmdr.Cmd
func FlushExec(bin string, args ...string) error
func QuickExec(cmdLine string, workDir ...string) (string, error)
func ExecLine(cmdLine string, workDir ...string) (string, error)
func ExecCmd(binName string, args []string, workDir ...string) (string, error)
func ShellExec(cmdLine string, shells ...string) (string, error)
// source at sysutil/stack.go
func CallersInfos(skip, num int, filters ...func(file string, fc *runtime.Func) bool) []*CallerInfo
// source at sysutil/sysenv.go
func IsMSys() bool
func IsConsole(out io.Writer) bool
func IsTerminal(fd uintptr) bool
func StdIsTerminal() bool
func Hostname() string
func CurrentShell(onlyName bool) (path string)
func HasShellEnv(shell string) bool
func IsShellSpecialVar(c uint8) bool
func EnvPaths() []string
func FindExecutable(binName string) (string, error)
func Executable(binName string) (string, error)
func HasExecutable(binName string) bool
func SearchPath(keywords string) []string
// source at sysutil/sysgo.go
func GoVersion() string
func ParseGoVersion(line string) (*GoInfo, error)
func OsGoInfo() (*GoInfo, error)
// source at sysutil/sysutil.go
func Workdir() string
func BinDir() string
func BinFile() string
// source at sysutil/sysutil_nonwin.go
func IsWin() bool
func IsWindows() bool
func IsMac() bool
func IsDarwin() bool
func IsLinux() bool
func Kill(pid int, signal syscall.Signal) error
func ProcessExists(pid int) bool
func OpenBrowser(URL string) error
// source at sysutil/user.go
func MustFindUser(uname string) *user.User
func LoginUser() *user.User
func CurrentUser() *user.User
func UHomeDir() string
func UserHomeDir() string
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
// source at testutil/buffer.go
func NewBuffer() *Buffer
// source at testutil/envmock.go
func MockEnvValue(key, val string, fn func(nv string))
func MockEnvValues(kvMap map[string]string, fn func())
func MockOsEnvByText(envText string, fn func())
func MockOsEnv(mp map[string]string, fn func())
func ClearOSEnv() { os.Clearenv() }
func RevertOSEnv()
func MockCleanOsEnv(mp map[string]string, fn func())
// source at testutil/httpmock.go
func NewHttpRequest(method, path string, data *MD) *http.Request
func MockRequest(h http.Handler, method, path string, data *MD) *httptest.ResponseRecorder
// source at testutil/testutil.go
func DiscardStdout() error
func ReadOutput() (s string)
func RewriteStdout()
func RestoreStdout(printData ...bool) (s string)
func RewriteStderr()
func RestoreStderr(printData ...bool) (s string)
// source at testutil/writer.go
func NewTestWriter() *TestWriter
```

### Timex

> Package `github.com/gookit/goutil/timex`

Provides an enhanced time.Time implementation, and add more commonly used functional methods.
```go
// source at timex/template.go
func ToLayout(template string) string
// source at timex/timex.go
func Now() *Time
func New(t time.Time) *Time
func Wrap(t time.Time) *Time
func FromTime(t time.Time) *Time
func Local() *Time
func FromUnix(sec int64) *Time
func FromDate(s string, template ...string) (*Time, error)
func FromString(s string, layouts ...string) (*Time, error)
func LocalByName(tzName string) *Time
// source at timex/util.go
func NowUnix() int64
func SetLocalByName(tzName string) error
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
func NowHourStart() time.Time
func NowHourEnd() time.Time
func AddDay(t time.Time, day int) time.Time
func AddHour(t time.Time, hour int) time.Time
func AddMinutes(t time.Time, minutes int) time.Time
func AddSeconds(t time.Time, seconds int) time.Time
func HourStart(t time.Time) time.Time
func HourEnd(t time.Time) time.Time
func DayStart(t time.Time) time.Time
func DayEnd(t time.Time) time.Time
func TodayStart() time.Time
func TodayEnd() time.Time
func HowLongAgo(sec int64) string
func ToDuration(s string) (time.Duration, error)
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
tx, err  := timex.FromDate("2022-04-20 19:40:34", "Y-m-d H:I:S")
```

**Use timex instance**

```go
tx := timex.Now()
```

change time:

```go
tx.Yesterday()
tx.Tomorrow()

tx.DayStart() // get time at Y-m-d 00:00:00
tx.DayEnd() // get time at Y-m-d 23:59:59
tx.HourStart() // get time at Y-m-d H:00:00
tx.HourEnd() // get time at Y-m-d H:59:59

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

Template Chars:

```text
 Y,y - year
  Y - year 2006
  y - year 06
 m - month 01-12
 d - day 01-31
 H,h - hour
  H - hour 00-23
  h - hour 01-12
 I,i - minute
  I - minute 00-59
  i - minute 0-59
 S,s - second
  S - second 00-59
  s - second 0-59
```

> More, please see [char map](./timex/template.go)

Examples, use timex:

```go
tx := timex.Now()
date := tx.DateFormat("Y-m-d H:I:S") // Output: 2022-04-20 19:09:03
date = tx.DateFormat("y-m-d h:i:s") // Output: 22-04-20 07:9:3
```

Format time.Time:

```go
tx := time.Now()
date := timex.DateFormat(tx, "Y-m-d H:I:S") // Output: 2022-04-20 19:40:34
```

More usage:

```go
ts := timex.NowUnix() // current unix timestamp

date := FormatUnix(ts, "2006-01-02 15:04:05") // Get: 2022-04-20 19:40:34
date := FormatUnixByTpl(ts, "Y-m-d H:I:S") // Get: 2022-04-20 19:40:34
```


## Code Check & Testing

```bash
gofmt -w -l ./
golint ./...

# testing
go test -v ./...
go test -v -run ^TestErr$
go test -v -run ^TestErr$ ./testutil/assert/...
```

Testing in docker:

```shell
cd goutil
docker run -ti -v $(pwd):/go/work golang:1.18
root@xx:/go/work# go test ./...
```

## Related

- https://github.com/duke-git/lancet
- https://github.com/samber/lo
- https://github.com/zyedidia/generic
- https://github.com/thoas/go-funk

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

