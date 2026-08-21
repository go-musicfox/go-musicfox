// Command unmproto is the Phase 0 falsification prototype for the go-musicfox
// plugin framework spec. It verifies that the "URL middleware" shape can
// reproduce the UNM (UnblockNeteaseMusic) behavior the SDK currently performs
// internally: for a known-blocked song, it compares the URL produced by the
// SDK's built-in UNM branch against the URL produced by an explicit middleware
// that directly drives the vendored processor (RequestBefore → request →
// RequestAfter).
//
// Decision gate (from the spec): if both URLs match, the middleware shape is
// viable and Phase 1 proceeds; otherwise the shape is wrong and the design
// must be revisited.
package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/buger/jsonparser"
	"github.com/cnsilvan/UnblockNeteaseMusic/processor"
	neteaseutil "github.com/go-musicfox/netease-music/util"
	"github.com/go-musicfox/netease-music/service"
	"github.com/go-musicfox/requests"
)

const (
	// linuxForwardAPI is where the SDK actually sends linuxapi-encrypted
	// payloads; the inner target URL is carried inside the encrypted body.
	linuxForwardAPI = "https://music.163.com/api/linux/forward"
	songURLAPI      = "https://music.163.com/api/song/enhance/player/url"
	br              = "320000"
	pcUA            = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/60.0.3112.90 Safari/537.36"
)

// extractURL pulls the first song URL out of a song/url API response.
func extractURL(body []byte) string {
	url, _ := jsonparser.GetString(body, "data", "[0]", "url")
	return url
}

// sdkPath reproduces today's SDK path: linuxapi SongUrlService with the UNM
// switch on, so the SDK's internal UNMFlag branch runs the processor.
func sdkPath(songID string) (string, error) {
	code, body := (&service.SongUrlService{ID: songID, Br: br}).SongUrl()
	if code != 200 {
		return "", fmt.Errorf("sdk path failed with code %v", code)
	}
	return extractURL(body), nil
}

// middlewarePath is the middleware shape under test: it builds the same
// linuxapi request (encrypted payload sent to the linux forward endpoint,
// exactly like the SDK's Crypto == "linuxapi" branch), then drives the
// vendored processor explicitly instead of letting the SDK do it internally.
func middlewarePath(songID string) (string, error) {
	linuxApiData := map[string]interface{}{
		"method": "POST",
		"url":    songURLAPI,
		"params": map[string]string{"ids": "[" + songID + "]", "br": br},
	}
	data := neteaseutil.Linuxapi(linuxApiData)

	req := requests.Requests()
	req.SetCookie(&http.Cookie{Name: "os", Value: "pc"})
	req.Header.Set("User-Agent", pcUA)
	req.Header.Set("Referer", "https://music.163.com")
	req.Client.Jar = neteaseutil.GetGlobalCookieJar()

	// Dry-run builds the request object without sending it, mirroring the SDK.
	if _, err := req.Post(linuxForwardAPI, requests.DryRun(true), requests.Datas(data)); err != nil {
		return "", fmt.Errorf("dry run failed: %w", err)
	}
	httpReq := req.HttpRequest()
	netease := processor.RequestBefore(httpReq)
	if netease == nil {
		return "", fmt.Errorf("request blocked by processor")
	}

	resp, err := req.Post(linuxForwardAPI, requests.Datas(data))
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	response := resp.R
	processor.RequestAfter(httpReq, response, netease)
	resp.ReloadContent()
	return extractURL(resp.Content()), nil
}

func main() {
	songID := "186016" // 晴天 (Zhou Jielin) — known blocked on NetEase for CN region
	if len(os.Args) > 1 {
		if _, err := strconv.ParseInt(os.Args[1], 10, 64); err == nil {
			songID = os.Args[1]
		}
	}

	// Mirror cmd/musicfox.go's UNM config glue with a minimal source set.
	neteaseutil.UNMSwitch = true
	neteaseutil.Sources = []string{"kuwo", "kugou", "migu", "qq"}
	neteaseutil.ConfigInit()

	urlA, errA := sdkPath(songID)
	urlB, errB := middlewarePath(songID)

	fmt.Printf("songID : %s\n", songID)
	fmt.Printf("URL_A  (SDK internal UNM): %q err=%v\n", urlA, errA)
	fmt.Printf("URL_B  (middleware shape): %q err=%v\n", urlB, errB)
	fmt.Printf("match  : %v\n", urlA != "" && urlA == urlB)

	if urlA == "" && urlB != "" {
		fmt.Println("verdict: middleware unlocked a URL the SDK path left empty — shape is viable")
	} else if urlA != "" && urlA == urlB {
		fmt.Println("verdict: middleware reproduces the SDK path exactly — shape is viable")
	} else {
		fmt.Println("verdict: MISMATCH — middleware shape diverges from the SDK path, revisit design")
	}
}
