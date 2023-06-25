/*
Package parser is a Parser for parse INI format content to golang data

There are example data:

	# comments
	name = inhere
	age = 28
	debug = true
	hasQuota1 = 'this is val'
	hasQuota2 = "this is val1"
	shell = ${SHELL}
	noEnv = ${NotExist|defValue}

	; array in def section
	tags[] = a
	tags[] = b
	tags[] = c

	; comments
	[sec1]
	key = val0
	some = value
	stuff = things
	; array in section
	types[] = x
	types[] = y

how to use, please see examples:
*/
package parser

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"reflect"
	"regexp"
	"strings"

	"github.com/gookit/goutil/strutil/textscan"
	"github.com/mitchellh/mapstructure"
)

// match: [section]
var sectionRegex = regexp.MustCompile(`^\[(.*)]$`)

// TokSection for mark a section
const TokSection = textscan.TokComments + 1 + iota

// SectionMatcher match section line: [section]
type SectionMatcher struct{}

// Match section line: [section]
func (m *SectionMatcher) Match(text string, _ textscan.Token) (textscan.Token, error) {
	line := strings.TrimSpace(text)

	if matched := sectionRegex.FindStringSubmatch(line); matched != nil {
		section := strings.TrimSpace(matched[1])
		tok := textscan.NewStringToken(TokSection, section)
		return tok, nil
	}

	return nil, nil
}

// Parser definition
type Parser struct {
	*Options
	// parsed bool
	// comments map, key is name
	comments map[string]string

	// for full parse(allow array, map section)
	fullData map[string]any
	// for simple parse(section only allow map[string]string)
	liteData map[string]map[string]string
}

// New a lite mode Parser with some options
func New(fns ...OptFunc) *Parser {
	return &Parser{Options: NewOptions(fns...)}
}

// NewLite create a lite mode Parser. alias of New()
func NewLite(fns ...OptFunc) *Parser { return New(fns...) }

// NewSimpled create a lite mode Parser
func NewSimpled(fns ...func(*Parser)) *Parser {
	return New().WithOptions(fns...)
}

// NewFulled create a full mode Parser with some options
func NewFulled(fns ...func(*Parser)) *Parser {
	return New(WithParseMode(ModeFull)).WithOptions(fns...)
}

// Parse a INI data string to golang
func Parse(data string, mode parseMode, opts ...func(*Parser)) (p *Parser, err error) {
	p = New(WithParseMode(mode)).WithOptions(opts...)
	err = p.ParseString(data)
	return
}

// Decode INI content to golang data
func Decode(blob []byte, ptr any) error {
	rv := reflect.ValueOf(ptr)
	if rv.Kind() != reflect.Ptr {
		return fmt.Errorf("ini: Decode of non-pointer %s", reflect.TypeOf(ptr))
	}

	p, err := Parse(string(blob), ModeFull, NoDefSection)
	if err != nil {
		return err
	}

	return p.MapStruct(ptr)
}

// NoDefSection set don't return DefSection title
//
// Usage:
//
//	Parser.NoDefSection()
func NoDefSection(p *Parser) { p.NoDefSection = true }

// IgnoreCase set ignore-case
func IgnoreCase(p *Parser) { p.IgnoreCase = true }

// WithOptions apply some options
func (p *Parser) WithOptions(opts ...func(p *Parser)) *Parser {
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// Unmarshal parse ini text and decode to struct
func (p *Parser) Unmarshal(v []byte, ptr any) error {
	if err := p.ParseBytes(v); err != nil {
		return err
	}
	return p.MapStruct(ptr)
}

/*************************************************************
 * do parsing
 *************************************************************/

// ParseString parse from string data
func (p *Parser) ParseString(str string) error {
	if str = strings.TrimSpace(str); str == "" {
		return nil
	}
	return p.ParseReader(strings.NewReader(str))
}

// ParseBytes parse from bytes data
func (p *Parser) ParseBytes(bts []byte) (err error) {
	if len(bts) == 0 {
		return nil
	}
	return p.ParseReader(bytes.NewBuffer(bts))
}

// ParseReader parse from io reader
func (p *Parser) ParseReader(r io.Reader) (err error) {
	_, err = p.ParseFrom(bufio.NewScanner(r))
	return
}

// init parser
func (p *Parser) init() {
	// if p.IgnoreCase {
	// 	p.DefSection = strings.ToLower(p.DefSection)
	// }
	p.comments = make(map[string]string)

	if p.ParseMode == ModeFull {
		p.fullData = make(map[string]any)

		if p.Collector == nil {
			p.Collector = p.collectFullValue
		}
	} else {
		p.liteData = make(map[string]map[string]string)

		if p.Collector == nil {
			p.Collector = p.collectLiteValue
		}
	}
}

// ParseFrom a data scanner
func (p *Parser) ParseFrom(in *bufio.Scanner) (count int64, err error) {
	p.init()
	count = -1

	// create scanner
	ts := textscan.NewScanner(in)
	ts.AddKind(TokSection, "Section")
	ts.AddMatchers(
		&textscan.CommentsMatcher{
			InlineChars: []byte{'#', ';'},
		},
		&SectionMatcher{},
		&textscan.KeyValueMatcher{
			MergeComments: true,
			InlineComment: p.InlineComment,
		},
	)

	section := p.DefSection

	// scan and parsing
	for ts.Scan() {
		tok := ts.Token()

		// comments has been merged to value token
		if !tok.IsValid() || tok.Kind() == textscan.TokComments {
			continue
		}

		if tok.Kind() == TokSection {
			section = tok.Value()

			// collect comments
			if textscan.IsKindToken(textscan.TokComments, ts.PrevToken()) {
				p.comments["_sec_"+section] = ts.PrevToken().Value()
			}
			continue
		}

		// collect value
		if tok.Kind() == textscan.TokValue {
			vt := tok.(*textscan.ValueToken)

			var isSli bool
			key := vt.Key()

			// is array index
			if strings.HasSuffix(key, "[]") {
				// skip parse array on lite mode
				if p.ParseMode == ModeLite {
					continue
				}

				key = key[:len(key)-2]
				isSli = true
			}

			p.collectValue(section, key, vt.Value(), isSli)
			if vt.HasComment() {
				p.comments[section+"_"+key] = vt.Comment()
			}
		}
	}

	count = 0
	err = ts.Err()
	return
}

func (p *Parser) collectValue(section, key, val string, isSlice bool) {
	if p.IgnoreCase {
		key = strings.ToLower(key)
		section = strings.ToLower(section)
	}

	if p.ReplaceNl {
		val = strings.ReplaceAll(val, `\n`, "\n")
	}

	p.Collector(section, key, val, isSlice)
}

func (p *Parser) collectFullValue(section, key, val string, isSlice bool) {
	defSec := p.DefSection
	// p.NoDefSection and current section is default section
	if p.NoDefSection && section == defSec {
		if isSlice {
			curVal, ok := p.fullData[key]
			if ok {
				switch cd := curVal.(type) {
				case []string:
					p.fullData[key] = append(cd, val)
				}
			} else {
				p.fullData[key] = []string{val}
			}
		} else {
			p.fullData[key] = val
		}
		return
	}

	secData, exists := p.fullData[section]
	// first create
	if !exists {
		if isSlice {
			p.fullData[section] = map[string]any{key: []string{val}}
		} else {
			p.fullData[section] = map[string]any{key: val}
		}
		return
	}

	switch sd := secData.(type) {
	case map[string]any: // existed section
		if curVal, ok := sd[key]; ok {
			switch cv := curVal.(type) {
			case string:
				if isSlice {
					sd[key] = []string{cv, val}
				} else {
					sd[key] = val
				}
			case []string:
				sd[key] = append(cv, val)
			default:
				return
			}
		} else {
			if isSlice {
				sd[key] = []string{val}
			} else {
				sd[key] = val
			}
		}
		p.fullData[section] = sd
	case string: // found default section value
		if isSlice {
			p.fullData[section] = map[string]any{key: []string{val}}
		} else {
			p.fullData[section] = map[string]any{key: val}
		}
	}
}

func (p *Parser) collectLiteValue(sec, key, val string, _ bool) {
	if p.IgnoreCase {
		key = strings.ToLower(key)
		sec = strings.ToLower(sec)
	}

	if strMap, ok := p.liteData[sec]; ok {
		strMap[key] = val
		p.liteData[sec] = strMap
	} else {
		// create the section if it does not exist
		p.liteData[sec] = map[string]string{key: val}
	}
}

/*************************************************************
 * export data
 *************************************************************/

// Decode the parsed data to struct ptr
func (p *Parser) Decode(ptr any) error {
	return p.MapStruct(ptr)
}

// MapStruct mapping the parsed data to struct ptr
func (p *Parser) MapStruct(ptr any) (err error) {
	if p.ParseMode == ModeFull {
		if p.NoDefSection {
			return mapStruct(p.TagName, p.fullData, ptr)
		}

		// collect all default section data to top
		anyMap := make(map[string]any, len(p.fullData)+4)
		if defData, ok := p.fullData[p.DefSection]; ok {
			for key, val := range defData.(map[string]any) {
				anyMap[key] = val
			}
		}

		for group, mp := range p.fullData {
			if group == p.DefSection {
				continue
			}
			anyMap[group] = mp
		}
		return mapStruct(p.TagName, anyMap, ptr)
	}

	defData := p.liteData[p.DefSection]
	defLen := len(defData)
	anyMap := make(map[string]any, len(p.liteData)+defLen)

	// collect all default section data to top
	if defLen > 0 {
		for key, val := range defData {
			anyMap[key] = val
		}
	}

	for group, smp := range p.liteData {
		if group == p.DefSection {
			continue
		}
		anyMap[group] = smp
	}

	return mapStruct(p.TagName, anyMap, ptr)
}

func mapStruct(tagName string, data any, ptr any) error {
	mapConf := &mapstructure.DecoderConfig{
		Metadata: nil,
		Result:   ptr,
		TagName:  tagName,
		// will auto convert string to int/uint
		WeaklyTypedInput: true,
	}

	decoder, err := mapstructure.NewDecoder(mapConf)
	if err != nil {
		return err
	}
	return decoder.Decode(data)
}

/*************************************************************
 * helper methods
 *************************************************************/

// Comments get
func (p *Parser) Comments() map[string]string {
	return p.comments
}

// ParsedData get parsed data
func (p *Parser) ParsedData() any {
	if p.ParseMode == ModeFull {
		return p.fullData
	}
	return p.liteData
}

// FullData get parsed data by full parse
func (p *Parser) FullData() map[string]any {
	return p.fullData
}

// LiteData get parsed data by simple parse
func (p *Parser) LiteData() map[string]map[string]string {
	return p.liteData
}

// SimpleData get parsed data by simple parse
func (p *Parser) SimpleData() map[string]map[string]string {
	return p.liteData
}

// LiteSection get parsed data by simple parse
func (p *Parser) LiteSection(name string) map[string]string {
	return p.liteData[name]
}

// Reset parser, clear parsed data
func (p *Parser) Reset() {
	// p.parsed = false
	if p.ParseMode == ModeFull {
		p.fullData = make(map[string]any)
	} else {
		p.liteData = make(map[string]map[string]string)
	}
}
