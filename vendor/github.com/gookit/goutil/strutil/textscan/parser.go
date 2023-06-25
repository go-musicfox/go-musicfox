package textscan

import (
	"bytes"
	"io"
	"strings"
)

// HandleFn for token
type HandleFn func(t Token)

// Parser struct
type Parser struct {
	ts *TextScanner
	// Func for handle tokens
	Func HandleFn
}

// NewParser instance
func NewParser(fn HandleFn) *Parser {
	return &Parser{
		Func: fn,
		ts:   &TextScanner{},
	}
}

// AddMatchers register token matchers
func (p *Parser) AddMatchers(ms ...Matcher) {
	p.ts.AddMatchers(ms...)
}

// Parse input bytes
func (p *Parser) Parse(bs []byte) error {
	return p.ParseFrom(bytes.NewReader(bs))
}

// ParseText input string
func (p *Parser) ParseText(text string) error {
	return p.ParseFrom(strings.NewReader(text))
}

// ParseFrom input reader
func (p *Parser) ParseFrom(r io.Reader) error {
	ts := NewScanner(r)

	for ts.Scan() {
		p.Func(ts.Token())
	}
	return nil
}
