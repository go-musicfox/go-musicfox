package ui

import (
	"io"
	"net/http"
	"net/http/cookiejar"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/anhoder/foxful-cli/model"
	neteaseutil "github.com/go-musicfox/netease-music/util"

	"github.com/go-musicfox/go-musicfox/internal/structs"
)

type dailyPlaylistsRoundTripper func(*http.Request) (*http.Response, error)

func (f dailyPlaylistsRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func mockDailyPlaylistsResponses(t *testing.T, responses ...string) *int {
	t.Helper()
	previousTransport := http.DefaultTransport
	previousJar := neteaseutil.GetGlobalCookieJar()
	previousProxy := neteaseutil.UNMProxyURL
	previousUNM := neteaseutil.UNMSwitch
	jar, _ := cookiejar.New(nil)
	neteaseutil.SetGlobalCookieJar(jar)
	neteaseutil.UNMProxyURL = ""
	neteaseutil.UNMSwitch = false
	t.Cleanup(func() {
		http.DefaultTransport = previousTransport
		neteaseutil.SetGlobalCookieJar(previousJar)
		neteaseutil.UNMProxyURL = previousProxy
		neteaseutil.UNMSwitch = previousUNM
	})
	requests := new(int)
	http.DefaultTransport = dailyPlaylistsRoundTripper(func(request *http.Request) (*http.Response, error) {
		if *requests >= len(responses) {
			t.Fatalf("unexpected request %d: %s", *requests+1, request.URL)
		}
		body := responses[*requests]
		*requests++
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(body)),
			Request:    request,
		}, nil
	})
	return requests
}

func TestDailyRecommendPlaylistsCachesOnlyCurrentDate(t *testing.T) {
	requests := mockDailyPlaylistsResponses(t,
		`{"code":200,"data":{"blockData":{"creatives":[{"creativeType":"list","resources":[{"resourceType":"list","resourceId":"1","uiElement":{"mainTitle":{"title":"First"}}}]}]}}}`,
		`{"code":200,"data":{"blockData":{"creatives":[{"creativeType":"list","resources":[{"resourceType":"list","resourceId":"2","uiElement":{"mainTitle":{"title":"Second"}}}]}]}}}`,
	)
	menu := NewDailyRecommendPlaylistMenu(newBaseMenu(&Netease{user: &structs.User{UserId: 1}}))
	hook := menu.BeforeEnterMenuHook()
	for range 2 {
		if ok, page := hook(nil); !ok || page != nil {
			t.Fatalf("enter menu = (%v, %T), want successful entry", ok, page)
		}
	}
	if *requests != 1 {
		t.Fatalf("same-day entries made %d requests, want 1", *requests)
	}
	menu.fetchTime = time.Now().AddDate(0, 0, -1)
	if ok, _ := hook(nil); !ok {
		t.Fatal("next-day entry failed")
	}
	if *requests != 2 || len(menu.playlists) != 1 || menu.playlists[0].Id != 2 || menu.menus[0].Title != "Second" {
		t.Fatalf("next-day entry made %d requests and returned %v / %v", *requests, menu.playlists, menu.menus)
	}
}

func TestDailyRecommendPlaylistsRetriesAfterFailedFetch(t *testing.T) {
	requests := mockDailyPlaylistsResponses(t,
		`{"code":500}`,
		`{"code":200,"data":{"blockData":{"creatives":[{"creativeType":"list","resources":[{"resourceType":"list","resourceId":"2","uiElement":{"mainTitle":{"title":"Recovered"}}}]}]}}}`,
	)
	menu := NewDailyRecommendPlaylistMenu(newBaseMenu(&Netease{user: &structs.User{UserId: 1}}))
	menu.fetchTime = time.Now().AddDate(0, 0, -1)
	menu.playlists = []structs.Playlist{{Id: 1, Name: "Previous day"}}
	menu.menus = []model.MenuItem{{Title: "Previous day"}}
	previousFetchTime := menu.fetchTime
	hook := menu.BeforeEnterMenuHook()
	if ok, page := hook(nil); ok || page != nil {
		t.Fatalf("failed fetch = (%v, %T), want failure without navigation", ok, page)
	}
	if menu.fetchTime != previousFetchTime || menu.playlists[0].Id != 1 || menu.menus[0].Title != "Previous day" {
		t.Fatal("failed fetch changed the cached date or playlists")
	}
	if ok, _ := hook(nil); !ok || *requests != 2 || menu.playlists[0].Id != 2 {
		t.Fatalf("retry: ok=%v, requests=%d, playlists=%v", ok, *requests, menu.playlists)
	}
}

func TestDailyRecommendPlaylistsPreservesCacheForInvalidSuccessfulResponse(t *testing.T) {
	for _, tc := range []struct {
		name     string
		response string
	}{
		{name: "empty creatives", response: `{"code":200,"data":{"blockData":{"creatives":[]}}}`},
		{name: "legacy shape", response: `{"code":200,"recommend":[{"id":9,"name":"Old endpoint"}]}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			requests := mockDailyPlaylistsResponses(t, tc.response,
				`{"code":200,"data":{"blockData":{"creatives":[{"creativeType":"list","resources":[{"resourceType":"list","resourceId":"2","uiElement":{"mainTitle":{"title":"Recovered"}}}]}]}}}`,
			)
			menu := NewDailyRecommendPlaylistMenu(newBaseMenu(&Netease{user: &structs.User{UserId: 1}}))
			previousFetchTime := time.Now().AddDate(0, 0, -1)
			previousPlaylists := []structs.Playlist{{Id: 1, Name: "Previous day"}}
			previousMenus := []model.MenuItem{{Title: "Previous day"}}
			menu.fetchTime = previousFetchTime
			menu.playlists = previousPlaylists
			menu.menus = previousMenus
			hook := menu.BeforeEnterMenuHook()
			if ok, page := hook(nil); ok || page != nil {
				t.Fatalf("invalid successful response = (%v, %T), want failure without navigation", ok, page)
			}
			if menu.fetchTime != previousFetchTime || !reflect.DeepEqual(menu.playlists, previousPlaylists) || !reflect.DeepEqual(menu.menus, previousMenus) {
				t.Fatal("invalid successful response changed the cached date or playlists")
			}
			if ok, _ := hook(nil); !ok || *requests != 2 {
				t.Fatalf("retry: ok=%v, requests=%d, want successful second request", ok, *requests)
			}
			if len(menu.playlists) != 1 || menu.playlists[0].Id != 2 || len(menu.menus) != 1 || menu.menus[0].Title != "Recovered" || !menu.fetchTime.After(previousFetchTime) {
				t.Fatalf("retry did not replace the cached recommendations: %v / %v at %v", menu.playlists, menu.menus, menu.fetchTime)
			}
		})
	}
}

func TestDailyRecommendPlaylistsContextRefreshFetchesAgain(t *testing.T) {
	requests := mockDailyPlaylistsResponses(t,
		`{"code":200,"data":{"blockData":{"creatives":[{"creativeType":"list","resources":[{"resourceType":"list","resourceId":"2","uiElement":{"mainTitle":{"title":"Refreshed"}}}]}]}}}`,
	)
	app, netease := newFormPageTestApp(t)
	netease.user = &structs.User{UserId: 1}
	menu := NewDailyRecommendPlaylistMenu(newBaseMenu(netease))
	menu.fetchTime = time.Now()
	menu.playlists = []structs.Playlist{{Id: 1, Name: "Cached"}}
	menu.menus = []model.MenuItem{{Title: "Cached"}}
	main := app.MustMain()
	if page := main.EnterMenu(menu, &model.MenuItem{Title: "Daily playlists"}); page != nil {
		t.Fatalf("enter cached menu returned %T", page)
	}
	page, cmd := handleGenericContextAction(netease, app, "generic:refresh")
	if page != nil || cmd == nil {
		t.Fatalf("context refresh: page=%T cmd=%v, want deferred rerender command", page, cmd)
	}
	if *requests != 0 {
		t.Fatal("context refresh requested data before rendering its loading state")
	}
	_, _ = main.Update(cmd(), app)
	if *requests != 1 || len(menu.playlists) != 1 || menu.playlists[0].Id != 2 {
		t.Fatalf("manual refresh made %d requests and returned %v", *requests, menu.playlists)
	}
	if items := main.MenuList(); len(items) != 1 || items[0].Title != "Refreshed" {
		t.Fatalf("rendered menu = %v, want Refreshed", items)
	}
}
