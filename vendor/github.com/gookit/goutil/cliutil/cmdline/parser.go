package cmdline

import (
	"bytes"
	"os/exec"
	"strings"

	"github.com/gookit/goutil/comdef"
	"github.com/gookit/goutil/internal/comfunc"
	"github.com/gookit/goutil/strutil"
)

// LineParser struct
// parse input command line to []string, such as cli os.Args
type LineParser struct {
	parsed bool
	// Line the full input command line text
	// eg `kite top sub -a "this is a message" --foo val1 --bar "val 2"`
	Line string
	// ParseEnv parse ENV var on the line.
	ParseEnv bool
	// the exploded nodes by space.
	nodes []string
	// the parsed args
	args []string

	// temp value
	quoteChar  byte
	quoteIndex int // if > 0, mark is not on start
	tempNode   bytes.Buffer
}

// NewParser create
func NewParser(line string) *LineParser {
	return &LineParser{Line: line}
}

// WithParseEnv with parse ENV var
func (p *LineParser) WithParseEnv() *LineParser {
	p.ParseEnv = true
	return p
}

// AlsoEnvParse input command line text to os.Args, will parse ENV var
func (p *LineParser) AlsoEnvParse() []string {
	p.ParseEnv = true
	return p.Parse()
}

// NewExecCmd quick create exec.Cmd by cmdline string
func (p *LineParser) NewExecCmd() *exec.Cmd {
	// parse get bin and args
	binName, args := p.BinAndArgs()

	// create a new Cmd instance
	return exec.Command(binName, args...)
}

// BinAndArgs get binName and args
func (p *LineParser) BinAndArgs() (bin string, args []string) {
	p.Parse() // ensure parsed.

	ln := len(p.args)
	if ln == 0 {
		return
	}

	bin = p.args[0]
	if ln > 1 {
		args = p.args[1:]
	}
	return
}

// Parse input command line text to os.Args
func (p *LineParser) Parse() []string {
	if p.parsed {
		return p.args
	}

	p.parsed = true
	p.Line = strings.TrimSpace(p.Line)
	if p.Line == "" {
		return p.args
	}

	// enable parse Env var
	if p.ParseEnv {
		p.Line = comfunc.ParseEnvVar(p.Line, nil)
	}

	p.nodes = strings.Split(p.Line, " ")
	if len(p.nodes) == 1 {
		p.args = p.nodes
		return p.args
	}

	for i := 0; i < len(p.nodes); i++ {
		node := p.nodes[i]
		if node == "" {
			continue
		}

		p.parseNode(node)
	}

	p.nodes = p.nodes[:0]
	if p.tempNode.Len() > 0 {
		p.appendTempNode()
	}
	return p.args
}

func (p *LineParser) parseNode(node string) {
	maxIdx := len(node) - 1
	start, end := node[0], node[maxIdx]

	// in quotes
	if p.quoteChar != 0 {
		p.tempNode.WriteByte(' ')

		// end quotes
		if end == p.quoteChar {
			if p.quoteIndex > 0 {
				p.tempNode.WriteString(node) // eg: node="--pretty=format:'one two'"
			} else {
				p.tempNode.WriteString(node[:maxIdx]) // remove last quote
			}
			p.appendTempNode()
		} else { // goon ... write to temp node
			p.tempNode.WriteString(node)
		}
		return
	}

	// quote start
	if start == comdef.DoubleQuote || start == comdef.SingleQuote {
		// only one words. eg: `-m "msg"`
		if end == start {
			p.args = append(p.args, node[1:maxIdx])
			return
		}

		p.quoteChar = start
		p.tempNode.WriteString(node[1:])
	} else if end == comdef.DoubleQuote || end == comdef.SingleQuote {
		p.args = append(p.args, node) // only one node: `msg"`
	} else {
		// eg: --pretty=format:'one two three'
		if strutil.ContainsByte(node, comdef.DoubleQuote) {
			p.quoteIndex = 1 // mark is not on start
			p.quoteChar = comdef.DoubleQuote
		} else if strutil.ContainsByte(node, comdef.SingleQuote) {
			p.quoteIndex = 1
			p.quoteChar = comdef.SingleQuote
		}

		// in quote, append to temp-node
		if p.quoteChar != 0 {
			p.tempNode.WriteString(node)
		} else {
			p.args = append(p.args, node)
		}
	}
}

func (p *LineParser) appendTempNode() {
	p.args = append(p.args, p.tempNode.String())

	// reset context value
	p.quoteChar = 0
	p.quoteIndex = 0
	p.tempNode.Reset()
}
