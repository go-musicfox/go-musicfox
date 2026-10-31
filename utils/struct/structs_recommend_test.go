package _struct //nolint:staticcheck // Match the existing package name.

import (
	"reflect"
	"testing"

	"github.com/go-musicfox/go-musicfox/internal/structs"
)

func TestGetDailyPlaylistsReadsMobileRecommendationBlock(t *testing.T) {
	response := []byte(`{"code":200,"data":{"blocks":[
		{"positionCode":"PAGE_RECOMMEND_OTHER_PLAYLIST","dslData":{"blockResource":{"resources":[
			{"resourceType":"playList","resourceId":"999","title":"Unrelated playlist"}
		]}}},
		{"positionCode":"PAGE_RECOMMEND_SPECIAL_CLOUD_VILLAGE_PLAYLIST","dslData":{"blockResource":{"resources":[
			{"resourceType":"playList","resourceId":"101","title":"First"},
			{"resourceType":"playList","resourceId":202,"title":"Second"},
			{"resourceType":"playList","resourceId":"101","title":"Duplicate"},
			{"resourceType":"song","resourceId":"303","title":"A song"}
		]}}}
	]}}`)
	want := []structs.Playlist{{Id: 101, Name: "First"}, {Id: 202, Name: "Second"}}
	if got := GetDailyPlaylists(response); !reflect.DeepEqual(got, want) {
		t.Fatalf("GetDailyPlaylists() = %+v, want %+v", got, want)
	}
}

func TestGetDailyPlaylistsRejectsNonMobileAndMalformedResponses(t *testing.T) {
	for _, response := range []string{
		`{"code":200,"recommend":[{"id":1,"name":"Legacy endpoint"}]}`,
		`{"code":200,"data":{"blockData":{"creatives":[{"creativeType":"list","action":"orpheus://playlist/101","uiElement":{"mainTitle":{"title":"PC endpoint"}},"resources":[{"resourceType":"list","resourceId":"101","uiElement":{"mainTitle":{"title":"PC endpoint"}}}]}]}}}`,
		`{"code":200,"data":{"blocks":[{"positionCode":"PAGE_RECOMMEND_OTHER_PLAYLIST","dslData":{"blockResource":{"resources":[{"resourceType":"playList","resourceId":"101","title":"Unrelated"}]}}}]}}`,
		`{"code":200,"data":{"blocks":[{"positionCode":"PAGE_RECOMMEND_SPECIAL_CLOUD_VILLAGE_PLAYLIST","dslData":{"blockResource":{"resources":[]}}}]}}`,
		`{"code":200,"data":{"blocks":[{"positionCode":"PAGE_RECOMMEND_SPECIAL_CLOUD_VILLAGE_PLAYLIST","dslData":{"blockResource":{"resources":[{"resourceType":"playList","resourceId":"101","title":"Truncated"}]}}}]`,
	} {
		if got := GetDailyPlaylists([]byte(response)); len(got) != 0 {
			t.Errorf("GetDailyPlaylists(%s) = %+v, want no playlists", response, got)
		}
	}
}

func TestGetDailyPlaylistsRequiresDirectValidResourceFields(t *testing.T) {
	response := []byte(`{"code":200,"data":{"blocks":[
		{"positionCode":"PAGE_RECOMMEND_SPECIAL_CLOUD_VILLAGE_PLAYLIST","dslData":{"blockResource":{"resources":[
			{"resourceType":"playList","resourceId":"invalid","title":"Invalid ID"},
			{"resourceType":"playList","resourceId":0,"title":"Zero ID"},
			{"resourceType":"playList","resourceId":-1,"title":"Negative ID"},
			{"resourceType":"playList","resourceId":true,"title":"Wrong ID type"},
			{"resourceType":"playList","resourceId":"101","title":" "},
			{"resourceType":"playList","resourceId":"202","title":{"text":"Not a plain title"}},
			{"resourceType":"playList","title":"No direct ID","action":"orpheus://playlist/303","logInfo":{"resourceId":"303"}},
			{"resourceType":"playList","resourceId":"404","uiElement":{"mainTitle":{"title":"No direct title"}}}
		]}}}
	]}}`)
	if got := GetDailyPlaylists(response); len(got) != 0 {
		t.Fatalf("GetDailyPlaylists() = %+v, want no playlists", got)
	}
}
