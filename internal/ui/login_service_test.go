package ui

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"

	neteaseutil "github.com/go-musicfox/netease-music/util"
	cookiejar "github.com/juju/persistent-cookiejar"

	"github.com/go-musicfox/go-musicfox/internal/structs"
)

// TestLoginServiceInitJarCreatesAndSyncsJar verifies InitJar is invocable
// standalone (as the startup sequence requires: jar 先于 userService 回调):
// it creates a persistent jar, assigns the appCookieJar slot and syncs the
// netease-music global cookie jar.
func TestLoginServiceInitJarCreatesAndSyncsJar(t *testing.T) {
	var jarSlot *cookiejar.Jar
	var userSlot *structs.User
	svc := &LoginService{CookieJar: &jarSlot, User: &userSlot}

	cookiePath := filepath.Join(t.TempDir(), "cookie")
	jar, err := svc.InitJar(cookiePath)
	if err != nil {
		t.Fatalf("InitJar() error = %v", err)
	}
	if jar == nil {
		t.Fatal("InitJar() returned nil jar")
	}
	if jarSlot != jar {
		t.Fatal("appCookieJar slot was not assigned")
	}
	if got := neteaseutil.GetGlobalCookieJar(); got != http.CookieJar(jar) {
		t.Fatalf("global cookie jar not synced: got %v, want %v", got, jar)
	}
}

// TestLoginServiceInitJarBacksUpCorruptFile verifies the corrupt-file
// backup/reset logic: a corrupt cookie file is renamed to a .bak.<ts> backup
// and a fresh jar is created at the original path.
func TestLoginServiceInitJarBacksUpCorruptFile(t *testing.T) {
	var jarSlot *cookiejar.Jar
	var userSlot *structs.User
	svc := &LoginService{CookieJar: &jarSlot, User: &userSlot}

	cookiePath := filepath.Join(t.TempDir(), "cookie")
	if err := os.WriteFile(cookiePath, []byte("this is not a valid cookie jar"), 0600); err != nil {
		t.Fatalf("write corrupt cookie file: %v", err)
	}

	jar, err := svc.InitJar(cookiePath)
	if err != nil {
		t.Fatalf("InitJar() on corrupt file error = %v", err)
	}
	if jar == nil {
		t.Fatal("InitJar() returned nil jar after corrupt reset")
	}
	if jarSlot != jar {
		t.Fatal("appCookieJar slot was not assigned after corrupt reset")
	}

	backups, err := filepath.Glob(cookiePath + ".bak.*")
	if err != nil || len(backups) != 1 {
		t.Fatalf("expected exactly one .bak backup, got %v (err %v)", backups, err)
	}
}
