package lastfm_go

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

func requireAuth(params *apiParams) (err error) {
	if params.sk == "" {
		err = newLibError(
			ErrorAuthRequired,
			Messages[ErrorAuthRequired],
		)
	}
	return
}

/*
func checkRequiredParams(params P, required ...string) (err error) {
    var missing []string
    ng := false
    for _, p := range required {
        if _, ok := params[p]; !ok {
            missing = append(missing, p)
            ng = true
        }
    }
    if ng {
        err = newLibError(
            ErrorParameterMissing,
            fmt.Sprintf(Messages[ErrorParameterMissing], required, missing),
        )
    }
    return
}
*/

func constructUrl(base string, params url.Values) (uri string) {
	//if ResponseFormat == "json" {
	//params.Add("format", ResponseFormat)
	//}
	p := params.Encode()
	uri = base + "?" + p
	return
}

func toString(val interface{}) (str string, err error) {
	switch val.(type) {
	case string:
		str = val.(string)
	case int:
		str = strconv.Itoa(val.(int))
	case []string:
		ss := val.([]string)
		if len(ss) > 10 {
			ss = ss[:10]
		}
		str = strings.Join(ss, ",")
	default:
		err = newLibError(
			ErrorInvalidTypeOfArgument,
			Messages[ErrorInvalidTypeOfArgument],
		)
	}
	return
}

func parseResponse(body []byte, result interface{}) (err error) {
	var base Base
	err = xml.Unmarshal(body, &base)
	if err != nil {
		return
	}
	if base.Status == ApiResponseStatusFailed {
		var errorDetail ApiError
		err = xml.Unmarshal(base.Inner, &errorDetail)
		if err != nil {
			return
		}
		err = newApiError(&errorDetail)
		return
	} else if result == nil {
		return
	}
	err = xml.Unmarshal(base.Inner, result)
	return
}

func getSignature(params map[string]string, secret string) (sig string) {
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sigPlain string
	for _, k := range keys {
		sigPlain += k + params[k]
	}
	sigPlain += secret

	hasher := md5.New()
	hasher.Write([]byte(sigPlain))
	sig = hex.EncodeToString(hasher.Sum(nil))
	return
}

func formatArgs(args, rules P) (result map[string]string, err error) {

	result = make(map[string]string)
	if _, ok := rules["indexing"]; ok {

		for _, p := range rules["indexing"].([]string) {
			if valI, ok := args[p]; ok {
				switch valI.(type) {
				case string:
					key := p + "[0]"
					val := valI.(string)
					result[key] = val
				case int:
					key := p + "[0]"
					val := strconv.Itoa(valI.(int))
					result[key] = val
				case int64: //timestamp
					key := p + "[0]"
					val := strconv.FormatInt(valI.(int64), 10)
					result[key] = val
				case []string: //with indeces
					for i, val := range valI.([]string) {
						key := fmt.Sprintf("%s[%d]", p, i)
						result[key] = val
					}
				default:
					err = newLibError(
						ErrorInvalidTypeOfArgument,
						Messages[ErrorInvalidTypeOfArgument],
					)
					break
				}
			} else if _, ok := args[p+"[0]"]; ok {
				for i := 0; ; i++ {
					key := fmt.Sprintf("%s[%d]", p, i)
					if valI, ok := args[key]; ok {
						var val string
						switch valI.(type) {
						case string:
							val = valI.(string)
						case int:
							val = strconv.Itoa(valI.(int))
						case int64:
							val = strconv.FormatInt(valI.(int64), 10)
						default:
							err = newLibError(
								ErrorInvalidTypeOfArgument,
								Messages[ErrorInvalidTypeOfArgument],
							)
							break
						}
						result[key] = val
					}
				}
			}
			if err != nil {
				break
			}
		}
	}
	if err != nil {
		return
	}

	if _, ok := rules["plain"]; ok {
		for _, key := range rules["plain"].([]string) {
			if valI, ok := args[key]; ok {
				var val string
				switch valI.(type) {
				case string:
					val = valI.(string)
				case int:
					val = strconv.Itoa(valI.(int))
				case int64:
					val = strconv.FormatInt(valI.(int64), 10)
				case []string: //comma delimited
					ss := valI.([]string)
					if len(ss) > 10 {
						ss = ss[:10]
					}
					val = strings.Join(ss, ",")
				default:
					err = newLibError(
						ErrorInvalidTypeOfArgument,
						Messages[ErrorInvalidTypeOfArgument],
					)
					break
				}
				result[key] = val
			}
		}
	}
	if err != nil {
		return
	}
	return
}

// ///////////
// GET API //
// ///////////
func callGet(apiMethod string, params *apiParams, args map[string]interface{}, result interface{}, rules P) (err error) {
	urlParams := url.Values{}
	urlParams.Add("method", apiMethod)
	urlParams.Add("api_key", params.apikey)

	formated, err := formatArgs(args, rules)
	if err != nil {
		return
	}
	for k, v := range formated {
		urlParams.Add(k, v)
	}

	uri := constructUrl(UriApiSecBase, urlParams)

	client := http.DefaultClient
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return
	}
	if params.useragent != "" {
		req.Header.Set("User-Agent", params.useragent)
	}

	res, err := client.Do(req)
	if err != nil {
		return
	}
	if res.StatusCode/100 == 5 { // only 5xx class errors
		err = newLibError(res.StatusCode, res.Status)
		return
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}
	err = parseResponse(body, result)
	return
}

// ////////////
// POST API //
// ////////////
func callPost(apiMethod string, params *apiParams, args P, result interface{}, rules P) (err error) {
	if err = requireAuth(params); err != nil {
		return
	}

	urlParams := url.Values{}
	uri := constructUrl(UriApiSecBase, urlParams)

	//post data
	postData := url.Values{}
	postData.Add("method", apiMethod)
	postData.Add("api_key", params.apikey)
	postData.Add("sk", params.sk)

	tmp := make(map[string]string)
	tmp["method"] = apiMethod
	tmp["api_key"] = params.apikey
	tmp["sk"] = params.sk

	formated, err := formatArgs(args, rules)
	for k, v := range formated {
		tmp[k] = v
		postData.Add(k, v)
	}

	sig := getSignature(tmp, params.secret)
	postData.Add("api_sig", sig)

	client := http.DefaultClient
	req, err := http.NewRequest("POST", uri, strings.NewReader(postData.Encode()))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if params.useragent != "" {
		req.Header.Set("User-Agent", params.useragent)
	}

	res, err := client.Do(req)
	//res, err := http.PostForm(uri, postData)
	if err != nil {
		return
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}
	err = parseResponse(body, result)
	return
}

func callPostWithoutSession(apiMethod string, params *apiParams, args P, result interface{}, rules P) (err error) {
	urlParams := url.Values{}
	uri := constructUrl(UriApiSecBase, urlParams)

	//post data
	postData := url.Values{}
	postData.Add("method", apiMethod)
	postData.Add("api_key", params.apikey)

	tmp := make(map[string]string)
	tmp["method"] = apiMethod
	tmp["api_key"] = params.apikey

	formated, err := formatArgs(args, rules)
	for k, v := range formated {
		tmp[k] = v
		postData.Add(k, v)
	}

	sig := getSignature(tmp, params.secret)
	postData.Add("api_sig", sig)

	//call API
	res, err := http.PostForm(uri, postData)
	if err != nil {
		return
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}
	err = parseResponse(body, result)
	return
}
