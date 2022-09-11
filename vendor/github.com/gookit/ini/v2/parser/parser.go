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
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/mitchellh/mapstructure"
)

// errSyntax is returned when there is a syntax error in an INI file.
type errSyntax struct {
	Line   int
	Source string // The contents of the erroneous line, without leading or trailing whitespace
}

// Error message return
func (e errSyntax) Error() string {
	return fmt.Sprintf("invalid INI syntax on line %d: %s", e.Line, e.Source)
}

var (
	// match: [section]
	sectionRegex = regexp.MustCompile(`^\[(.*)]$`)
	// match: foo[] = val
	assignArrRegex = regexp.MustCompile(`^([^=\[\]]+)\[][^=]*=(.*)$`)
	// match: key = val
	assignRegex = regexp.MustCompile(`^([^=]+)=(.*)$`)
	// quote ' "
	quotesRegex = regexp.MustCompile(`^(['"])(.*)(['"])$`)
)

// DefSection default section key name
const DefSection = "__default"

// mode of parse data
// ModeFull  - will parse array
// ModeSimple - don't parse array value
const (
	ModeFull   parseMode = 1
	ModeSimple parseMode = 2
)

type parseMode uint8

// Unit8 mode value to uint8
func (m parseMode) Unit8() uint8 {
	return uint8(m)
}

// UserCollector custom data collector.
// Notice: in simple mode, isSlice always is false.
type UserCollector func(section, key, val string, isSlice bool)

// Parser definition
type Parser struct {
	// for full parse(allow array, map section)
	fullData map[string]interface{}
	// for simple parse(section only allow map[string]string)
	simpleData map[string]map[string]string
	// parsed    bool
	parseMode parseMode

	// ---- options ----

	// TagName of mapping data to struct
	TagName string
	// Ignore case for key name
	IgnoreCase bool
	// default section name. default is "__default"
	DefSection string
	// only for full parse mode
	NoDefSection bool
	// you can custom data collector
	Collector UserCollector
}

// NewFulled create a full mode Parser with some options
func NewFulled(opts ...func(*Parser)) *Parser {
	p := &Parser{
		TagName: 	TagName,
		DefSection: DefSection,
		parseMode:  ModeFull,
		fullData:   make(map[string]interface{}),
	}

	return p.WithOptions(opts...)
}

// NewSimpled create a simple mode Parser
func NewSimpled(opts ...func(*Parser)) *Parser {
	p := &Parser{
		TagName: 	TagName,
		DefSection: DefSection,
		parseMode:  ModeSimple,
		simpleData: make(map[string]map[string]string),
	}

	return p.WithOptions(opts...)
}

// NoDefSection set don't return DefSection title
// Usage:
// 	Parser.NewWithOptions(ini.ParseEnv)
func NoDefSection(p *Parser) {
	p.NoDefSection = true
}

// IgnoreCase set ignore-case
func IgnoreCase(p *Parser) {
	p.IgnoreCase = true
}

// WithOptions apply some options
func (p *Parser) WithOptions(opts ...func(*Parser)) *Parser {
	for _, opt := range opts {
		opt(p)
	}
	return p
}

/*************************************************************
 * do parsing
 *************************************************************/

// Parse a INI data string to golang
func Parse(data string, mode parseMode, opts ...func(*Parser)) (p *Parser, err error) {
	if mode == ModeFull {
		p = NewFulled(opts...)
	} else {
		p = NewSimpled(opts...)
	}

	err = p.ParseString(data)
	return
}

// ParseString parse from string data
func (p *Parser) ParseString(str string) error {
	if str = strings.TrimSpace(str); str == "" {
		return errors.New("cannot input empty string to parse")
	}

	return p.ParseBytes([]byte(str))
}

// ParseBytes parse from bytes data
func (p *Parser) ParseBytes(bts []byte) (err error) {
	buf := &bytes.Buffer{}
	buf.Write(bts)

	scanner := bufio.NewScanner(buf)
	_, err = p.ParseFrom(scanner)
	return
}

// ParseFrom a data scanner
func (p *Parser) ParseFrom(in *bufio.Scanner) (int64, error) {
	return p.parse(in)
}

// fullParse will parse array item
// ref github.com/dombenson/go-ini
func (p *Parser) parse(in *bufio.Scanner) (bytes int64, err error) {
	bytes = -1
	lineNum := 0
	section := p.DefSection

	var readLine bool
	for readLine = in.Scan(); readLine; readLine = in.Scan() {
		line := in.Text()

		bytes++
		bytes += int64(len(line))

		lineNum++
		line = strings.TrimSpace(line)
		if len(line) == 0 { // Skip blank lines
			continue
		}

		if line[0] == ';' || line[0] == '#' { // Skip comments
			continue
		}

		// array/slice data
		if groups := assignArrRegex.FindStringSubmatch(line); groups != nil {
			// skip array parse on simple mode
			if p.parseMode == ModeSimple {
				continue
			}

			// key, val := groups[1], groups[2]
			key, val := strings.TrimSpace(groups[1]), trimWithQuotes(groups[2])

			if p.Collector != nil {
				p.Collector(section, key, val, true)
			} else {
				p.collectFullValue(section, key, val, true)
			}
		} else if groups := assignRegex.FindStringSubmatch(line); groups != nil {
			// key, val := groups[1], groups[2]
			key, val := strings.TrimSpace(groups[1]), trimWithQuotes(groups[2])

			if p.Collector != nil {
				p.Collector(section, key, val, false)
			} else if p.parseMode == ModeFull {
				p.collectFullValue(section, key, val, false)
			} else {
				p.collectMapValue(section, key, val)
			}
		} else if groups := sectionRegex.FindStringSubmatch(line); groups != nil {
			name := strings.TrimSpace(groups[1])
			section = name
		} else {
			err = errSyntax{lineNum, line}
			return
		}
	}

	if bytes < 0 {
		bytes = 0
	}

	err = in.Err()
	return
}

func (p *Parser) collectFullValue(section, key, val string, isSlice bool) {
	defSection := p.DefSection
	if p.IgnoreCase {
		key = strings.ToLower(key)
		section = strings.ToLower(section)
		defSection = strings.ToLower(defSection)
	}

	// p.NoDefSection and current section is default section
	if p.NoDefSection && section == defSection {
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
			p.fullData[section] = map[string]interface{}{key: []string{val}}
		} else {
			p.fullData[section] = map[string]interface{}{key: val}
		}
		return
	}

	switch sd := secData.(type) {
	case map[string]interface{}: // existed section
		curVal, ok := sd[key]
		if ok {
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
			p.fullData[section] = map[string]interface{}{key: []string{val}}
		} else {
			p.fullData[section] = map[string]interface{}{key: val}
		}
	}
}

func (p *Parser) collectMapValue(name string, key, val string) {
	if p.IgnoreCase {
		key = strings.ToLower(key)
		name = strings.ToLower(name)
	}

	if sec, ok := p.simpleData[name]; ok {
		sec[key] = val
		p.simpleData[name] = sec
	} else {
		// create the section if it does not exist
		p.simpleData[name] = map[string]string{key: val}
	}
}

/*************************************************************
 * helper methods
 *************************************************************/

// ParsedData get parsed data
func (p *Parser) ParsedData() interface{} {
	if p.parseMode == ModeFull {
		return p.fullData
	}

	return p.simpleData
}

// ParseMode get current mode
func (p *Parser) ParseMode() uint8 {
	return uint8(p.parseMode)
}

// FullData get parsed data by full parse
func (p *Parser) FullData() map[string]interface{} {
	return p.fullData
}

// SimpleData get parsed data by simple parse
func (p *Parser) SimpleData() map[string]map[string]string {
	return p.simpleData
}

// Reset parser, clear parsed data
func (p *Parser) Reset() {
	// p.parsed = false
	if p.parseMode == ModeFull {
		p.fullData = make(map[string]interface{})
	} else {
		p.simpleData = make(map[string]map[string]string)
	}
}

// MapStruct mapping the parsed data to struct ptr
func (p *Parser) MapStruct(ptr interface{}) (err error) {
	if p.parseMode == ModeFull {
		err = mapStruct(p.TagName, p.fullData, ptr)
	} else {
		err = mapStruct(p.TagName, p.simpleData, ptr)
	}
	return
}

func mapStruct(tagName string, data interface{}, ptr interface{}) error {
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

func trimWithQuotes(inputVal string) (filtered string) {
	filtered = strings.TrimSpace(inputVal)
	groups := quotesRegex.FindStringSubmatch(filtered)

	if len(groups) > 2 && groups[1] == groups[3] {
		filtered = groups[2]
	}
	return
}
