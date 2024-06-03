/* Copyright（2） 2018 by  asmcos .
Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package requests

import (
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"
)

var VERSION string = "0.8"

type Request struct {
	httpreq *http.Request
	Header  *http.Header
	Client  *http.Client
	Debug   int
	Cookies []*http.Cookie
}

type Response struct {
	R       *http.Response
	content []byte
	text    string
	req     *Request
}

type Header map[string]string
type Params map[string]string
type Datas map[string]string // for post form
type Files map[string]string // name ,filename
type DryRun bool

// {username,password}
type Auth []string

func Requests() *Request {

	req := new(Request)

	req.httpreq = &http.Request{
		Method:     "GET",
		Header:     make(http.Header),
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
	}
	req.Header = &req.httpreq.Header
	req.httpreq.Header.Set("User-Agent", "Go-Requests "+VERSION)

	req.Client = &http.Client{}

	// auto with Cookies
	// cookiejar.New source code return jar, nil
	jar, _ := cookiejar.New(nil)

	req.Client.Jar = jar

	return req
}

// Get ,req.Get

func Get(origurl string, args ...interface{}) (resp *Response, err error) {
	req := Requests()

	// call request Get
	resp, err = req.Get(origurl, args...)
	return resp, err
}

func (req *Request) Get(origurl string, args ...interface{}) (resp *Response, err error) {

	req.httpreq.Method = "GET"

	// set params ?a=b&b=c
	//set Header
	params := []map[string]string{}

	//reset Cookies,
	//Client.Do can copy cookie from client.Jar to req.Header
	delete(req.httpreq.Header, "Cookie")

	var dryRun DryRun
	for _, arg := range args {
		switch a := arg.(type) {
		// arg is Header , set to request header
		case Header:

			for k, v := range a {
				req.Header.Set(k, v)
			}
			// arg is "GET" params
			// ?title=website&id=1860&from=login
		case Params:
			params = append(params, a)
		case Auth:
			// a{username,password}
			req.httpreq.SetBasicAuth(a[0], a[1])
		case DryRun:
			dryRun = a
		}
	}

	disturl, _ := buildURLParams(origurl, params...)

	//prepare to Do
	URL, err := url.Parse(disturl)
	if err != nil {
		return nil, err
	}
	req.httpreq.URL = URL

	req.ClientSetCookies()

	req.RequestDebug()

	if dryRun {
		return nil, nil
	}
	res, err := req.Client.Do(req.httpreq)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	resp = &Response{}
	resp.R = res
	resp.req = req

	resp.Content()
	defer res.Body.Close()

	resp.ResponseDebug()
	return resp, nil
}

// handle URL params
func buildURLParams(userURL string, params ...map[string]string) (string, error) {
	parsedURL, err := url.Parse(userURL)

	if err != nil {
		return "", err
	}

	parsedQuery, err := url.ParseQuery(parsedURL.RawQuery)

	if err != nil {
		return "", nil
	}

	for _, param := range params {
		for key, value := range param {
			parsedQuery.Add(key, value)
		}
	}
	return addQueryParams(parsedURL, parsedQuery), nil
}

func addQueryParams(parsedURL *url.URL, parsedQuery url.Values) string {
	if len(parsedQuery) > 0 {
		return strings.Join([]string{strings.Replace(parsedURL.String(), "?"+parsedURL.RawQuery, "", -1), parsedQuery.Encode()}, "?")
	}
	return strings.Replace(parsedURL.String(), "?"+parsedURL.RawQuery, "", -1)
}

func (req *Request) RequestDebug() {

	if req.Debug != 1 {
		return
	}

	log.Println("===========Go RequestDebug ============")

	message, err := httputil.DumpRequestOut(req.httpreq, false)
	if err != nil {
		return
	}
	log.Println(string(message))

	if len(req.Client.Jar.Cookies(req.httpreq.URL)) > 0 {
		log.Println("Cookies:")
		for _, cookie := range req.Client.Jar.Cookies(req.httpreq.URL) {
			log.Println(cookie)
		}
	}
}

// cookies
// cookies only save to Client.Jar
// req.Cookies is temporary
func (req *Request) SetCookie(cookie *http.Cookie) {
	req.Cookies = append(req.Cookies, cookie)
}

func (req *Request) ClearCookies() {
	req.Cookies = req.Cookies[0:0]
}

func (req *Request) ClientSetCookies() {

	if len(req.Cookies) > 0 {
		// 1. Cookies have content, Copy Cookies to Client.jar
		// 2. Clear  Cookies
		req.Client.Jar.SetCookies(req.httpreq.URL, req.Cookies)
		req.ClearCookies()
	}

}

// set timeout s = second
func (req *Request) SetTimeout(n time.Duration) {
	req.Client.Timeout = time.Duration(n * time.Second)
}

func (req *Request) Close() {
	req.httpreq.Close = true
}

func (req *Request) Proxy(proxyurl string) {

	urli := url.URL{}
	urlproxy, err := urli.Parse(proxyurl)
	if err != nil {
		log.Println("Set proxy failed")
		return
	}
	req.Client.Transport = &http.Transport{
		Proxy:           http.ProxyURL(urlproxy),
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

}

func (req *Request) HttpRequest() *http.Request {
	return req.httpreq
}

/**************/
func (resp *Response) ResponseDebug() {

	if resp.req.Debug != 1 {
		return
	}

	log.Println("===========Go ResponseDebug ============")

	message, err := httputil.DumpResponse(resp.R, false)
	if err != nil {
		return
	}

	log.Println(string(message))

}

func (resp *Response) Content() []byte {

	var err error

	if len(resp.content) > 0 {
		return resp.content
	}

	var Body = resp.R.Body
	if resp.R.Header.Get("Content-Encoding") == "gzip" && resp.req.Header.Get("Accept-Encoding") != "" {
		// log.Println("gzip")
		reader, err := gzip.NewReader(Body)
		if err != nil {
			return nil
		}
		Body = reader
	}

	defer Body.Close()
	data, err := ioutil.ReadAll(Body)
	if err != nil {
		log.Println(err)
		return nil
	}

	resp.R.Body = ioutil.NopCloser(bytes.NewBuffer(data))
	resp.content = data

	return resp.content
}

func (resp *Response) ReloadContent() []byte {
	content := resp.content
	resp.content = nil
	b := resp.Content()
	if len(b) == 0 {
		// reload content failed
		resp.content = content
	}
	return resp.content
}

func (resp *Response) Text() string {
	if resp.content == nil {
		resp.Content()
	}
	resp.text = string(resp.content)
	return resp.text
}

func (resp *Response) SaveFile(filename string) error {
	if resp.content == nil {
		resp.Content()
	}
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(resp.content)
	f.Sync()

	return err
}

func (resp *Response) Json(v interface{}) error {
	if resp.content == nil {
		resp.Content()
	}
	return json.Unmarshal(resp.content, v)
}

func (resp *Response) Cookies() (cookies []*http.Cookie) {
	httpreq := resp.req.httpreq
	client := resp.req.Client

	cookies = client.Jar.Cookies(httpreq.URL)

	return cookies

}

func (resp *Response) SetRequest(req *Request) {
	resp.req = req
}

/**************post*************************/
// call req.Post ,only for easy
func Post(origurl string, args ...interface{}) (resp *Response, err error) {
	req := Requests()

	// call request Get
	resp, err = req.Post(origurl, args...)
	return resp, err
}

func PostJson(origurl string, args ...interface{}) (resp *Response, err error) {
	req := Requests()

	// call request Get
	resp, err = req.PostJson(origurl, args...)
	return resp, err
}

// POST requests

func (req *Request) PostJson(origurl string, args ...interface{}) (resp *Response, err error) {

	req.httpreq.Method = "POST"

	req.Header.Set("Content-Type", "application/json")

	//reset Cookies,
	//Client.Do can copy cookie from client.Jar to req.Header
	delete(req.httpreq.Header, "Cookie")

	var dryRun DryRun
	for _, arg := range args {
		switch a := arg.(type) {
		// arg is Header , set to request header
		case Header:

			for k, v := range a {
				req.Header.Set(k, v)
			}
		case string:
			req.setBodyRawBytes(ioutil.NopCloser(strings.NewReader(arg.(string))))
		case Auth:
			// a{username,password}
			req.httpreq.SetBasicAuth(a[0], a[1])
		case DryRun:
			dryRun = a
		default:
			b := new(bytes.Buffer)
			err = json.NewEncoder(b).Encode(a)
			if err != nil {
				return nil, err
			}
			req.setBodyRawBytes(ioutil.NopCloser(b))
		}
	}

	//prepare to Do
	URL, err := url.Parse(origurl)
	if err != nil {
		return nil, err
	}
	req.httpreq.URL = URL

	req.ClientSetCookies()

	req.RequestDebug()

	if dryRun {
		return nil, nil
	}

	res, err := req.Client.Do(req.httpreq)

	// clear post  request information
	req.httpreq.Body = nil
	req.httpreq.GetBody = nil
	req.httpreq.ContentLength = 0

	if err != nil {
		log.Println(err)
		return nil, err
	}

	resp = &Response{}
	resp.R = res
	resp.req = req

	resp.Content()
	defer res.Body.Close()
	resp.ResponseDebug()
	return resp, nil
}

func (req *Request) Post(origurl string, args ...interface{}) (resp *Response, err error) {

	req.httpreq.Method = "POST"

	//set default
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// set params ?a=b&b=c
	//set Header
	params := []map[string]string{}
	datas := []map[string]string{} // POST
	files := []map[string]string{} //post file

	//reset Cookies,
	//Client.Do can copy cookie from client.Jar to req.Header
	delete(req.httpreq.Header, "Cookie")

	var dryRun DryRun
	for _, arg := range args {
		switch a := arg.(type) {
		// arg is Header , set to request header
		case Header:

			for k, v := range a {
				req.Header.Set(k, v)
			}
			// arg is "GET" params
			// ?title=website&id=1860&from=login
		case Params:
			params = append(params, a)

		case Datas: //Post form data,packaged in body.
			datas = append(datas, a)
		case Files:
			files = append(files, a)
		case Auth:
			// a{username,password}
			req.httpreq.SetBasicAuth(a[0], a[1])
		case DryRun:
			dryRun = a
		}
	}

	disturl, _ := buildURLParams(origurl, params...)

	if len(files) > 0 {
		req.buildFilesAndForms(files, datas)

	} else {
		Forms := req.buildForms(datas...)
		req.setBodyBytes(Forms) // set forms to body
	}
	//prepare to Do
	URL, err := url.Parse(disturl)
	if err != nil {
		return nil, err
	}
	req.httpreq.URL = URL

	req.ClientSetCookies()

	req.RequestDebug()

	if dryRun {
		return nil, nil
	}
	res, err := req.Client.Do(req.httpreq)

	// clear post param
	req.httpreq.Body = nil
	req.httpreq.GetBody = nil
	req.httpreq.ContentLength = 0

	if err != nil {
		log.Println(err)
		return nil, err
	}

	resp = &Response{}
	resp.R = res
	resp.req = req

	resp.Content()
	defer res.Body.Close()

	resp.ResponseDebug()
	return resp, nil
}

// only set forms
func (req *Request) setBodyBytes(Forms url.Values) {

	// maybe
	data := Forms.Encode()
	req.httpreq.Body = ioutil.NopCloser(strings.NewReader(data))
	req.httpreq.ContentLength = int64(len(data))
}

// only set forms
func (req *Request) setBodyRawBytes(read io.ReadCloser) {
	req.httpreq.Body = read
}

// upload file and form
// build to body format
func (req *Request) buildFilesAndForms(files []map[string]string, datas []map[string]string) {

	//handle file multipart

	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	for _, file := range files {
		for k, v := range file {
			part, err := w.CreateFormFile(k, v)
			if err != nil {
				log.Printf("Upload %s failed!", v)
				panic(err)
			}
			file := openFile(v)
			_, err = io.Copy(part, file)
			if err != nil {
				panic(err)
			}
		}
	}

	for _, data := range datas {
		for k, v := range data {
			w.WriteField(k, v)
		}
	}

	w.Close()
	// set file header example:
	// "Content-Type": "multipart/form-data; boundary=------------------------7d87eceb5520850c",
	req.httpreq.Body = ioutil.NopCloser(bytes.NewReader(b.Bytes()))
	req.httpreq.ContentLength = int64(b.Len())
	req.Header.Set("Content-Type", w.FormDataContentType())
}

// build post Form data
func (req *Request) buildForms(datas ...map[string]string) (Forms url.Values) {
	Forms = url.Values{}
	for _, data := range datas {
		for key, value := range data {
			Forms.Add(key, value)
		}
	}
	return Forms
}

// open file for post upload files

func openFile(filename string) *os.File {
	r, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	return r
}
