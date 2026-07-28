package edit

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/pelletier/go-toml/v2/unstable"
)

// span is a range of bytes in the document.
type span struct {
	start, end int
}

// leaf is a key-value in the document.
type leaf struct {
	// line covers the whole key-value expression: from the start of the
	// full-line comments attached above it to past the end of its last line
	// (including any comment trailing the value).
	line span
	// value covers exactly the bytes of the value.
	value span
	// exprLine is the start of the line the key-value itself is on; the
	// bytes between line.start and exprLine are the attached comments.
	exprLine int
}

// table is a table of the document: the root, a [table] section, an element
// of an [[array table]], or a table implied by a header or a dotted key.
type table struct {
	path   []string
	parent *table
	root   bool
	// explicit tables have their own section in the document: a [header] (or
	// [[header]] for array-table elements) followed by key-values.
	explicit bool
	// viaArray is set on array-table elements and everything below them:
	// their absolute path is only meaningful relative to the last element of
	// each array, so new sub-tables cannot be created with [headers].
	viaArray   bool
	section    span // for explicit tables: header start (incl. attached comments) to end of section
	headerLine int  // for explicit tables: start of the [header] line itself
	exprEnd    int  // for explicit tables: past the closing bracket of the header
	insertAt   int  // for explicit tables and the root: where to insert a new key-value line
	// dotted tables are defined by a dotted key inside the section of their
	// nearest section ancestor; they have no header and are extended with
	// dotted keys.
	dotted bool
	items  map[string]*item
}

// item is a named entry in a table: exactly one of leaf, tbl, or arr is set.
type item struct {
	leaf *leaf
	tbl  *table
	arr  []*table // array of tables: one table per [[element]]
}

func newTable(parent *table, name string) *table {
	t := &table{items: map[string]*item{}}
	if parent != nil {
		t.path = childPath(parent, name)
		t.parent = parent
		t.viaArray = parent.viaArray
	}
	return t
}

func childPath(t *table, name string) []string {
	p := make([]string, len(t.path)+1)
	copy(p, t.path)
	p[len(t.path)] = name
	return p
}

// nearestSection returns the closest ancestor-or-self of t that has its own
// section to insert key-values into: an explicit table or the root.
func nearestSection(t *table) *table {
	for !t.explicit && !t.root {
		t = t.parent
	}
	return t
}

// reindex rebuilds the span index from d.data, which must be a valid TOML
// document. Errors are only possible on invalid input and indicate a bug in
// the caller.
func (d *Document) reindex() error {
	root := newTable(nil, "")
	root.root = true
	root.section = span{0, len(d.data)}
	root.insertAt = -1

	p := &unstable.Parser{KeepComments: true}
	p.Reset(d.data)

	current := root
	firstHeaderAttach := -1
	// Current run of contiguous full-line comments. A run that ends on the
	// line right above an expression is attached to that expression: it
	// moves, or is removed, with it.
	runStart, runEnd := -1, -1

	for p.NextExpression() {
		e := p.Expression()

		if e.Kind == unstable.Comment {
			ls := d.lineStart(int(e.Raw.Offset))
			le := d.lineEnd(int(e.Raw.Offset + e.Raw.Length))
			if runEnd != ls {
				runStart = ls
			}
			runEnd = le
			continue
		}

		exprStart, exprEnd, err := d.exprSpan(e)
		if err != nil {
			return err
		}
		ls := d.lineStart(exprStart)
		le := d.lineEnd(exprEnd)
		attach := ls
		if runEnd == ls {
			attach = runStart
		}
		runStart, runEnd = -1, -1

		switch e.Kind {
		case unstable.KeyValue:
			value, err := d.valueSpan(e)
			if err != nil {
				return err
			}
			lf := &leaf{line: span{attach, le}, value: value, exprLine: ls}
			if err := addKeyValue(current, keyParts(e), lf); err != nil {
				return err
			}
			current.insertAt = le
		case unstable.Table:
			current.section.end = attach
			t, err := resolveHeader(root, keyParts(e))
			if err != nil {
				return err
			}
			t.explicit = true
			t.section = span{attach, len(d.data)}
			t.headerLine = ls
			t.exprEnd = exprEnd
			t.insertAt = le
			if firstHeaderAttach < 0 {
				firstHeaderAttach = attach
			}
			current = t
		case unstable.ArrayTable:
			current.section.end = attach
			elem, err := appendArrayElement(root, keyParts(e))
			if err != nil {
				return err
			}
			elem.section = span{attach, len(d.data)}
			elem.headerLine = ls
			elem.exprEnd = exprEnd
			elem.insertAt = le
			if firstHeaderAttach < 0 {
				firstHeaderAttach = attach
			}
			current = elem
		default:
			return fmt.Errorf("toml/edit: internal error: unexpected %s expression", e.Kind)
		}
	}
	if err := p.Error(); err != nil {
		return err
	}
	if current != root {
		current.section.end = len(d.data)
	}
	if root.insertAt < 0 {
		// No top-level key-value: new ones go before the first section
		// (comments attached to its header included), or at the end of a
		// document without sections.
		if firstHeaderAttach >= 0 {
			root.insertAt = firstHeaderAttach
		} else {
			root.insertAt = len(d.data)
		}
	}
	d.root = root
	return nil
}

// exprSpan returns the byte range of a top-level expression, from the first
// byte of its key or opening bracket to past its value or closing bracket,
// excluding surrounding whitespace and trailing comment.
func (d *Document) exprSpan(e *unstable.Node) (int, int, error) {
	if e.Kind == unstable.KeyValue {
		return int(e.Raw.Offset), int(e.Raw.Offset + e.Raw.Length), nil
	}

	// Table and ArrayTable nodes carry no Raw range: recover the header span
	// from the key parts. Only whitespace can separate them from the
	// brackets.
	first, last := firstLastKey(e)
	brackets := 1
	if e.Kind == unstable.ArrayTable {
		brackets = 2
	}
	start := int(first.Raw.Offset)
	for start > 0 && (d.data[start-1] == ' ' || d.data[start-1] == '\t') {
		start--
	}
	for i := 0; i < brackets; i++ {
		if start == 0 || d.data[start-1] != '[' {
			return 0, 0, fmt.Errorf("toml/edit: internal error: cannot find opening bracket of %s header", e.Kind)
		}
		start--
	}
	end := int(last.Raw.Offset + last.Raw.Length)
	for end < len(d.data) && (d.data[end] == ' ' || d.data[end] == '\t') {
		end++
	}
	for i := 0; i < brackets; i++ {
		if end >= len(d.data) || d.data[end] != ']' {
			return 0, 0, fmt.Errorf("toml/edit: internal error: cannot find closing bracket of %s header", e.Kind)
		}
		end++
	}
	return start, end, nil
}

// valueSpan returns the byte range of the value of a KeyValue node, which
// may be a top-level expression or a key-value inside an inline table.
// Composite values (arrays, inline tables) carry no usable Raw range, so it
// is recovered from the last key part instead: the value starts after the
// '=' separator and ends where the key-value does.
func (d *Document) valueSpan(e *unstable.Node) (span, error) {
	_, last := firstLastKey(e)
	i := int(last.Raw.Offset + last.Raw.Length)
	for i < len(d.data) && (d.data[i] == ' ' || d.data[i] == '\t') {
		i++
	}
	if i >= len(d.data) || d.data[i] != '=' {
		return span{}, errors.New("toml/edit: internal error: cannot find '=' of key-value")
	}
	i++
	for i < len(d.data) && (d.data[i] == ' ' || d.data[i] == '\t') {
		i++
	}
	return span{i, int(e.Raw.Offset + e.Raw.Length)}, nil
}

// firstLastKey returns the first and last parts of the key of e. Key-value
// and header expressions produced by the parser always have at least one key
// part.
func firstLastKey(e *unstable.Node) (*unstable.Node, *unstable.Node) {
	var first, last *unstable.Node
	it := e.Key()
	for it.Next() {
		if first == nil {
			first = it.Node()
		}
		last = it.Node()
	}
	return first, last
}

func keyParts(e *unstable.Node) []string {
	var parts []string
	it := e.Key()
	for it.Next() {
		parts = append(parts, string(it.Node().Data))
	}
	return parts
}

// resolveHeader returns the table for a [header] expression, creating
// implicit intermediate tables as needed. Traversing an array of tables
// descends into its last element, per TOML semantics.
func resolveHeader(root *table, parts []string) (*table, error) {
	t := root
	for i, name := range parts {
		it := t.items[name]
		switch {
		case it == nil:
			nt := newTable(t, name)
			t.items[name] = &item{tbl: nt}
			t = nt
		case it.tbl != nil:
			t = it.tbl
		case it.arr != nil && i < len(parts)-1:
			t = it.arr[len(it.arr)-1]
		default:
			return nil, fmt.Errorf("toml/edit: internal error: header %q redefines a value", parts)
		}
	}
	return t, nil
}

// appendArrayElement appends a new element table to the array of tables
// designated by an [[header]] expression, creating it if needed.
func appendArrayElement(root *table, parts []string) (*table, error) {
	t := root
	for _, name := range parts[:len(parts)-1] {
		it := t.items[name]
		switch {
		case it == nil:
			nt := newTable(t, name)
			t.items[name] = &item{tbl: nt}
			t = nt
		case it.tbl != nil:
			t = it.tbl
		case it.arr != nil:
			t = it.arr[len(it.arr)-1]
		default:
			return nil, fmt.Errorf("toml/edit: internal error: array table header %q crosses a value", parts)
		}
	}
	name := parts[len(parts)-1]
	elem := newTable(t, name)
	elem.explicit = true
	elem.viaArray = true
	it := t.items[name]
	switch {
	case it == nil:
		t.items[name] = &item{arr: []*table{elem}}
	case it.arr != nil:
		it.arr = append(it.arr, elem)
	default:
		return nil, fmt.Errorf("toml/edit: internal error: array table header %q redefines a table or value", parts)
	}
	return elem, nil
}

// addKeyValue records a key-value parsed in the section of table sec,
// creating the tables implied by dotted key parts as needed.
func addKeyValue(sec *table, parts []string, lf *leaf) error {
	t := sec
	for _, name := range parts[:len(parts)-1] {
		it := t.items[name]
		switch {
		case it == nil:
			nt := newTable(t, name)
			nt.dotted = true
			t.items[name] = &item{tbl: nt}
			t = nt
		case it.tbl != nil:
			t = it.tbl
		default:
			return fmt.Errorf("toml/edit: internal error: dotted key %q crosses a value", parts)
		}
	}
	name := parts[len(parts)-1]
	if t.items[name] != nil {
		return fmt.Errorf("toml/edit: internal error: duplicate key %q", parts)
	}
	t.items[name] = &item{leaf: lf}
	return nil
}

// subtreeEnd returns the offset right after the last byte of the document
// that belongs to t or any of its descendants. It is where a new section for
// a child of t is inserted.
func subtreeEnd(t *table) int {
	end := t.section.end
	for _, it := range t.items {
		var e int
		switch {
		case it.leaf != nil:
			e = it.leaf.line.end
		case it.tbl != nil:
			e = subtreeEnd(it.tbl)
		default:
			e = arraySubtreeEnd(it.arr)
		}
		if e > end {
			end = e
		}
	}
	return end
}

func arraySubtreeEnd(arr []*table) int {
	end := 0
	for _, elem := range arr {
		if se := subtreeEnd(elem); se > end {
			end = se
		}
	}
	return end
}

// lineStart returns the offset of the first byte of the line off is on.
func (d *Document) lineStart(off int) int {
	return bytes.LastIndexByte(d.data[:off], '\n') + 1
}

// lineEnd returns the offset right past the newline ending the line off is
// on, or the end of the document.
func (d *Document) lineEnd(off int) int {
	i := bytes.IndexByte(d.data[off:], '\n')
	if i < 0 {
		return len(d.data)
	}
	return off + i + 1
}

// parseIndex interprets a key path element as an array index: a non-empty
// string of decimal digits.
func parseIndex(s string) (int, bool) {
	if s == "" || len(s) > 18 {
		return 0, false
	}
	n := 0
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return 0, false
		}
		n = n*10 + int(s[i]-'0')
	}
	return n, true
}
