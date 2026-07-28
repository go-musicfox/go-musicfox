package edit

// Editing inside inline values (arrays and inline tables). The index only
// records document-level structures, so when a path descends into the value
// of a key-value, the value is re-parsed on demand and its sub-spans are
// recovered from the AST: precise splices then rewrite only the element
// being edited.

import (
	"errors"
	"fmt"

	"github.com/pelletier/go-toml/v2/unstable"
)

func rawEnd(n *unstable.Node) int {
	return int(n.Raw.Offset + n.Raw.Length)
}

// leafValueNode re-parses the document and returns the AST node of the value
// of lf. The node references the parser's arena, which stays alive as long
// as the node does.
func (d *Document) leafValueNode(lf *leaf) (*unstable.Node, error) {
	p := &unstable.Parser{}
	p.Reset(d.data)
	for p.NextExpression() {
		e := p.Expression()
		if e.Kind == unstable.KeyValue && rawEnd(e) == lf.value.end {
			return e.Value(), nil
		}
	}
	return nil, errors.New("toml/edit: internal error: cannot locate key-value to edit")
}

// skipInlineTrivia advances past the bytes that may separate the elements of
// an inline value: whitespace, newlines, comments, and (optionally) commas.
func (d *Document) skipInlineTrivia(i int, skipComma bool) int {
	for i < len(d.data) {
		switch c := d.data[i]; {
		case c == ' ' || c == '\t' || c == '\n' || c == '\r':
			i++
		case c == ',' && skipComma:
			i++
		case c == '#':
			i = d.lineEnd(i)
		default:
			return i
		}
	}
	return i
}

// valueNodeEnd returns the offset right past a value node starting at start.
// Composite nodes carry no usable Raw range, so their end is recovered by
// walking their children and scanning for the closing bracket.
func (d *Document) valueNodeEnd(n *unstable.Node, start int) (int, error) {
	switch n.Kind {
	case unstable.InlineTable:
		pos := start + 1
		it := n.Children()
		for it.Next() {
			pos = rawEnd(it.Node())
		}
		pos = d.skipInlineTrivia(pos, true)
		if pos >= len(d.data) || d.data[pos] != '}' {
			return 0, errors.New("toml/edit: internal error: cannot find '}' closing an inline table")
		}
		return pos + 1, nil
	case unstable.Array:
		elems, err := d.arrayElems(n, start)
		if err != nil {
			return 0, err
		}
		pos := start + 1
		if len(elems) > 0 {
			pos = elems[len(elems)-1].sp.end
		}
		pos = d.skipInlineTrivia(pos, true)
		if pos >= len(d.data) || d.data[pos] != ']' {
			return 0, errors.New("toml/edit: internal error: cannot find ']' closing an array")
		}
		return pos + 1, nil
	default:
		return rawEnd(n), nil
	}
}

// elemNode is an element of an array with its byte span.
type elemNode struct {
	node *unstable.Node
	sp   span
}

// arrayElems returns the elements of the array node starting at start ('[')
// with their spans.
func (d *Document) arrayElems(n *unstable.Node, start int) ([]elemNode, error) {
	var elems []elemNode
	pos := start + 1
	it := n.Children()
	for it.Next() {
		c := it.Node()
		cs := d.skipInlineTrivia(pos, true)
		if c.Raw.Length > 0 {
			cs = int(c.Raw.Offset)
		}
		ce, err := d.valueNodeEnd(c, cs)
		if err != nil {
			return nil, err
		}
		elems = append(elems, elemNode{c, span{cs, ce}})
		pos = ce
	}
	return elems, nil
}

// kvNode is a key-value of an inline table with its key parts and byte span.
type kvNode struct {
	kv    *unstable.Node
	parts []string
	sp    span
}

func inlineTableKVs(n *unstable.Node) []kvNode {
	var kvs []kvNode
	it := n.Children()
	for it.Next() {
		kv := it.Node()
		kvs = append(kvs, kvNode{kv, keyParts(kv), span{int(kv.Raw.Offset), rawEnd(kv)}})
	}
	return kvs
}

func kvSpans(kvs []kvNode) []span {
	spans := make([]span, len(kvs))
	for i, e := range kvs {
		spans[i] = e.sp
	}
	return spans
}

func elemSpans(elems []elemNode) []span {
	spans := make([]span, len(elems))
	for i, e := range elems {
		spans[i] = e.sp
	}
	return spans
}

func commonPrefix(a, b []string) int {
	n := 0
	for n < len(a) && n < len(b) && a[n] == b[n] {
		n++
	}
	return n
}

// setInline handles Set when the path continues inside the value of lf.
func (d *Document) setInline(lf *leaf, key []string, from int, value interface{}) error {
	node, err := d.leafValueNode(lf)
	if err != nil {
		return err
	}
	s, err := d.inlineSetSplice(node, lf.value, key, from, value)
	if err != nil {
		return err
	}
	return d.apply(s)
}

func (d *Document) inlineSetSplice(n *unstable.Node, ns span, key []string, from int, value interface{}) (splice, error) {
	switch n.Kind {
	case unstable.InlineTable:
		parts := key[from:]
		kvs := inlineTableKVs(n)
		for _, e := range kvs {
			m := commonPrefix(e.parts, parts)
			switch {
			case m == len(e.parts):
				// The key-value's full key lies on the path.
				vs, err := d.valueSpan(e.kv)
				if err != nil {
					return splice{}, err
				}
				if m == len(parts) {
					raw, err := d.renderValue(value)
					if err != nil {
						return splice{}, err
					}
					return splice{vs, raw}, nil
				}
				return d.inlineSetSplice(e.kv.Value(), vs, key, from+m, value)
			case m == len(parts):
				// The path stops at a table implied by this dotted key.
				return splice{}, fmt.Errorf("toml/edit: cannot set %q: it is a table; delete it first or set its keys individually", pathString(key))
			}
		}
		// Not found: add a new key-value to this inline table.
		text, err := d.renderInlineKV(parts, value)
		if err != nil {
			return splice{}, err
		}
		return d.containerInsert(ns, kvSpans(kvs), text), nil

	case unstable.Array:
		idx, ok := parseIndex(key[from])
		if !ok {
			return splice{}, fmt.Errorf("toml/edit: cannot set %q: %q is an array and %q is not an index", pathString(key), pathString(key[:from]), key[from])
		}
		elems, err := d.arrayElems(n, ns.start)
		if err != nil {
			return splice{}, err
		}
		if idx > len(elems) {
			return splice{}, fmt.Errorf("toml/edit: cannot set %q: index %d out of range: %q has %d elements",
				pathString(key), idx, pathString(key[:from]), len(elems))
		}
		if idx == len(elems) {
			// Append; the remaining key parts wrap the value in tables.
			v := value
			rest := key[from+1:]
			for j := len(rest) - 1; j >= 0; j-- {
				v = map[string]interface{}{rest[j]: v}
			}
			raw, err := d.renderValue(v)
			if err != nil {
				return splice{}, err
			}
			return d.containerInsert(ns, elemSpans(elems), raw), nil
		}
		if from == len(key)-1 {
			raw, err := d.renderValue(value)
			if err != nil {
				return splice{}, err
			}
			return splice{elems[idx].sp, raw}, nil
		}
		return d.inlineSetSplice(elems[idx].node, elems[idx].sp, key, from+1, value)

	default:
		return splice{}, fmt.Errorf("toml/edit: cannot set %q: %q is not a table or array", pathString(key), pathString(key[:from]))
	}
}

// deleteInline computes the splice removing the target of key inside the
// value of lf. A dotted key group inside an inline table is removed one
// key-value at a time: again reports that the caller must re-resolve the
// path and call again to remove the rest of the group.
func (d *Document) deleteInline(lf *leaf, key []string, from int) (splice, bool, bool) {
	node, err := d.leafValueNode(lf)
	if err != nil {
		return splice{}, false, false
	}
	return d.inlineDeleteSplice(node, lf.value, key, from)
}

func (d *Document) inlineDeleteSplice(n *unstable.Node, ns span, key []string, from int) (splice, bool, bool) {
	switch n.Kind {
	case unstable.InlineTable:
		parts := key[from:]
		kvs := inlineTableKVs(n)
		for k, e := range kvs {
			m := commonPrefix(e.parts, parts)
			switch {
			case m == len(e.parts) && m == len(parts):
				return d.containerDelete(ns, kvSpans(kvs), k), false, true
			case m == len(e.parts):
				vs, err := d.valueSpan(e.kv)
				if err != nil {
					return splice{}, false, false
				}
				return d.inlineDeleteSplice(e.kv.Value(), vs, key, from+m)
			case m == len(parts):
				// The path addresses a table implied by this dotted key:
				// remove the whole key-value, then let the caller resolve
				// again for the other key-values of the group.
				return d.containerDelete(ns, kvSpans(kvs), k), true, true
			}
		}
		return splice{}, false, false

	case unstable.Array:
		idx, ok := parseIndex(key[from])
		if !ok {
			return splice{}, false, false
		}
		elems, err := d.arrayElems(n, ns.start)
		if err != nil || idx >= len(elems) {
			return splice{}, false, false
		}
		if from == len(key)-1 {
			return d.containerDelete(ns, elemSpans(elems), idx), false, true
		}
		return d.inlineDeleteSplice(elems[idx].node, elems[idx].sp, key, from+1)

	default:
		return splice{}, false, false
	}
}

// containerInsert builds the splice adding text as a new element of the
// inline container at ns whose current element spans are elems.
func (d *Document) containerInsert(ns span, elems []span, text []byte) splice {
	if len(elems) == 0 {
		return splice{span{ns.start + 1, ns.start + 1}, text}
	}
	e := elems[len(elems)-1].end
	if p := d.skipInlineTrivia(e, false); p < len(d.data) && d.data[p] == ',' {
		// Reuse the trailing comma as the separator.
		return splice{span{p + 1, p + 1}, append([]byte(" "), text...)}
	}
	return splice{span{e, e}, append([]byte(", "), text...)}
}

// containerDelete builds the splice removing element k of the inline
// container at ns, keeping the separating commas balanced.
func (d *Document) containerDelete(ns span, elems []span, k int) splice {
	var s span
	switch {
	case len(elems) == 1:
		s = span{ns.start + 1, elems[0].end}
	case k < len(elems)-1:
		return splice{span: span{elems[k].start, elems[k+1].start}}
	default:
		s = span{elems[k-1].end, elems[k].end}
	}
	// Absorb the trailing comma, if any.
	if p := d.skipInlineTrivia(s.end, false); p < len(d.data) && d.data[p] == ',' {
		s.end = p + 1
	}
	return splice{span: s}
}
