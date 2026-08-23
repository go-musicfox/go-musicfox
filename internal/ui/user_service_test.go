package ui

import (
	"path/filepath"
	"testing"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	cookiejar "github.com/juju/persistent-cookiejar"
)

// seqRecorder records startup events in call order so tests can assert the
// InitHook order constraints (3.1.3).
type seqRecorder struct {
	events []string
}

func (r *seqRecorder) record(ev string) {
	r.events = append(r.events, ev)
}

func (r *seqRecorder) indexOf(ev string) int {
	for i, e := range r.events {
		if e == ev {
			return i
		}
	}
	return -1
}

// TestJarInitPrecedesUserCallback asserts the cross-service order constraint:
// loginService.InitJar (jar 初始化) MUST run before the userService callback
// (LoginCallback persists the jar via appCookieJar.Save()). The sequence is
// InitJar → LoadFromStorage → LoginWithCookie, and the recorded order must put
// "jar-init" before "user-callback".
func TestJarInitPrecedesUserCallback(t *testing.T) {
	previousConfig := configs.AppConfig
	configs.AppConfig = &configs.Config{}
	t.Cleanup(func() { configs.AppConfig = previousConfig })

	rec := &seqRecorder{}
	var jarSlot *cookiejar.Jar
	var userSlot *structs.User

	loginSvc := &LoginService{CookieJar: &jarSlot, User: &userSlot}
	userSvc := &UserService{
		User:           &userSlot,
		Jar:            &jarSlot,
		Login:          func() error { rec.record("user-callback"); return nil },
		loadStoredUser: func() (*structs.User, bool) { return nil, false },
		refreshJar:     func() (*cookiejar.Jar, error) { return jarSlot, nil },
	}

	// A cookie in the env drives LoginWithCookie through the cookie-login path,
	// reaching the Login callback after a (stubbed) successful token refresh.
	t.Setenv("MUSICFOX_COOKIE", "MUSIC_U=test-cookie")

	jar, err := loginSvc.InitJar(filepath.Join(t.TempDir(), "cookie"))
	if err != nil {
		t.Fatalf("InitJar() error = %v", err)
	}
	rec.record("jar-init")

	userSvc.LoadFromStorage(jar)
	userSvc.LoginWithCookie(jar, filepath.Join(t.TempDir(), "cookie"))

	jarIdx := rec.indexOf("jar-init")
	cbIdx := rec.indexOf("user-callback")
	if jarIdx == -1 {
		t.Fatalf("jar-init was not recorded; events=%v", rec.events)
	}
	if cbIdx == -1 {
		t.Fatalf("user-callback was not recorded (LoginWithCookie must reach the callback); events=%v", rec.events)
	}
	if jarIdx >= cbIdx {
		t.Fatalf("order constraint violated: jar-init(%d) must precede user-callback(%d); events=%v", jarIdx, cbIdx, rec.events)
	}
}

// TestUserServiceTouristMode verifies the guest path: without a stored user and
// without a cookie, LoadFromStorage + LoginWithCookie leave the user slot nil,
// never invoke the Login callback and never panic.
func TestUserServiceTouristMode(t *testing.T) {
	previousConfig := configs.AppConfig
	configs.AppConfig = &configs.Config{}
	t.Cleanup(func() { configs.AppConfig = previousConfig })

	rec := &seqRecorder{}
	var jarSlot *cookiejar.Jar
	var userSlot *structs.User

	loginSvc := &LoginService{CookieJar: &jarSlot, User: &userSlot}
	userSvc := &UserService{
		User:           &userSlot,
		Jar:            &jarSlot,
		Login:          func() error { rec.record("user-callback"); return nil },
		loadStoredUser: func() (*structs.User, bool) { return nil, false },
		refreshJar:     func() (*cookiejar.Jar, error) { return jarSlot, nil },
	}

	t.Setenv("MUSICFOX_COOKIE", "") // no cookie → tourist mode

	jar, err := loginSvc.InitJar(filepath.Join(t.TempDir(), "cookie"))
	if err != nil {
		t.Fatalf("InitJar() error = %v", err)
	}

	userSvc.LoadFromStorage(jar)
	userSvc.LoginWithCookie(jar, filepath.Join(t.TempDir(), "cookie"))

	if userSlot != nil {
		t.Fatalf("tourist mode: user should stay nil, got %v", userSlot)
	}
	if cbIdx := rec.indexOf("user-callback"); cbIdx != -1 {
		t.Fatalf("tourist mode: Login callback must not fire, got event at %d; events=%v", cbIdx, rec.events)
	}
}
