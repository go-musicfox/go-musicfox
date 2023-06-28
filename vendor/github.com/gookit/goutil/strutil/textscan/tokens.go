package textscan

import (
	"github.com/gookit/goutil"
	"github.com/gookit/goutil/errorx"
)

// Kind type
type Kind uint8

// String name for kind
func (k Kind) String() string {
	return KindString(k)
}

// builtin defined kinds
const (
	TokInvalid Kind = iota
	TokKey
	TokValue
	TokComments
)

// global kinds
var kinds = map[Kind]string{
	TokInvalid:  "Invalid",
	TokKey:      "Key",
	TokValue:    "Value",
	TokComments: "Comments",
}

// AddKind add global kind to kinds
func AddKind(k Kind, name string) {
	if _, ok := kinds[k]; ok {
		goutil.Panicf("cannot repeat register kind(%d): %s", int(k), name)
	}
	kinds[k] = name
}

// HasKind check
func HasKind(k Kind) bool {
	_, ok := kinds[k]
	return ok
}

// KindString name
func KindString(k Kind) string {
	if name, ok := kinds[k]; ok {
		return name
	}
	return "Invalid"
}

// IsKindToken check
func IsKindToken(k Kind, tok Token) bool {
	if tok != nil {
		return tok.Kind() == k
	}
	return false
}

// LiteToken interface
type LiteToken interface {
	Kind() Kind
	Value() string
	IsValid() bool
}

// Token parser
type Token interface {
	LiteToken
	String() string
	// HasMore is multi line values
	HasMore() bool
	// ScanMore scan multi line values
	ScanMore(ts *TextScanner) error
	MergeSame(tok Token) error
}

// BaseToken struct
type BaseToken struct {
	kind  Kind
	value string

	// Offset int // byte offset, starting at 0
	// Line   int // line number, starting at 1
	// Column int // column number, starting at 1 (character count per line)
}

// Kind type
func (t *BaseToken) Kind() Kind {
	return t.kind
}

// IsValid token
func (t *BaseToken) IsValid() bool {
	return t.kind != TokInvalid
}

// Value of token
func (t *BaseToken) Value() string {
	return t.value
}

// String of token
func (t *BaseToken) String() string {
	if t.kind == TokInvalid {
		return "<Invalid>"
	}
	return t.value
}

// StringToken struct
type StringToken struct {
	BaseToken
}

// NewEmptyToken instance.
// Can use for want skip parse some contents
func NewEmptyToken() *StringToken {
	return &StringToken{}
}

// NewStringToken instance.
func NewStringToken(k Kind, val string) *StringToken {
	return &StringToken{
		BaseToken{kind: k, value: val},
	}
}

// HasMore is multi line values
func (t *StringToken) HasMore() bool {
	return false
}

// ScanMore implements
func (t *StringToken) ScanMore(ts *TextScanner) error {
	return nil
}

// MergeSame implements
func (t *StringToken) MergeSame(tok Token) error {
	return errorx.Raw("cannot merge any token to Invalid token")
}
