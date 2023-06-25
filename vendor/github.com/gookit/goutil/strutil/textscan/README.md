# TextScan

Package `textscan` implements text scanner for quickly parse text contents.
Can use for parse like INI, Properties format contents.

## Install

```shell
go get github.com/gookit/goutil/strutil/textscan
```

## Examples

```go
package main

import (
	"fmt"

	"github.com/gookit/goutil/dump"
	"github.com/gookit/goutil/strutil/textscan"
	"github.com/gookit/goutil/testutil/assert"
)

func main() {
	ts := textscan.TextScanner{}
	ts.AddMatchers(
		&textscan.CommentsMatcher{},
		&textscan.KeyValueMatcher{},
	)

	ts.SetInput(`
# comments 1
name = inhere

// comments 2
age = 28

/*
multi line
comments 3
*/
desc = '''
a multi
line string
'''
`)

	data := make(map[string]string)
	err := ts.Each(func(t textscan.Token) {
		fmt.Println("====> Token kind:", t.Kind())
		fmt.Println(t.String())

		if t.Kind() == textscan.TokValue {
			v := t.(*textscan.ValueToken)
			data[v.Key()] = v.Value()
		}
	})

	dump.P(data, err)
}
```

**Output:**

```shell
====> Token kind: Comments
# comments 1
====> Token kind: Value
key: name
value: "inhere"
comments: 
====> Token kind: Comments
// comments 2
====> Token kind: Value
key: age
value: "28"
comments: 
====> Token kind: Comments
/*
multi line
comments 3
*/
====> Token kind: Value
key: desc
value: "\n\na multi\nline string\n"
comments: 

==== Collected data:
map[string]string { #len=3
  "desc": string("
a multi
line string
"), #len=22
  "name": string("inhere"), #len=6
  "age": string("28"), #len=2
},
```

## Projects using `textscan`

`textscan` is used in these projects:

- https://github.com/gookit/ini
- https://github.com/gookit/properties
