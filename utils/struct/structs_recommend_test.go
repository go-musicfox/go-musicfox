package _struct //nolint:staticcheck // Match the existing package name.

import (
	"reflect"
	"testing"

	"github.com/go-musicfox/go-musicfox/internal/structs"
)

func TestGetDailyPlaylistsReadsPCRecommendations(t *testing.T) {
	response := []byte(`{"code":200,"data":{"blockData":{"creatives":[
		{"creativeType":"list","uiElement":{"mainTitle":{"title":"First"}},"resources":[
			{"resourceType":"list","resourceId":"101","uiElement":{"mainTitle":{"title":"First"}}}
		]},
		{"creativeType":"scroll_playlist","resources":[
			{"resourceType":"list","resourceId":202,"uiElement":{"mainTitle":{"title":"Second"}}},
			{"resourceType":"list","resourceId":"101","uiElement":{"mainTitle":{"title":"Duplicate"}}},
			{"resourceType":"song","resourceId":"303","uiElement":{"mainTitle":{"title":"A song"}}}
		]},
		{"creativeType":"list","action":"orpheus://playlist/999","uiElement":{"mainTitle":{"title":"Unused fallback"}},"resources":[
			{"resourceType":"list","resourceId":"101","uiElement":{"mainTitle":{"title":"Duplicate"}}}
		]}
	]}}}`)
	want := []structs.Playlist{{Id: 101, Name: "First"}, {Id: 202, Name: "Second"}}
	if got := GetDailyPlaylists(response); !reflect.DeepEqual(got, want) {
		t.Fatalf("GetDailyPlaylists() = %+v, want %+v", got, want)
	}
}

func TestGetDailyPlaylistsFallsBackToPlaylistActionWithoutUsableResources(t *testing.T) {
	response := []byte(`{"code":200,"data":{"blockData":{"creatives":[
		{"creativeType":"list","action":"orpheus://playlist/101?autoplay=0","uiElement":{"mainTitle":{"title":"First"}}},
		{"creativeType":"scroll_playlist","resources":[],"action":"orpheus://playlist/202","uiElement":{"mainTitle":{"title":"Second"}}},
		{"creativeType":"list","resources":[{"resourceExtInfo":{"playCount":18302}}],"action":"orpheus://playlist/3015788101","uiElement":{"mainTitle":{"title":"Metadata only"}}},
		{"creativeType":"list","action":"orpheus://song/303","uiElement":{"mainTitle":{"title":"A song"}}},
		{"creativeType":"song","action":"orpheus://playlist/404","uiElement":{"mainTitle":{"title":"Not a playlist card"}}}
	]}}}`)
	want := []structs.Playlist{{Id: 101, Name: "First"}, {Id: 202, Name: "Second"}, {Id: 3015788101, Name: "Metadata only"}}
	if got := GetDailyPlaylists(response); !reflect.DeepEqual(got, want) {
		t.Fatalf("GetDailyPlaylists() = %+v, want %+v", got, want)
	}
}

func TestGetDailyPlaylistsRejectsInvalidOrEmptyRecommendations(t *testing.T) {
	for _, response := range []string{
		`{"code":200,"recommend":[{"id":1,"name":"Old endpoint"}]}`,
		`{"code":200,"data":{"blockData":{"creatives":[]}}}`,
		`{"code":200,"data":{"blockData":{"creatives":[{"creativeType":"list","resources":[
			{"resourceType":"list","resourceId":"invalid","uiElement":{"mainTitle":{"title":"Invalid ID"}}},
			{"resourceType":"list","resourceId":0,"uiElement":{"mainTitle":{"title":"Zero ID"}}},
			{"resourceType":"list","resourceId":1,"uiElement":{"mainTitle":{"title":" "}}}
		]}]}}}`,
		`{"code":200,"data":{"blockData":{"creatives":[{"creativeType":"list","resources":[{"resourceType":"list","resourceId":1,"uiElement":{"mainTitle":{"title":"Truncated"}}}]}]`,
	} {
		if got := GetDailyPlaylists([]byte(response)); len(got) != 0 {
			t.Errorf("GetDailyPlaylists(%s) = %+v, want no playlists", response, got)
		}
	}
}
