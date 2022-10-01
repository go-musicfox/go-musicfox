// Copyright 2017 Meng Zhuo.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package cookiestxt implement parser of cookies txt format that commonly supported by
// curl / wget / aria2c / chrome / firefox
//
// see http://www.cookiecentral.com/faq/#3.5 for more detail
package cookiestxt

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	// http://www.cookiecentral.com/faq/#3.5
	// The domain that created AND that can read the variable.
	domainIdx = iota
	// A TRUE/FALSE value indicating if all machines within a given domain can access the variable. This value is set automatically by the browser, depending on the value you set for domain.
	flagIdx
	// The path within the domain that the variable is valid for.
	pathIdx
	// A TRUE/FALSE value indicating if a secure connection with the domain is needed to access the variable.
	secureIdx
	// The UNIX time that the variable will expire on. UNIX time is defined as the number of seconds since Jan 1, 1970 00:00:00 GMT.
	expirationIdx
	// The name of the variable.
	nameIdx
	// The value of the variable.
	valueIdx
)

const (
	httpOnlyPrefix = "#HttpOnly_"
	fieldsCount    = 6
)

// Parse cookie txt file format from input stream
func Parse(rd io.Reader) (cl []*http.Cookie, err error) {
	scanner := bufio.NewScanner(rd)
	var line int
	for scanner.Scan() {
		line++

		trimed := strings.TrimSpace(scanner.Text())
		if len(trimed) < fieldsCount {
			continue
		}

		if trimed[0] == '#' && !strings.HasPrefix(trimed, httpOnlyPrefix) {
			// comment
			continue
		}

		var c *http.Cookie
		c, err = ParseLine(scanner.Text())
		if err != nil {
			fmt.Errorf("cookiestxt line:%d, err:%s", line, err)
			continue
		}
		cl = append(cl, c)
		line++
	}

	err = scanner.Err()
	return
}

// ParseLine parse single cookie from one line
func ParseLine(raw string) (c *http.Cookie, err error) {
	f := strings.Fields(raw)
	fl := len(f)
	if fl < fieldsCount {
		err = fmt.Errorf("expecting fields=6, got=%d", fl)
		return
	}
	value := ""
	if fl > 6 {
		value = f[valueIdx]
	}
	c = &http.Cookie{
		Raw:    raw,
		Name:   f[nameIdx],
		Value:  value,
		Path:   f[pathIdx],
		MaxAge: 0,
		Secure: parseBool(f[secureIdx]),
	}

	var ts int64
	ts, err = strconv.ParseInt(f[expirationIdx], 10, 64)
	if err != nil {
		return
	}
	c.Expires = time.Unix(ts, 0)

	c.Domain = f[domainIdx]
	if strings.HasPrefix(c.Domain, httpOnlyPrefix) {
		c.HttpOnly = true
		c.Domain = c.Domain[len(httpOnlyPrefix):]
	}

	return
}

func parseBool(input string) bool {
	return input == "TRUE"
}
