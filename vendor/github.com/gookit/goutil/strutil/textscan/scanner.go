// Package textscan Implemented a parser that quickly scans and analyzes text content.
// It can be used to parse INI, Properties and other formats
package textscan

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
)

// ErrScan error on scan or parse contents
type ErrScan struct {
	Msg  string // error message
	Line int    // error line number, start 1
	Text string // text contents on error
}

// Error string
func (e ErrScan) Error() string {
	return fmt.Sprintf("%s. line %d: %q", e.Msg, e.Line, e.Text)
}

// Matcher interface
type Matcher interface {
	// Match text line by kind, if success returns a new Token
	Match(line string, prev Token) (tok Token, err error)
}

// TextScanner struct.
type TextScanner struct {
	in *bufio.Scanner
	// ks map[Kind]string

	// token matchers
	matchers []Matcher
	prevTok  Token

	line int
	next string // not used
	tok  Token
	err  error
}

// NewScanner instance
func NewScanner(in any) *TextScanner {
	ts := &TextScanner{}
	if in != nil {
		ts.SetInput(in)
	}
	return ts
}

// SetInput for scan and parse
func (s *TextScanner) SetInput(in any) {
	// init
	// if s.ks == nil {
	// 	s.ks = make(map[Kind]string, len(kinds))
	// }
	// for kind, name := range kinds {
	// 	s.ks[kind] = name
	// }

	switch typIn := in.(type) {
	case *bufio.Scanner:
		s.in = typIn
	case io.Reader:
		s.in = bufio.NewScanner(typIn)
	case []byte:
		s.in = bufio.NewScanner(bytes.NewReader(typIn))
	case string:
		s.in = bufio.NewScanner(strings.NewReader(typIn))
	default:
		panic("invalid input data for parse")
	}
}

// SetSplit set split func on scan
func (s *TextScanner) SetSplit(fn bufio.SplitFunc) {
	if s.in == nil {
		panic("must be set input before set split func")
	}
	s.in.Split(fn)
}

// AddKind register new kind
func (s *TextScanner) AddKind(k Kind, name string) {
	if !HasKind(k) {
		AddKind(k, name)
	}
}

// AddMatchers register token matchers
func (s *TextScanner) AddMatchers(ms ...Matcher) {
	s.matchers = append(s.matchers, ms...)
}

// Each every token by given func
func (s *TextScanner) Each(fn func(t Token)) error {
	for s.Scan() {
		fn(s.Token())
	}
	return s.err
}

// Scan source input and parsing.
// Can use Token() get current parsed token value
//
// Usage:
//
//	ts := textscan.NewScanner(`source ...`)
//	for ts.Scan() {
//		tok := ts.Token()
//		// do something...
//	}
//	fmt.Println(ts.Err())
func (s *TextScanner) Scan() bool {
	if s.next != "" {
		return s.matchToken(s.next)
	}

	if ok, text := s.ScanNext(); ok {
		return s.matchToken(text)
	}

	s.tok = nil // at end.
	return false
}

func (s *TextScanner) matchToken(text string) (ok bool) {
	for _, m := range s.matchers {
		s.prevTok = s.tok
		tok, err := m.Match(text, s.prevTok)
		if err != nil {
			s.err = ErrScan{Msg: err.Error(), Line: s.line, Text: text}
			return false
		}

		if tok != nil {
			if tok.HasMore() {
				if err := tok.ScanMore(s); err != nil {
					s.err = ErrScan{Msg: err.Error(), Line: s.line, Text: text}
					return false
				}
			}

			s.tok = tok
			return true
		}
	}

	// emtpy line, match next valid token
	if strings.TrimSpace(text) == "" {
		ok, text := s.ScanNext()
		if ok {
			return s.matchToken(text)
		}
		return false // end EOF
	}

	s.err = ErrScan{Msg: "invalid syntax, no matcher available", Line: s.line, Text: text}
	return false
}

// ScanNext advance and fetch next line text
func (s *TextScanner) ScanNext() (ok bool, text string) {
	if s.in.Scan() {
		s.line++
		return true, s.in.Text()
	}
	return
}

// SetNext text for scan and parse
func (s *TextScanner) SetNext(text string) {
	s.next = text
}

// Token get of current scan.
func (s *TextScanner) Token() Token {
	return s.tok
}

// PrevToken get of previous scan.
func (s *TextScanner) PrevToken() Token {
	return s.prevTok
}

// Line on current
func (s *TextScanner) Line() int {
	return s.line
}

// Err get
func (s *TextScanner) Err() error {
	return s.err
}
