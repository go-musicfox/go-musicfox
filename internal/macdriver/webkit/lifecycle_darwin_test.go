//go:build darwin

package webkit

import (
	"fmt"
	"testing"
	"time"

	"github.com/ebitengine/purego/objc"

	"github.com/go-musicfox/go-musicfox/internal/macdriver/core"
)

// TestCookieStoreLifecycle reproduces the login_webview flow: create a
// non-persistent data store, attach it to a config, grab the cookie store,
// then call getAllCookiesWithCompletionHandler: from a non-main goroutine
// after the creating Autorelease pool has popped.
func TestCookieStoreLifecycle(t *testing.T) {
	var block objc.Block
	core.Autorelease(func() {
		ds := WKWebsiteDataStore_NonPersistentDataStore()
		cfg := WKWebViewConfiguration_alloc().Init()
		cfg.SetWebsiteDataStore(ds)
		cs := ds.HttpCookieStore()
		fmt.Println("cookieStore ID:", cs.ID, "config ID:", cfg.ID)
		block = objc.NewBlock(func(_ objc.Block, cookies objc.ID) {
			fmt.Println("callback 1 fired, cookies:", cookies)
		})
		fmt.Println("calling getAllCookiesWithCompletionHandler inside pool...")
		cs.GetAllCookiesWithCompletionHandler(block)
		time.Sleep(2 * time.Second)
	})
	fmt.Println("pool popped, calling again from same goroutine...")
	ds2 := WKWebsiteDataStore_NonPersistentDataStore()
	cs2 := ds2.HttpCookieStore()
	cs2.GetAllCookiesWithCompletionHandler(block)
	time.Sleep(2 * time.Second)
	fmt.Println("DONE")
}
