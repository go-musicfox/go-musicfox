/*
Package ini is a ini config file/data manage implement

Source code and other details for the project are available at GitHub:

	https://github.com/gookit/ini

INI parser is: https://github.com/gookit/ini/parser
*/
package ini

import (
	"errors"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"sync"
)

// some default constants
const (
	SepSection = "."
	DefTagName = "ini"
)

var (
	errEmptyKey = errors.New("ini: key name cannot be empty")
	errNotFound = errors.New("ini: key does not exist in the config")
	errReadonly = errors.New("ini: config manager instance in 'readonly' mode")
	// default instance
	dc = New()
)

// Section in INI config
type Section map[string]string

// Ini config data manager
type Ini struct {
	err  error
	opts *Options
	lock sync.RWMutex
	data map[string]Section
	// regex for match user var
	varRegex *regexp.Regexp
}

/*************************************************************
 * config instance
 *************************************************************/

// New a config instance, with default options
func New() *Ini {
	return &Ini{
		data: make(map[string]Section),
		opts: newDefaultOptions(),
	}
}

// NewWithOptions new an instance and with some options
//
// Usage:
//
//	ini.NewWithOptions(ini.ParseEnv, ini.Readonly)
func NewWithOptions(opts ...func(*Options)) *Ini {
	c := New()
	// apply options
	c.WithOptions(opts...)
	return c
}

// Default config instance
func Default() *Ini { return dc }

// ResetStd instance
func ResetStd() { dc = New() }

func (c *Ini) ensureInit() {
	if !c.IsEmpty() {
		return
	}

	if c.data == nil {
		c.data = make(map[string]Section)
	}

	if c.opts == nil {
		c.opts = newDefaultOptions()
	}

	// build var regex. default is `%\(([\w-:]+)\)s`
	if c.opts.ParseVar && c.varRegex == nil {
		// regexStr := `%\([\w-:]+\)s`
		l := regexp.QuoteMeta(c.opts.VarOpen)
		r := regexp.QuoteMeta(c.opts.VarClose)

		// build like: `%\(([\w-:]+)\)s`
		regStr := l + `([\w-` + c.opts.SectionSep + `]+)` + r
		c.varRegex = regexp.MustCompile(regStr)
	}
}

/*************************************************************
 * options func
 *************************************************************/

// GetOptions get options info.
//
// Notice: return is value. so, cannot change Ini instance
func GetOptions() Options {
	return dc.Options()
}

// Options get options info.
//
// Notice: return is value. so, cannot change options
func (c *Ini) Options() Options {
	return *c.opts
}

// WithOptions apply some options
func WithOptions(opts ...func(*Options)) {
	dc.WithOptions(opts...)
}

// WithOptions apply some options
func (c *Ini) WithOptions(opts ...func(*Options)) {
	if !c.IsEmpty() {
		panic("ini: cannot set options after data has been load")
	}

	// apply options
	for _, opt := range opts {
		opt(c.opts)
	}
}

// DefSection get default section name
func DefSection() string {
	return dc.opts.DefSection
}

// DefSection get default section name
func (c *Ini) DefSection() string {
	return c.opts.DefSection
}

/*************************************************************
 * data load
 *************************************************************/

// LoadFiles load data from files
func LoadFiles(files ...string) error { return dc.LoadFiles(files...) }

// LoadFiles load data from files
func (c *Ini) LoadFiles(files ...string) (err error) {
	c.ensureInit()

	for _, file := range files {
		err = c.loadFile(file, false)
		if err != nil {
			return
		}
	}
	return
}

// LoadExists load files, will ignore not exists
func LoadExists(files ...string) error { return dc.LoadExists(files...) }

// LoadExists load files, will ignore not exists
func (c *Ini) LoadExists(files ...string) (err error) {
	c.ensureInit()

	for _, file := range files {
		err = c.loadFile(file, true)
		if err != nil {
			return
		}
	}
	return
}

// LoadStrings load data from strings
func LoadStrings(strings ...string) error { return dc.LoadStrings(strings...) }

// LoadStrings load data from strings
func (c *Ini) LoadStrings(strings ...string) (err error) {
	c.ensureInit()

	for _, str := range strings {
		err = c.parse(str)
		if err != nil {
			return
		}
	}
	return
}

// LoadData load data map
func LoadData(data map[string]Section) error { return dc.LoadData(data) }

// LoadData load data map
func (c *Ini) LoadData(data map[string]Section) (err error) {
	c.ensureInit()

	if len(c.data) == 0 {
		c.data = data
		return
	}

	// append or override setting data
	for name, sec := range data {
		err = c.SetSection(name, sec)
		if err != nil {
			return
		}
	}
	return
}

func (c *Ini) loadFile(file string, loadExist bool) (err error) {
	// open file
	fd, err := os.Open(file)
	if err != nil {
		// skip not exist file
		if os.IsNotExist(err) && loadExist {
			return nil
		}

		return
	}
	//noinspection GoUnhandledErrorResult
	defer fd.Close()

	// read file content
	bts, err := ioutil.ReadAll(fd)
	if err == nil {
		err = c.parse(string(bts))
		if err != nil {
			return
		}
	}
	return
}

/*************************************************************
 * helper methods
 *************************************************************/

// HasKey check key exists
func HasKey(key string) bool { return dc.HasKey(key) }

// HasKey check key exists
func (c *Ini) HasKey(key string) (ok bool) {
	_, ok = c.GetValue(key)
	return
}

// Delete value by key
func Delete(key string) bool { return dc.Delete(key) }

// Delete value by key
func (c *Ini) Delete(key string) (ok bool) {
	if c.opts.Readonly {
		return
	}

	key = c.formatKey(key)
	if key == "" {
		return
	}

	sec, key := c.splitSectionAndKey(key)
	mp, ok := c.data[sec]
	if !ok {
		return
	}

	// key in a section
	if _, ok = mp[key]; ok {
		delete(mp, key)
		c.data[sec] = mp
	}
	return
}

// Reset all data for the default
func Reset() { dc.Reset() }

// Reset all data
func (c *Ini) Reset() {
	c.data = make(map[string]Section)
}

// IsEmpty config data is empty
func IsEmpty() bool { return len(dc.data) == 0 }

// IsEmpty config data is empty
func (c *Ini) IsEmpty() bool {
	return len(c.data) == 0
}

// Data get all data from default instance
func Data() map[string]Section { return dc.data }

// Data get all data
func (c *Ini) Data() map[string]Section {
	return c.data
}

// Error get
func Error() error { return dc.Error() }

// Error get
func (c *Ini) Error() error {
	return c.err
}

/*************************************************************
 * internal helper methods
 *************************************************************/

func (c *Ini) splitSectionAndKey(key string) (string, string) {
	sep := c.opts.SectionSep
	// default find from default Section
	name := c.opts.DefSection

	// get val by path. eg "log.dir"
	if strings.Contains(key, sep) {
		ss := strings.SplitN(key, sep, 2)
		name, key = strings.TrimSpace(ss[0]), strings.TrimSpace(ss[1])
	}

	return name, key
}

// format key by some options
func (c *Ini) formatKey(key string) string {
	sep := c.opts.SectionSep
	key = strings.Trim(strings.TrimSpace(key), sep)

	if c.opts.IgnoreCase {
		key = strings.ToLower(key)
	}

	return key
}

func mapKeyToLower(src map[string]string) map[string]string {
	newMp := make(map[string]string)

	for k, v := range src {
		k = strings.ToLower(k)
		newMp[k] = v
	}
	return newMp
}
