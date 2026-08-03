package edit

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
	"strings"

	toml "github.com/pelletier/go-toml/v2"
	"github.com/pelletier/go-toml/v2/unstable"
)

// splice is a pending replacement of a byte range of the document. An empty
// range is an insertion; an empty repl is a deletion.
type splice struct {
	span
	repl []byte
}

// Set sets the value at the given key path, creating the tables leading to
// it as needed.
//
// If the path designates an existing value, only the bytes of that value are
// replaced: comments and layout around it are preserved. Paths descend into
// arrays and inline tables; elements of arrays (of tables or not) are
// addressed by a decimal index, and the index equal to the current length
// appends a new element. Otherwise a new `key = value` line is added at the
// end of the section of the closest existing parent table, under a new
// [table] header or dotted keys as appropriate. Values are rendered like
// toml.Marshal would, in inline (single-line) form; pass an
// unstable.RawMessage to use an exact TOML representation instead.
//
// Setting a path that designates an existing table or array of tables is an
// error: Set never silently discards parts of the document. Delete first to
// replace a whole table.
func (d *Document) Set(key []string, value interface{}) error {
	if len(key) == 0 {
		return errors.New("toml/edit: key must have at least one element")
	}

	t := d.root
	i := 0
	for i < len(key) {
		it := t.items[key[i]]
		if it == nil {
			break
		}
		last := i == len(key)-1
		switch {
		case it.leaf != nil:
			if last {
				raw, err := d.renderValue(value)
				if err != nil {
					return err
				}
				return d.apply(splice{it.leaf.value, raw})
			}
			return d.setInline(it.leaf, key, i+1, value)
		case it.arr != nil:
			if last {
				return fmt.Errorf("toml/edit: cannot set %q: it is an array of tables; address its elements by index or delete it first", pathString(key))
			}
			idx, ok := parseIndex(key[i+1])
			if !ok {
				return fmt.Errorf("toml/edit: cannot set %q: %q is an array of tables and %q is not an index", pathString(key), pathString(key[:i+1]), key[i+1])
			}
			if idx > len(it.arr) {
				return fmt.Errorf("toml/edit: cannot set %q: index %d out of range: %q has %d elements", pathString(key), idx, pathString(key[:i+1]), len(it.arr))
			}
			if idx == len(it.arr) {
				return d.appendElement(it.arr, key, i+2, value)
			}
			if i+1 == len(key)-1 {
				return fmt.Errorf("toml/edit: cannot set %q: it is a table; delete it first or set its keys individually", pathString(key))
			}
			t = it.arr[idx]
			i += 2
		default:
			if last {
				return fmt.Errorf("toml/edit: cannot set %q: it is a table; delete it first or set its keys individually", pathString(key))
			}
			t = it.tbl
			i++
		}
	}

	// key[i:] does not exist and t is the deepest existing table on the path.
	suffix := key[i:]
	if len(suffix) == 0 {
		return fmt.Errorf("toml/edit: internal error: no keys left to create for %q", pathString(key))
	}

	// Tables defined by dotted keys, and tables inside array-of-tables
	// elements that have no section of their own, are extended with dotted
	// key-values in their nearest section.
	if t.dotted || (t.viaArray && !t.explicit) {
		sec := nearestSection(t)
		rel := make([]string, 0, len(t.path)-len(sec.path)+len(suffix))
		rel = append(rel, t.path[len(sec.path):]...)
		rel = append(rel, suffix...)
		line, err := d.renderDottedKV(rel, value)
		if err != nil {
			return err
		}
		return d.apply(d.insertion(sec.insertAt, line, false))
	}

	if t.explicit || t.root {
		if len(suffix) == 1 {
			line, err := d.renderDottedKV(suffix, value)
			if err != nil {
				return err
			}
			return d.apply(d.insertion(t.insertAt, line, false))
		}
		if t.viaArray {
			// A [header] below an array element would attach to the last
			// element of the array, not necessarily this one: dotted keys
			// stay unambiguous.
			line, err := d.renderDottedKV(suffix, value)
			if err != nil {
				return err
			}
			return d.apply(d.insertion(t.insertAt, line, false))
		}
	}

	// The remaining tables need a section of their own, placed right after
	// everything that belongs to t. A single header with the full dotted
	// path creates all of them. t may be an implicit table (only defined
	// through the headers of its sub-tables): declaring its header afterwards
	// is valid TOML.
	headerPath := make([]string, 0, len(t.path)+len(suffix)-1)
	headerPath = append(headerPath, t.path...)
	headerPath = append(headerPath, suffix[:len(suffix)-1]...)
	headerKey, err := d.renderKeyPath(headerPath)
	if err != nil {
		return err
	}
	kv, err := d.renderDottedKV(suffix[len(suffix)-1:], value)
	if err != nil {
		return err
	}
	text := append([]byte("["+headerKey+"]"), d.eol()...)
	text = append(text, kv...)
	return d.apply(d.insertion(subtreeEnd(t), text, true))
}

// appendElement appends a new [[element]] to an existing array of tables,
// initialized with the key-value described by key[restFrom:] and value.
func (d *Document) appendElement(arr []*table, key []string, restFrom int, value interface{}) error {
	rest := key[restFrom:]
	if len(rest) == 0 {
		return fmt.Errorf("toml/edit: cannot set %q: a new array element is a table; set its keys individually", pathString(key))
	}
	headerKey, err := d.renderKeyPath(arr[0].path)
	if err != nil {
		return err
	}
	kv, err := d.renderDottedKV(rest, value)
	if err != nil {
		return err
	}
	text := append([]byte("[["+headerKey+"]]"), d.eol()...)
	text = append(text, kv...)
	return d.apply(d.insertion(arraySubtreeEnd(arr), text, true))
}

// Delete removes the value, key-value, table, or array of tables at the
// given key path, along with the comments attached to it. It returns true if
// something was deleted. Elements of arrays are addressed by a decimal
// index. Deleting a table removes all its content, including sections
// defined elsewhere in the document.
func (d *Document) Delete(key []string) bool {
	// Document-level targets are removed in one batch. A target inside an
	// inline value may be expressed by several key-values of a dotted group:
	// those are removed one at a time, re-resolving the path in between.
	deleted := false
	for {
		splices, again, found := d.resolveDelete(key)
		if !found || d.apply(splices...) != nil {
			return deleted
		}
		deleted = true
		if !again {
			return true
		}
	}
}

// resolveDelete computes the splices removing the target of key. again
// reports that the deletion may be partial (one key-value of a dotted group)
// and that the caller should resolve the path again after applying.
func (d *Document) resolveDelete(key []string) ([]splice, bool, bool) {
	if len(key) == 0 {
		return nil, false, false
	}
	t := d.root
	i := 0
	for i < len(key) {
		it := t.items[key[i]]
		if it == nil {
			return nil, false, false
		}
		last := i == len(key)-1
		switch {
		case it.leaf != nil:
			if last {
				return d.deletionSplices([]span{it.leaf.line}), false, true
			}
			s, again, ok := d.deleteInline(it.leaf, key, i+1)
			if !ok {
				return nil, false, false
			}
			return []splice{s}, again, true
		case it.arr != nil:
			if last {
				return d.deletionSplices(collectSpans(it, nil)), false, true
			}
			idx, ok := parseIndex(key[i+1])
			if !ok || idx >= len(it.arr) {
				return nil, false, false
			}
			if i+1 == len(key)-1 {
				return d.deletionSplices(collectTableSpans(it.arr[idx], nil)), false, true
			}
			t = it.arr[idx]
			i += 2
		default:
			if last {
				return d.deletionSplices(collectSpans(it, nil)), false, true
			}
			t = it.tbl
			i++
		}
	}
	return nil, false, false
}

func (d *Document) deletionSplices(spans []span) []splice {
	spans = mergeSpans(spans)
	splices := make([]splice, len(spans))
	for i, s := range spans {
		splices[i] = splice{span: d.absorbTrailingBlanks(s)}
	}
	return splices
}

// collectSpans accumulates the byte ranges expressing an item: key-value
// lines and whole sections. Ranges may overlap (a section contains the lines
// of its key-values): mergeSpans resolves that.
func collectSpans(it *item, spans []span) []span {
	switch {
	case it.leaf != nil:
		spans = append(spans, it.leaf.line)
	case it.tbl != nil:
		spans = collectTableSpans(it.tbl, spans)
	default:
		for _, elem := range it.arr {
			spans = collectTableSpans(elem, spans)
		}
	}
	return spans
}

func collectTableSpans(t *table, spans []span) []span {
	if t.explicit {
		spans = append(spans, t.section)
	}
	for _, child := range t.items {
		spans = collectSpans(child, spans)
	}
	return spans
}

func mergeSpans(spans []span) []span {
	sort.Slice(spans, func(i, j int) bool { return spans[i].start < spans[j].start })
	out := spans[:0]
	for _, s := range spans {
		if n := len(out); n > 0 && s.start <= out[n-1].end {
			if s.end > out[n-1].end {
				out[n-1].end = s.end
			}
			continue
		}
		out = append(out, s)
	}
	return out
}

// apply replaces the given byte ranges, then validates and re-indexes the
// document. If the result is not a valid TOML document, no change is made.
func (d *Document) apply(splices ...splice) error {
	sort.Slice(splices, func(i, j int) bool { return splices[i].start < splices[j].start })
	var buf bytes.Buffer
	prev := 0
	for _, s := range splices {
		if s.start < prev {
			return errors.New("toml/edit: internal error: overlapping edits")
		}
		buf.Write(d.data[prev:s.start])
		buf.Write(s.repl)
		prev = s.end
	}
	buf.Write(d.data[prev:])
	data := buf.Bytes()

	var v interface{}
	if err := toml.Unmarshal(data, &v); err != nil {
		return fmt.Errorf("toml/edit: edit would produce an invalid TOML document: %w", err)
	}

	oldData, oldRoot := d.data, d.root
	d.data = data
	if err := d.reindex(); err != nil {
		d.data, d.root = oldData, oldRoot
		return fmt.Errorf("toml/edit: edit would produce an invalid TOML document: %w", err)
	}
	return nil
}

// insertion builds the splice inserting text (one or more full lines, ending
// with a newline) at pos, which is always a line boundary or the end of the
// document. Sections get a blank line separating them from the surrounding
// expressions.
func (d *Document) insertion(pos int, text []byte, section bool) splice {
	eol := d.eol()
	var b []byte
	freshLine := pos == 0 || d.data[pos-1] == '\n'
	if !freshLine {
		// Only possible at the end of a document with no final newline.
		b = append(b, eol...)
	}
	if section && pos > 0 && (!freshLine || !d.blankLineBefore(pos)) {
		b = append(b, eol...)
	}
	b = append(b, text...)
	if section && pos < len(d.data) {
		b = append(b, eol...)
	}
	return splice{span{pos, pos}, b}
}

// absorbTrailingBlanks extends a deletion reaching the end of the document
// over the blank lines that separated it from the content above, so that the
// document does not end with stray blank lines.
func (d *Document) absorbTrailingBlanks(s span) span {
	if s.end != len(d.data) {
		return s
	}
	for s.start > 0 && d.data[s.start-1] == '\n' && d.blankLineBefore(s.start) {
		i := s.start - 1
		if i > 0 && d.data[i-1] == '\r' {
			i--
		}
		s.start = i
	}
	return s
}

// blankLineBefore reports whether the line ending at pos (excluded) is
// empty. Only meaningful when pos is a line boundary.
func (d *Document) blankLineBefore(pos int) bool {
	i := pos - 1 // d.data[i] == '\n'
	if i > 0 && d.data[i-1] == '\r' {
		i--
	}
	return i == 0 || d.data[i-1] == '\n'
}

// eol returns the line ending to use for inserted lines: CRLF if the
// document already uses it, LF otherwise.
func (d *Document) eol() []byte {
	if bytes.Contains(d.data, []byte("\r\n")) {
		return []byte("\r\n")
	}
	return []byte("\n")
}

// renderKVLine renders a complete `key = value` line (terminated by '\n')
// with the toml encoder, so that key quoting and value formatting follow
// Marshal conventions exactly. Values render in inline form; a value that
// cannot render to a single line is an error.
func renderKVLine(name string, value interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := toml.NewEncoder(&buf)
	enc.SetTablesInline(true)
	if err := enc.Encode(map[string]interface{}{name: value}); err != nil {
		return nil, fmt.Errorf("toml/edit: cannot render value: %w", err)
	}
	line := buf.Bytes()
	// A value the encoder omits instead of erroring on (e.g. a nil map
	// value) renders to nothing.
	if len(line) == 0 || line[len(line)-1] != '\n' || bytes.IndexByte(line, '\n') != len(line)-1 {
		return nil, fmt.Errorf("toml/edit: cannot render value of type %T as a single key-value line", value)
	}
	return line, nil
}

// renderValue renders the bare TOML representation of a value. An
// unstable.RawMessage is used as-is.
func (d *Document) renderValue(value interface{}) ([]byte, error) {
	if raw, ok := value.(unstable.RawMessage); ok {
		raw = bytes.TrimSpace(raw)
		if len(raw) == 0 {
			return nil, errors.New("toml/edit: raw value is empty")
		}
		return bytes.Clone(raw), nil
	}
	line, err := renderKVLine("v", value)
	if err != nil {
		return nil, err
	}
	prefix := []byte("v = ")
	if !bytes.HasPrefix(line, prefix) {
		return nil, errors.New("toml/edit: internal error: unexpected encoder output for value")
	}
	return line[len(prefix) : len(line)-1], nil
}

// renderInlineKV renders a `dotted.key = value` fragment without a line
// ending, as used inside inline tables.
func (d *Document) renderInlineKV(parts []string, value interface{}) ([]byte, error) {
	kp, err := d.renderKeyPath(parts)
	if err != nil {
		return nil, err
	}
	v, err := d.renderValue(value)
	if err != nil {
		return nil, err
	}
	b := make([]byte, 0, len(kp)+len(v)+3)
	b = append(b, kp...)
	b = append(b, " = "...)
	b = append(b, v...)
	return b, nil
}

// renderDottedKV renders a `dotted.key = value` line ending with the
// document's line ending.
func (d *Document) renderDottedKV(parts []string, value interface{}) ([]byte, error) {
	b, err := d.renderInlineKV(parts, value)
	if err != nil {
		return nil, err
	}
	return append(b, d.eol()...), nil
}

// renderKeyPath renders a dotted key path, quoting the parts that need it
// the same way the toml encoder does.
func (d *Document) renderKeyPath(parts []string) (string, error) {
	rendered := make([]string, len(parts))
	for i, part := range parts {
		line, err := renderKVLine(part, 0)
		if err != nil {
			return "", err
		}
		suffix := []byte(" = 0\n")
		if !bytes.HasSuffix(line, suffix) {
			return "", errors.New("toml/edit: internal error: unexpected encoder output for key")
		}
		rendered[i] = string(line[:len(line)-len(suffix)])
	}
	return strings.Join(rendered, "."), nil
}

func pathString(key []string) string {
	return strings.Join(key, ".")
}
