package ui

import (
	"testing"

	"github.com/anhoder/foxful-cli/model"
)

func TestBuildGroupItems(t *testing.T) {
	actions := []ActionItem{
		{title: model.MenuItem{Title: "下载"}, group: "download"},
		{title: model.MenuItem{Title: "分享", Subtitle: "当前歌曲"}, group: "share"},
	}

	items := buildGroupItems("sel", "当前选中：「歌曲」", actions, false)
	// 移除段内分隔符后：Header + 2个操作项 = 3项（不再有 group 分隔线）
	if len(items) != 3 {
		t.Fatalf("context item count = %d, want 3", len(items))
	}
	if !items[0].Header || items[0].Label != "当前选中：「歌曲」" {
		t.Errorf("header = %#v, want current selection header", items[0])
	}
	if got, want := items[1].ID, "sel:0"; got != want {
		t.Errorf("first context ID = %q, want %q", got, want)
	}
	if got, want := items[2].Label, itemIndent+"分享 当前歌曲"; got != want {
		t.Errorf("second context label = %q, want %q", got, want)
	}
}

func TestGenericContextMenuItemsExcludePlaybackControls(t *testing.T) {
	items := appendContextMenuGlobalItems(nil, false)
	if len(items) != 2 {
		t.Fatalf("empty-playlist generic item count = %d, want refresh + switchTheme", len(items))
	}
	if got, want := items[0].ID, "generic:refresh"; got != want {
		t.Fatalf("first generic item ID = %q, want %q", got, want)
	}
	if got, want := items[1].ID, "generic:switchTheme"; got != want {
		t.Fatalf("second generic item ID = %q, want %q", got, want)
	}

	withPlaylist := appendContextMenuGlobalItems(nil, true)
	if len(withPlaylist) == 0 || !withPlaylist[0].Header || withPlaylist[0].Label != iconTune+"播放控制" {
		t.Fatalf("non-empty playlist is missing playback controls: %#v", withPlaylist)
	}
}

func TestSongTitleBriefUsesCornerBrackets(t *testing.T) {
	if got, want := songTitleBrief("歌曲"), "「歌曲」"; got != want {
		t.Fatalf("songTitleBrief = %q, want %q", got, want)
	}
	if got, want := songTitleBrief("一二三四五六七八九十一二三四五六七八九十一"), "「一二三四五六七八九十一二三四五六七八九十…」"; got != want {
		t.Fatalf("truncated songTitleBrief = %q, want %q", got, want)
	}
}
