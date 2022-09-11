## CookieJar
A cookiejar drive with Go.

[![996.icu](https://img.shields.io/badge/link-996.icu-red.svg)](https://996.icu)
[![LICENSE](https://img.shields.io/badge/license-NPL%20(The%20996%20Prohibited%20License)-blue.svg)](https://github.com/996icu/996.ICU/blob/master/LICENSE)

## Example

```go
package main

import (
	"fmt"
	"net/http"
	"io/ioutil"
	"github.com/telanflow/cookiejar"
)

func main() {

	// Netscape HTTP Cookie File
	jar, _ = cookiejar.NewFileJar("cookie.txt", nil)

	client := &http.Client{
		Jar: jar,
	}

	req, _ := http.NewRequest(http.MethodGet, "https://telan.me", nil)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer resp.Body.Close()

	c, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(c))
}
```
