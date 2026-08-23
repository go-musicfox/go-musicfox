package ui

import (
	"strings"
	"testing"
)

// init registers the after-anchor test-doubles every NewMainMenu construction
// in this binary needs. The built-in entries anchor on plugin keys (搜索 after
// album_menu, 帮助 after last_fm) that the ui test binary cannot link (ui must
// not import plugins), so the binary registers behavior-equivalent items to
// keep the chain complete. The chain here mirrors the production 16-item order
// minus 检查更新: _main_start → daily_songs → daily_playlists → user_playlist →
// user_collect → personal_fm → album_menu → search_type → ranks →
// high_quality_playlists → hot_artists → recent_songs → could → radio_dj_type
// → last_fm → help.
func init() {
	chain := []struct{ key, title, after string }{
		{"daily_songs", "每日推荐歌曲", MainMenuStart},
		{"daily_playlists", "每日推荐歌单", "daily_songs"},
		{"user_playlist", "我的歌单", "daily_playlists"},
		{"user_collect", "我的收藏", "user_playlist"},
		{"personal_fm", "私人FM", "user_collect"},
		{"album_menu", "专辑列表", "personal_fm"},
		{"ranks", "排行榜", "search_type"},
		{"high_quality_playlists", "精选歌单", "ranks"},
		{"hot_artists", "热门歌手", "high_quality_playlists"},
		{"recent_songs", "最近播放歌曲", "hot_artists"},
		{"could", "云盘", "recent_songs"},
		{"radio_dj_type", "主播电台", "could"},
		{"last_fm", "LastFM", "radio_dj_type"},
	}
	for _, c := range chain {
		RegisterMenu(c.key, func(base baseMenu, _ NoArgMenuOpts) (Menu, error) {
			return &testCheckUpdateMenu{baseMenu: base}, nil
		})
		RegisterMainMenuItemAfter(c.key, c.title, c.after, nil)
	}
}

// chainKeysOf returns the ordered entry keys for assertion messages.
func chainKeysOf(entries []mainMenuEntry) []string {
	keys := make([]string, len(entries))
	for i, e := range entries {
		keys[i] = e.key
	}
	return keys
}

// --- orderMainMenuEntries unit tests (after-anchor chain) ---

// TestOrderMainMenuEntriesChain walks the full 16-item production chain in any
// registration order (the input is reversed to prove the walk does not depend
// on registration order).
func TestOrderMainMenuEntriesChain(t *testing.T) {
	entries := []mainMenuEntry{
		{key: "daily_songs", after: MainMenuStart, title: "每日推荐歌曲"},
		{key: "daily_playlists", after: "daily_songs", title: "每日推荐歌单"},
		{key: "user_playlist", after: "daily_playlists", title: "我的歌单"},
		{key: "user_collect", after: "user_playlist", title: "我的收藏"},
		{key: "personal_fm", after: "user_collect", title: "私人FM"},
		{key: "album_menu", after: "personal_fm", title: "专辑列表"},
		{key: "search_type", after: "album_menu", title: "搜索", builtin: true},
		{key: "ranks", after: "search_type", title: "排行榜"},
		{key: "high_quality_playlists", after: "ranks", title: "精选歌单"},
		{key: "hot_artists", after: "high_quality_playlists", title: "热门歌手"},
		{key: "recent_songs", after: "hot_artists", title: "最近播放歌曲"},
		{key: "could", after: "recent_songs", title: "云盘"},
		{key: "radio_dj_type", after: "could", title: "主播电台"},
		{key: "last_fm", after: "radio_dj_type", title: "LastFM"},
		{key: "help", after: "last_fm", title: "帮助", builtin: true},
		{key: "check_update", after: "help", title: "检查更新"},
	}
	// 反序输入（模拟任意 init() 注册顺序）也必须走通链。
	shuffled := append([]mainMenuEntry(nil), entries...)
	for i, j := 0, len(shuffled)-1; i < j; i, j = i+1, j-1 {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}
	got := orderMainMenuEntries(shuffled)
	if len(got) != len(entries) {
		t.Fatalf("chain length = %d, want %d: %v", len(got), len(entries), chainKeysOf(got))
	}
	for i, e := range entries {
		if got[i].key != e.key {
			t.Fatalf("chain[%d] = %q, want %q (full: %v)", i, got[i].key, e.key, chainKeysOf(got))
		}
	}
}

// TestOrderMainMenuEntriesInsertion proves the insertion scenario: a new entry
// declaring After = help lands right after 帮助 with a single-anchor change —
// nothing renumbers.
func TestOrderMainMenuEntriesInsertion(t *testing.T) {
	base := []mainMenuEntry{
		{key: "search_type", after: MainMenuStart, title: "搜索", builtin: true},
		{key: "help", after: "search_type", title: "帮助", builtin: true},
	}
	got := orderMainMenuEntries(append(base, mainMenuEntry{key: "new_item", after: "help", title: "新项"}))
	want := []string{"search_type", "help", "new_item"}
	for i, key := range want {
		if got[i].key != key {
			t.Fatalf("chain[%d] = %q, want %q (full: %v)", i, got[i].key, key, chainKeysOf(got))
		}
	}
}

// TestOrderMainMenuEntriesEndAppend proves the empty-After convenience forms
// append at the end of the chain in registration order (the pre-chain "append
// after built-ins" behavior).
func TestOrderMainMenuEntriesEndAppend(t *testing.T) {
	entries := []mainMenuEntry{
		{key: "first", after: MainMenuStart, title: "第一项"},
		{key: "end_a", title: "末尾A"},
		{key: "second", after: "first", title: "第二项"},
		{key: "end_b", title: "末尾B"},
	}
	got := orderMainMenuEntries(entries)
	want := []string{"first", "second", "end_a", "end_b"}
	for i, key := range want {
		if got[i].key != key {
			t.Fatalf("chain[%d] = %q, want %q (full: %v)", i, got[i].key, key, chainKeysOf(got))
		}
	}
}

func TestOrderMainMenuEntriesMissingAnchorPanics(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("orderMainMenuEntries(missing anchor) did not panic")
		}
		msg, ok := r.(string)
		if !ok || !strings.Contains(msg, "After anchor not registered") || !strings.Contains(msg, "new_item") {
			t.Fatalf("panic message = %v, want it to list the missing anchor", r)
		}
	}()
	orderMainMenuEntries([]mainMenuEntry{
		{key: "first", after: MainMenuStart, title: "第一项"},
		{key: "new_item", after: "no_such_anchor", title: "新项"},
	})
}

// TestOrderMainMenuEntriesCyclePanics proves a cycle reached while walking from
// MainMenuStart is detected (an entry re-visited). Registration rejects
// duplicate keys, so the cycle is built with a duplicated key — the chain-level
// assertion is the defensive backstop.
func TestOrderMainMenuEntriesCyclePanics(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("orderMainMenuEntries(cycle) did not panic")
		}
		msg, ok := r.(string)
		if !ok || !strings.Contains(msg, "cycle") {
			t.Fatalf("panic message = %v, want it to report the cycle", r)
		}
	}()
	orderMainMenuEntries([]mainMenuEntry{
		{key: "a", after: MainMenuStart, title: "A"},
		{key: "b", after: "a", title: "B"},
		{key: "a", after: "b", title: "A2"}, // duplicate key closes the cycle
	})
}

// TestOrderMainMenuEntriesOrphanPanics proves chain-length mismatch detection:
// two entries declaring the same After anchor collapse to one walk, orphaning
// the other.
func TestOrderMainMenuEntriesOrphanPanics(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("orderMainMenuEntries(orphan) did not panic")
		}
		msg, ok := r.(string)
		if !ok || !strings.Contains(msg, "reachable") || !strings.Contains(msg, "orphan") {
			t.Fatalf("panic message = %v, want it to report the orphaned entry", r)
		}
	}()
	orderMainMenuEntries([]mainMenuEntry{
		{key: "first", after: MainMenuStart, title: "第一项"},
		{key: "second", after: "first", title: "第二项"},
		{key: "orphan", after: "first", title: "孤儿"}, // duplicate After anchor
	})
}

// --- NewMainMenu integration (chain-driven) ---
//
// The registry-level insertion scenario lives in TestNewMainMenuChainOrdersPluginItems
// (registry_test.go): in this binary's chain the leaf anchor is 帮助, so a test
// item anchored after "help" lands right after it without renumbering the
// existing items.
