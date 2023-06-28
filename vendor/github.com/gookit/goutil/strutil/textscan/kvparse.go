package textscan

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gookit/goutil/comdef"
	"github.com/gookit/goutil/errorx"
	"github.com/gookit/goutil/strutil"
)

// define special chars consts
const (
	MultiLineValMarkS = "'''"
	MultiLineValMarkD = `"""`
	MultiLineValMarkH = "<<<" // heredoc at start. <<<TXT ... TXT
	MultiLineValMarkQ = "\\"  // at end. eg: properties contents
	MultiLineCmtEnd   = "*/"
	// VarRefStartChars  = "${"
)

// KeyValueMatcher match key-value token.
// Support parse `KEY=VALUE` line text contents.
type KeyValueMatcher struct {
	// Separator string for split key and value, default is "="
	Separator string
	// MergeComments collect previous comments token to value token.
	// If set as True, on each s.Scan() please notice skip TokComments
	MergeComments bool
	// InlineComment parse and split inline comment
	InlineComment bool
	// DisableMultiLine value parse
	DisableMultiLine bool
	// KeyCheckFn set func check key string is valid
	KeyCheckFn func(key string) error
}

// Match text line.
func (m *KeyValueMatcher) Match(text string, prev Token) (Token, error) {
	str := strings.TrimSpace(text)
	ln := len(str)
	if ln == 0 {
		return nil, nil
	}

	if m.Separator == "" {
		m.Separator = comdef.EqualStr
	}

	nodes := strutil.SplitNTrimmed(str, m.Separator, 2)
	if len(nodes) != 2 {
		return nil, nil
	}

	key, val := nodes[0], nodes[1]
	if len(key) == 0 {
		return nil, errors.New("key cannot be empty")
	}

	// check key string.
	if m.KeyCheckFn != nil {
		if err := m.KeyCheckFn(key); err != nil {
			return nil, err
		}
	}

	// handle value
	vln := len(val)
	tok := &ValueToken{
		m:   m,
		key: key,
	}

	tok.kind = TokValue

	// collect prev comments token
	if m.MergeComments && IsKindToken(TokComments, prev) {
		tok.comment = prev
	}

	if m.DisableMultiLine {
		// split inline comments and clear quotes
		if vln > 1 {
			val = m.inlineCommentsAndUnquote(tok, val)
		}

		tok.value = val
		return tok, nil
	}

	// multi line value ended by \
	if vln > 0 && strings.HasSuffix(val, MultiLineValMarkQ) {
		val = val[:vln-1]
		tok.more = true
		tok.mark = MultiLineValMarkQ
		tok.value = val
		tok.values = []string{val}
		return tok, nil
	}

	if vln > 2 {
		// multi line value start
		hasPfx := strutil.HasOnePrefix(val, []string{MultiLineValMarkD, MultiLineValMarkS})
		if hasPfx {
			tok.more = true
			tok.mark = MultiLineValMarkS
			if val[0] == '"' {
				tok.mark = MultiLineValMarkD
			}

			val = val[3:]
			tok.value = val
			tok.values = []string{val}
			return tok, nil
		}

		// split inline comments and clear quotes
		val = m.inlineCommentsAndUnquote(tok, val)
	}

	tok.value = val
	return tok, nil
}

func (m *KeyValueMatcher) inlineCommentsAndUnquote(vt *ValueToken, val string) string {
	if m.InlineComment {
		// split inline comments
		var comment string
		val, comment = strutil.SplitInlineComment(val, true)

		if len(comment) > 0 {
			cmt := NewCommentToken(comment)
			// merge comments token
			if vt.comment != nil {
				_ = vt.comment.MergeSame(cmt)
			} else {
				vt.comment = cmt
			}
		}
	}

	// clear quotes
	if val[0] == '"' || val[0] == '\'' {
		val = strutil.Unquote(val)
	}
	return val
}

// DetectEnd for multi line value
func (m *KeyValueMatcher) DetectEnd(mark, text string) (ok bool, val string) {
	str := strings.TrimSpace(text)
	ln := len(str)

	// multi line value
	if mark == MultiLineValMarkS || mark == MultiLineValMarkD {
		if strings.HasSuffix(str, mark) { // end
			val = str[:ln-3]
			ok = true
		} else {
			val = text // goon
		}
	} else if mark == MultiLineValMarkQ {
		if strings.HasSuffix(str, MultiLineValMarkQ) { // goon
			val += str[:ln-1]
		} else { // end
			val = str
			ok = true
		}
	}

	return
}

// ValueToken contains key and value contents
type ValueToken struct {
	BaseToken
	m *KeyValueMatcher

	more bool
	mark string // end mark for multi line

	// key for token
	key string
	// for multi line value.
	values []string
	// comment for the item
	comment Token
}

// Key name
func (t *ValueToken) Key() string {
	return t.key
}

// Mark for multi line values
func (t *ValueToken) Mark() string {
	return t.mark
}

// Values for multi line values
func (t *ValueToken) Values() []string {
	return t.values
}

// Comment lines string
func (t *ValueToken) Comment() string {
	if t.comment != nil {
		return t.comment.Value()
	}
	return ""
}

// Value text string.
func (t *ValueToken) Value() string {
	if len(t.values) > 0 {
		return strings.Join(t.values, "\n")
	}
	return t.value
}

// HasMore is multi line values
func (t *ValueToken) HasMore() bool {
	return t.more
}

// HasComment for the value
func (t *ValueToken) HasComment() bool {
	return t.comment != nil
}

// MergeSame comments token
func (t *ValueToken) MergeSame(_ Token) error {
	return errors.New("merge value token not allowed")
}

// String of token
func (t *ValueToken) String() string {
	return fmt.Sprintf("key: %s\nvalue: %q\ncomments: %s", t.key, t.Value(), t.Comment())
}

// ErrMLineValueNotEnd error
var ErrMLineValueNotEnd = errors.New("not end of multi line value")

// ScanMore scan multi line values
func (t *ValueToken) ScanMore(ts *TextScanner) error {
	for {
		ok, line := ts.ScanNext()
		if !ok {
			return ErrMLineValueNotEnd
		}

		// detect value end line
		if ok, val := t.m.DetectEnd(t.mark, line); ok {
			t.values = append(t.values, val)
			return nil
		}

		t.values = append(t.values, line)
	}
}

// CommentsMatcher match comments lines.
// will auto merge prev comments token
type CommentsMatcher struct {
	// InlineChars for match inline comments. default is: #
	InlineChars []byte
	// MatchFn for comments line
	// - mark 	useful on multi line comments
	MatchFn func(text string) (ok, more bool, err error)
	// DetectEnd for multi line comments
	DetectEnd func(text string) bool
}

// Match comments token
func (m *CommentsMatcher) Match(text string, prev Token) (Token, error) {
	if m.MatchFn == nil {
		if len(m.InlineChars) == 0 {
			m.InlineChars = []byte{'#'}
		}

		m.MatchFn = func(text string) (ok, more bool, err error) {
			return CommentsDetect(text, m.InlineChars)
		}
	}

	// skip empty line
	if text = strings.TrimSpace(text); text == "" {
		return nil, nil
	}

	ok, more, err := m.MatchFn(text)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}

	tok := &CommentToken{m: m}
	tok.more = more
	tok.kind = TokComments
	tok.value = text

	if more {
		tok.comments = []string{text}
	}

	if IsKindToken(TokComments, prev) {
		if err := tok.MergeSame(prev); err != nil {
			return nil, err
		}
	}

	return tok, nil
}

// CommentsDetect check.
//
// - inlineChars: #
//
// default match:
//
//   - inline #, //
//   - multi line: /*
func CommentsDetect(str string, inlineChars []byte) (ok, more bool, err error) {
	ln := len(str)
	if ln == 0 {
		return
	}

	// match inline comments by prefix char.
	for _, prefix := range inlineChars {
		if str[0] == prefix {
			ok = true
			return
		}
	}

	// match start withs // OR /*
	if str[0] == '/' {
		if ln < 2 {
			err = errors.New("invalid contents")
			return
		}

		if str[1] == '/' {
			ok = true
			return
		}

		// multi line comments start
		if str[1] == '*' {
			ok = true
			more = true

			// end at line
			if strings.HasSuffix(str, MultiLineCmtEnd) {
				more = false
			}
		}
	}
	return
}

// MatchEnd for multi line comments
func (m *CommentsMatcher) MatchEnd(text string) bool {
	if m.DetectEnd == nil {
		m.DetectEnd = CommentsDetectEnd
	}
	return m.DetectEnd(text)
}

// CommentToken struct
type CommentToken struct {
	BaseToken
	m *CommentsMatcher

	more bool
	// mark string // end mark for multi line

	// for multi line comments.
	comments []string
}

// NewCommentToken instance.
func NewCommentToken(val string) *CommentToken {
	tok := &CommentToken{}
	tok.value = val
	tok.kind = TokComments
	return tok
}

// Value fo token
func (t *CommentToken) Value() string {
	if len(t.comments) > 0 {
		return strings.Join(t.comments, "\n")
	}
	return t.value
}

// String for token
func (t *CommentToken) String() string {
	return t.Value()
}

// MergeSame comments token
func (t *CommentToken) MergeSame(tok Token) error {
	if tok == nil {
		return nil
	}

	if tok.Kind() == t.Kind() {
		t.comments = append(t.comments, tok.Value())
		return nil
	}
	return errorx.Rawf("cannot merge %s token to an Comments token", tok.Kind())
}

// HasMore is multi line values
func (t *CommentToken) HasMore() bool {
	return t.more
}

// ErrCommentsNotEnd error
var ErrCommentsNotEnd = errors.New("not end of multi-line comments")

// ScanMore scan multi line values
func (t *CommentToken) ScanMore(ts *TextScanner) error {
	if t.m.DetectEnd == nil {
		t.m.DetectEnd = CommentsDetectEnd
	}

	for {
		ok, line := ts.ScanNext()
		if !ok {
			return ErrCommentsNotEnd
		}

		t.comments = append(t.comments, line)

		// detect comments end line
		if t.m.DetectEnd(line) {
			return nil
		}
	}
}

// CommentsDetectEnd multi line comments end
func CommentsDetectEnd(line string) bool {
	return strings.HasSuffix(line, MultiLineCmtEnd)
}
