package edit

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
)

// locateExpr resolves key to the expression carrying comments: a key-value,
// or a table with its own [header] line (including array-of-tables elements
// addressed by index). Exactly one of the returned leaf and table is non-nil
// on success.
func (d *Document) locateExpr(key []string) (*leaf, *table, error) {
	if len(key) == 0 {
		return nil, nil, errors.New("toml/edit: key must have at least one element")
	}
	t := d.root
	i := 0
	for i < len(key) {
		it := t.items[key[i]]
		if it == nil {
			return nil, nil, fmt.Errorf("toml/edit: %q does not exist", pathString(key[:i+1]))
		}
		last := i == len(key)-1
		switch {
		case it.leaf != nil:
			if last {
				return it.leaf, nil, nil
			}
			return nil, nil, fmt.Errorf("toml/edit: %q is inside an inline value and has no comments of its own", pathString(key))
		case it.arr != nil:
			if last {
				return nil, nil, fmt.Errorf("toml/edit: %q is an array of tables; address one of its elements", pathString(key))
			}
			idx, ok := parseIndex(key[i+1])
			if !ok || idx >= len(it.arr) {
				return nil, nil, fmt.Errorf("toml/edit: %q has no element %q", pathString(key[:i+1]), key[i+1])
			}
			t = it.arr[idx]
			i += 2
			if i == len(key) {
				return nil, t, nil
			}
		default:
			t = it.tbl
			i++
			if last {
				if !t.explicit {
					return nil, nil, fmt.Errorf("toml/edit: %q has no header line to carry comments", pathString(key))
				}
				return nil, t, nil
			}
		}
	}
	return nil, nil, errors.New("toml/edit: internal error: key resolution did not terminate")
}

// commentSpan returns the byte range of the full-line comment block attached
// above the located expression. The range is empty when there is none.
func commentSpan(lf *leaf, tb *table) span {
	if lf != nil {
		return span{lf.line.start, lf.exprLine}
	}
	return span{tb.section.start, tb.headerLine}
}

// trailingSpan returns the byte range holding the comment trailing the
// located expression on its last line, including the whitespace separating
// it from the expression. The range is empty when there is none.
func (d *Document) trailingSpan(lf *leaf, tb *table) span {
	var start int
	if lf != nil {
		start = lf.value.end
	} else {
		start = tb.exprEnd
	}
	end := d.lineEnd(start)
	if end > start && d.data[end-1] == '\n' {
		end--
		if end > start && d.data[end-1] == '\r' {
			end--
		}
	}
	return span{start, end}
}

// Comment returns the text of the full-line comment block attached above
// the key-value or table header at key, with the comment markers stripped.
// It returns ok=false when key does not address a key-value or a table with
// a header, and an empty string with ok=true when it does but carries no
// comment.
func (d *Document) Comment(key []string) (string, bool) {
	lf, tb, err := d.locateExpr(key)
	if err != nil {
		return "", false
	}
	s := commentSpan(lf, tb)
	var lines []string
	for _, line := range strings.Split(string(d.data[s.start:s.end]), "\n") {
		line = strings.TrimSuffix(line, "\r")
		line = strings.TrimLeft(line, " \t")
		if line == "" {
			continue // the final split element after the last newline
		}
		line = strings.TrimPrefix(line, "#")
		line = strings.TrimPrefix(line, " ")
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n"), true
}

// SetComment sets the full-line comment block attached above the key-value
// or table header at key, replacing any existing one. Each line of text
// becomes one comment line. An empty text removes the block.
func (d *Document) SetComment(key []string, text string) error {
	lf, tb, err := d.locateExpr(key)
	if err != nil {
		return err
	}
	return d.apply(splice{commentSpan(lf, tb), d.renderCommentBlock(text)})
}

func (d *Document) renderCommentBlock(text string) []byte {
	if text == "" {
		return nil
	}
	eol := d.eol()
	var b []byte
	for _, line := range strings.Split(strings.TrimSuffix(text, "\n"), "\n") {
		line = strings.TrimSuffix(line, "\r")
		if line == "" {
			b = append(b, '#')
		} else {
			b = append(b, "# "...)
			b = append(b, line...)
		}
		b = append(b, eol...)
	}
	return b
}

// TrailingComment returns the text of the comment trailing the key-value or
// table header at key on the same line, with the comment marker stripped.
// The ok semantics are the same as Comment's.
func (d *Document) TrailingComment(key []string) (string, bool) {
	lf, tb, err := d.locateExpr(key)
	if err != nil {
		return "", false
	}
	s := d.trailingSpan(lf, tb)
	b := bytes.TrimLeft(d.data[s.start:s.end], " \t")
	b = bytes.TrimPrefix(b, []byte("#"))
	b = bytes.TrimPrefix(b, []byte(" "))
	return string(b), true
}

// SetTrailingComment sets the comment trailing the key-value or table header
// at key on the same line, replacing any existing one. The text must not
// contain newlines. An empty text removes the comment.
func (d *Document) SetTrailingComment(key []string, text string) error {
	if strings.ContainsAny(text, "\r\n") {
		return errors.New("toml/edit: a trailing comment must be a single line")
	}
	lf, tb, err := d.locateExpr(key)
	if err != nil {
		return err
	}
	var repl []byte
	if text != "" {
		repl = []byte(" # " + text)
	}
	return d.apply(splice{d.trailingSpan(lf, tb), repl})
}
