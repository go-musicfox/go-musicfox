# INI

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/gookit/ini?style=flat-square)
[![GitHub tag (latest SemVer)](https://img.shields.io/github/tag/gookit/ini)](https://github.com/gookit/ini)
[![GoDoc](https://godoc.org/github.com/gookit/ini?status.svg)](https://pkg.go.dev/github.com/gookit/ini)
[![Coverage Status](https://coveralls.io/repos/github/gookit/ini/badge.svg?branch=master)](https://coveralls.io/github/gookit/ini?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/gookit/ini)](https://goreportcard.com/report/github.com/gookit/ini)
[![Unit-Tests](https://github.com/gookit/ini/actions/workflows/go.yml/badge.svg)](https://github.com/gookit/ini)

INI格式内容解析; 使用INI格式作为配置，配置数据的加载，管理，使用。

> **[EN README](README.md)**

## 功能简介

- 使用简单(获取: `Int` `Int64` `Bool` `String` `StringMap` ..., 设置: `Set` )
- 支持多文件，数据加载
- 支持数据覆盖合并
- 支持将数据绑定到结构体
- 支持解析 `ENV` 变量名
- 支持使用 `;` `#` 注释一行, 同时支持多行注释 `/* .. */`
- 支持使用 `"""` or `'''` 编写多行值
- 支持变量参考引用
  - 默认兼容 Python 的 configParser 格式 `%(VAR)s`
- 完善的单元测试(coverage > 90%)

### [Parser](./parser)

子包 `parser` - 实现了解析 `INI` 格式内容为 Go 数据

### [Dotenv](./dotenv)

子包 `dotenv` - 提供了加载解析 `.env` 文件数据为ENV环境变量

## 更多格式

如果你想要更多文件内容格式的支持，推荐使用 `gookit/config`

- [gookit/config](https://github.com/gookit/config) - 支持多种格式: `JSON`(default), `INI`, `YAML`, `TOML`, `HCL`

## GoDoc

- [doc on gowalker](https://gowalker.org/github.com/gookit/ini)
- [godoc for github](https://pkg.go.dev/github.com/gookit/ini)

## 安装

```bash
go get github.com/gookit/ini/v2
```

## 快速使用

- 示例数据(`testdata/test.ini`):

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

### 载入数据

```go
package main

import (
	"github.com/gookit/ini/v2"
)

// go run ./examples/demo.go
func main() {
	// err := ini.LoadFiles("testdata/tesdt.ini")
	// LoadExists 将忽略不存在的文件
	err := ini.LoadExists("testdata/test.ini", "not-exist.ini")
	if err != nil {
		panic(err)
	}
	
	// 加载更多，相同的键覆盖之前数据
	err = ini.LoadStrings(`
age = 100
[sec1]
newK = newVal
some = change val
`)
	// fmt.Printf("%v\n", ini.Data())
}
```

### 获取数据

- 获取整型

```go
age := ini.Int("age")
fmt.Print(age) // 100
```

- 获取布尔值

```go
val := ini.Bool("debug")
fmt.Print(val) // true
```

- 获取字符串

```go
name := ini.String("name")
fmt.Print(name) // inhere
```

- 获取section数据(string map)

```go
val := ini.StringMap("sec1")
fmt.Println(val) 
// map[string]string{"key":"val0", "some":"change val", "stuff":"things", "newK":"newVal"}
```

- 获取的值是环境变量

```go
value := ini.String("shell")
fmt.Printf("%q", value)  // "/bin/zsh"
```

- 通过key path来直接获取子级值

```go
value := ini.String("sec1.key")
fmt.Print(value) // val0
```

- 支持变量参考

```go
value := ini.String("sec1.varRef")
fmt.Printf("%q", value)  // "val in default section"
```

- 设置新的值

```go
// set value
ini.Set("name", "new name")
name = ini.String("name")
fmt.Printf("%q", name)  // "new name"
```

### 将数据映射到结构

```go
type User struct {
	Name string
	Age int
}

user := &User{}
ini.MapStruct(ini.DefSection(), user)

dump.P(user)
```

特殊的，绑定所有数据：

```go
ini.MapStruct("", ptr)
```

## 变量参考解析

```ini
[portal] 
url = http://%(host)s:%(port)s/Portal
host = localhost 
port = 8080
```

启用变量解析后，将会解析这里的 `%(host)s` 并替换为相应的变量值 `localhost`：

```go
cfg := ini.New()
// 启用变量解析
cfg.WithOptions(ini.ParseVar)

fmt.Print(cfg.String("portal.url"))
// OUT: 
// http://localhost:8080/Portal 
```

## 可用选项

```go
type Options struct {
	// 设置为只读模式. default False
	Readonly bool
	// 解析 ENV 变量名称. default True
	ParseEnv bool
	// 解析变量引用 "%(varName)s". default False
	ParseVar bool

	// 变量左侧字符. default "%("
	VarOpen string
	// 变量右侧字符. default ")s"
	VarClose string

	// 忽略键名称大小写. default False
	IgnoreCase bool
	// 默认的section名称. default "__default"
	DefSection string
	// 路径分隔符，当通过key获取子级值时. default ".", 例如 "section.subKey"
	SectionSep string
}
```

- 应用选项

```go
cfg := ini.New()
cfg.WithOptions(ini.ParseEnv,ini.ParseVar, func (opts *Options) {
	opts.SectionSep = ":"
	opts.DefSection = "default"
})
```

## Dotenv

Package `dotenv` that supports importing data from files (eg `.env`) to ENV

### 使用说明

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

## 测试

- 测试并输出覆盖率

```bash
go test ./... -cover
```

- 运行 GoLint 检查

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

## 相关项目参考

- [go-ini/ini](https://github.com/go-ini/ini) ini parser and config manage
- [dombenson/go-ini](https://github.com/dombenson/go-ini) ini parser and config manage

## License

**MIT**
