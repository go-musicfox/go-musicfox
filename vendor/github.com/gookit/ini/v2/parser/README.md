# INI Parser

This is a parser for parse INI format content to golang data 

## Feature

- Support parse section, array value.
- Support comments start with  `;` `#`
- Support multi line comments `/* .. */`
- Support multi line value with `"""` or `'''`

## Install

```bash
go get github.com/gookit/ini/v2/parser
```

## Usage

```go
package main

import (
	"github.com/gookit/goutil"
	"github.com/gookit/goutil/dump"
	"github.com/gookit/ini/v2/parser"
)

func main() {
	p := parser.New()

	err := p.ParseString(`
# comments 1
name = inhere
age = 28

; comments 2
[sec1]
key = val0
some = value
`)

	goutil.PanicErr(err)
	// dump parsed data and collected comments map
	dump.P(p.ParsedData(), p.Comments())
}
```

Output:

```shell
map[string]map[string]string { #len=2
  "__default": map[string]string { #len=7
    "name": string("inhere"), #len=6
    "age": string("28"), #len=2
  },
  "sec1": map[string]string { #len=3
    "key": string("val0"), #len=4
    "some": string("value"), #len=5
  },
},
# collected comments
map[string]string { #len=2
  "_sec_sec1": string("; comments 2"), #len=12
  "__default_name": string("# comments 1"), #len=12
},
```

## Functions API

```go
func Decode(blob []byte, ptr any) error
func Encode(v any) ([]byte, error)
func EncodeFull(data map[string]any, defSection ...string) (out []byte, err error)
func EncodeLite(data map[string]map[string]string, defSection ...string) (out []byte, err error)
func EncodeSimple(data map[string]map[string]string, defSection ...string) ([]byte, error)
func EncodeWithDefName(v any, defSection ...string) (out []byte, err error)
func IgnoreCase(p *Parser)
func InlineComment(opt *Options)
func NoDefSection(p *Parser)
func WithReplaceNl(opt *Options)
type OptFunc func(opt *Options)
    func WithDefSection(name string) OptFunc
    func WithParseMode(mode parseMode) OptFunc
    func WithTagName(name string) OptFunc
type Options struct{ ... }
    func NewOptions(fns ...OptFunc) *Options
type Parser struct{ ... }
    func New(fns ...OptFunc) *Parser
    func NewFulled(fns ...func(*Parser)) *Parser
    func NewLite(fns ...OptFunc) *Parser
    func NewSimpled(fns ...func(*Parser)) *Parser
    func Parse(data string, mode parseMode, opts ...func(*Parser)) (p *Parser, err error)
```

## Related

- [dombenson/go-ini](https://github.com/dombenson/go-ini) ini parser and config manage
- [go-ini/ini](https://github.com/go-ini/ini) ini parser and config manage

## License

**MIT**
