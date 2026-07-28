// Package edit provides comment- and layout-preserving editing of TOML
// documents.
//
// A Document is created from TOML source with Parse. Bytes returns the
// current source, byte-for-byte identical to the input except for the parts
// modified through Set, Delete, SetComment, and SetTrailingComment: an edit
// rewrites only the bytes that express it, leaving the comments, whitespace,
// and ordering of everything else untouched.
//
// # Key paths
//
// Values are addressed by their key path, one element per key part:
// []string{"servers", "alpha", "ip"} addresses ip in [servers.alpha]. Path
// elements are plain strings, never quoted or dotted: quoting is applied as
// needed when writing.
//
// Paths descend through tables (whether defined by [table] headers, dotted
// keys, or inline), through arrays of tables, and through arrays. When a
// path element steps into an array, it is interpreted as a 0-based decimal
// index; everywhere else it is a key, so keys that look like numbers are
// unambiguous. For Set, an index equal to the array's length appends a new
// element, including a new [[table]] section for arrays of tables.
//
// # Values
//
// Values passed to Set are rendered with the same encoder as toml.Marshal,
// in inline (single-line) form. To control the exact TOML representation
// (formatting, multi-line strings, ...), pass an unstable.RawMessage: its
// bytes are used verbatim as the value. New tables created by Set get their
// own [header] section, appended after the section of their closest existing
// parent; dotted keys are used instead when the parent was defined with
// dotted keys or lives inside an array-of-tables element (where a new header
// would attach to the wrong element).
//
// Setting a path that designates an existing table or array of tables is an
// error: Set never silently discards parts of the document. Delete it first
// to replace it wholesale.
//
// # Comments
//
// Comments travel with the expression they annotate: the contiguous
// full-line comments directly above a key-value or table header, and the
// comment trailing it on the same line, move and are deleted with it. They
// can be read and written with Comment, SetComment, TrailingComment, and
// SetTrailingComment.
//
// # Guarantees
//
// Every mutation is validated: if an edit would produce an invalid TOML
// document, the document is left unchanged and an error is returned. A few
// exotic combinations (for example extending, from outside, a table defined
// implicitly inside an array-of-tables element) are rejected that way when
// TOML offers no valid syntax for them.
//
// Like the rest of the unstable API, this package does not follow the
// backward compatibility guarantees of go-toml. It also favors correctness
// and fidelity over speed: the document is re-validated and re-indexed after
// every mutation, so it is meant for editing configuration files, not for
// hot paths.
//
// A Document is not safe for concurrent use.
package edit

import (
	"bytes"

	toml "github.com/pelletier/go-toml/v2"
)

// Document is a TOML document whose layout (comments, whitespace, order) is
// preserved across edits.
type Document struct {
	data []byte
	root *table
}

// Parse reads a TOML document and returns a Document ready to be inspected
// and edited. The input must be a valid TOML document: syntactic or semantic
// errors (for example duplicate keys) are returned, as editing an invalid
// document would not produce meaningful results. The input slice is copied
// and can be reused by the caller.
func Parse(b []byte) (*Document, error) {
	var v interface{}
	if err := toml.Unmarshal(b, &v); err != nil {
		return nil, err
	}
	d := &Document{data: bytes.Clone(b)}
	if err := d.reindex(); err != nil {
		return nil, err
	}
	return d, nil
}

// Bytes returns the current TOML source of the document. If the document has
// not been modified, it is identical to the input of Parse.
func (d *Document) Bytes() []byte {
	return bytes.Clone(d.data)
}

// String returns the current TOML source of the document, like Bytes.
func (d *Document) String() string {
	return string(d.data)
}

// Unmarshal decodes the current state of the document into v, like
// toml.Unmarshal.
func (d *Document) Unmarshal(v interface{}) error {
	return toml.Unmarshal(d.data, v)
}

// Get returns the decoded value at the given key path, and whether it
// exists. Values are decoded like toml.Unmarshal into an interface{}: tables
// (inline or not) become map[string]interface{}, arrays and arrays of tables
// become []interface{}, and scalars follow the usual decoding rules. Array
// elements are addressed by a decimal index. An empty path returns the whole
// document.
func (d *Document) Get(key []string) (interface{}, bool) {
	var v interface{}
	if err := toml.Unmarshal(d.data, &v); err != nil {
		// The document is valid by construction.
		return nil, false
	}
	for _, k := range key {
		switch c := v.(type) {
		case map[string]interface{}:
			var ok bool
			v, ok = c[k]
			if !ok {
				return nil, false
			}
		case []interface{}:
			idx, ok := parseIndex(k)
			if !ok || idx >= len(c) {
				return nil, false
			}
			v = c[idx]
		default:
			return nil, false
		}
	}
	return v, true
}

// Has reports whether a value exists at the given key path.
func (d *Document) Has(key []string) bool {
	_, ok := d.Get(key)
	return ok
}
