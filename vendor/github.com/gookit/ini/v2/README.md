# INI

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/gookit/ini?style=flat-square)
[![GitHub tag (latest SemVer)](https://img.shields.io/github/tag/gookit/ini)](https://github.com/gookit/ini)
[![Coverage Status](https://coveralls.io/repos/github/gookit/ini/badge.svg?branch=master)](https://coveralls.io/github/gookit/ini?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/gookit/ini)](https://goreportcard.com/report/github.com/gookit/ini)
[![Unit-Tests](https://github.com/gookit/ini/actions/workflows/go.yml/badge.svg)](https://github.com/gookit/ini)
[![Go Reference](https://pkg.go.dev/badge/github.com/gookit/ini/v2.svg)](https://pkg.go.dev/github.com/gookit/ini/v2)

`INI` contents parser by Golang, `INI` config data management.

> **[中文说明](README.zh-CN.md)**

## Features

- Easy to use(get: `Int` `Int64` `Bool` `String` `StringMap` ..., set: `Set`)
- Support multi file, data load
- Support for decode data to struct
- Support data override merge
- Support parse ENV variable
- Support comments start with  `;` `#`, multi line comments `/* .. */`
- Support multi line value with `"""` or `'''`
- Complete unit test(coverage > 90%)
- Support variable reference, default compatible with Python's configParser format `%(VAR)s`

### [Parser](./parser)

Package `parser` is a Parser for parse INI format content to golang data

### [Dotenv](./dotenv)

Package `dotenv` that supports importing ENV data from files (eg `.env`)

## More formats

If you want more support for file content formats, recommended use `gookit/config`

- [gookit/config](https://github.com/gookit/config) - Support multi formats: `JSON`(default), `INI`, `YAML`, `TOML`, `HCL`

## GoDoc

- [godoc](https://pkg.go.dev/github.com/gookit/ini)

## Install

```bash
go get github.com/gookit/ini/v2
```

## Usage

- example data(`testdata/test.ini`):

```ini
# comments
name = inhere
age = 50
debug = true
hasQuota1 = 'this is val'
hasQuota2 = "this is val1"
can2arr = val0,val1,val2
shell = ${SHELL}
noEnv = ${NotExist|defValue}
nkey = val in default section

; comments
[sec1]
key = val0
some = value
stuff = things
varRef = %(nkey)s
```

### Load data

```go
package main

import (
	"github.com/gookit/ini/v2"
)

// go run ./examples/demo.go
func main() {
	// config, err := ini.LoadFiles("testdata/tesdt.ini")
	// LoadExists will ignore not exists file
	err := ini.LoadExists("testdata/test.ini", "not-exist.ini")
	if err != nil {
		panic(err)
	}

	// load more, will override prev data by key
	err = ini.LoadStrings(`
age = 100
[sec1]
newK = newVal
some = change val
`)
	// fmt.Printf("%v\n", config.Data())
}
```

### Read data

- Get integer

```go
age := ini.Int("age")
fmt.Print(age) // 100
```

- Get bool

```go
val := ini.Bool("debug")
fmt.Print(val) // true
```

- Get string

```go
name := ini.String("name")
fmt.Print(name) // inhere
```

- Get section data(string map)

```go
val := ini.StringMap("sec1")
fmt.Println(val) 
// map[string]string{"key":"val0", "some":"change val", "stuff":"things", "newK":"newVal"}
```

- Value is ENV var

```go
value := ini.String("shell")
fmt.Printf("%q", value)  // "/bin/zsh"
```

- **Get value by key path**

```go
value := ini.String("sec1.key")
fmt.Print(value) // val0
```

- Use var refer

```go
value := ini.String("sec1.varRef")
fmt.Printf("%q", value) // "val in default section"
```

- Set new value

```go
// set value
ini.Set("name", "new name")
name = ini.String("name")
fmt.Printf("%q", name) // "new name"
```

## Mapping data to struct

```go
type User struct {
	Name string
	Age int
}

user := &User{}
ini.MapStruct(ini.DefSection(), user)

dump.P(user)
```

Special, mapping all data:

```go
ini.MapStruct("", ptr)
```

## Variable reference resolution

```ini
[portal] 
url = http://%(host)s:%(port)s/api
host = localhost 
port = 8080
```

If variable resolution is enabled，will parse `%(host)s` and replace it：

```go
cfg := ini.New()
// enable ParseVar
cfg.WithOptions(ini.ParseVar)

fmt.Print(cfg.MustString("portal.url"))
// OUT: 
// http://localhost:8080/api 
```

## Available options

```go
type Options struct {
	// set to read-only mode. default False
	Readonly bool
	// parse ENV var name. default True
	ParseEnv bool
	// parse variable reference "%(varName)s". default False
	ParseVar bool

	// var left open char. default "%("
	VarOpen string
	// var right close char. default ")s"
	VarClose string

	// ignore key name case. default False
	IgnoreCase bool
	// default section name. default "__default"
	DefSection string
	// sep char for split key path. default ".", use like "section.subKey"
	SectionSep string
}
```

Setting options for default instance:

```go
ini.WithOptions(ini.ParseEnv,ini.ParseVar)
```

Setting options with new instance:

```go
cfg := ini.New()
cfg.WithOptions(ini.ParseEnv, ini.ParseVar, func (opts *Options) {
	opts.SectionSep = ":"
	opts.DefSection = "default"
})
```

## Dotenv

Package `dotenv` that supports importing data from files (eg `.env`) to ENV

### Usage

```go
err := dotenv.Load("./", ".env")
// err := dotenv.LoadExists("./", ".env")

val := dotenv.Get("ENV_KEY")
// Or use 
// val := os.Getenv("ENV_KEY")

// get int value
intVal := dotenv.Int("LOG_LEVEL")

// with default value
val := dotenv.Get("ENV_KEY", "default value")
```

## Tests

- go tests with cover

```bash
go test ./... -cover
```

- run lint by GoLint

```bash
golint ./...
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

## Related

- [go-ini/ini](https://github.com/go-ini/ini) ini parser and config manage
- [dombenson/go-ini](https://github.com/dombenson/go-ini) ini parser and config manage

## License

**MIT**
