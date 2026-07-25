package netease

import "testing"

func TestParseCloudLyricResponse(t *testing.T) {
	const want = "[00:00.00]作词 : 刘畊宏\n[00:01.00]作曲 : 周杰伦"

	got, err := parseCloudLyricResponse(200, []byte(`{"code":200,"lrc":"[00:00.00]作词 : 刘畊宏\n[00:01.00]作曲 : 周杰伦"}`))
	if err != nil {
		t.Fatalf("parse cloud lyric response: %v", err)
	}
	if got.Original != want {
		t.Fatalf("original lyric = %q, want %q", got.Original, want)
	}
}

func TestParseCloudLyricResponseRejectsMissingLyrics(t *testing.T) {
	if _, err := parseCloudLyricResponse(200, []byte(`{"code":200}`)); err == nil {
		t.Fatal("expected missing cloud lyric to return an error")
	}
}

func TestParseCloudLyricResponseRejectsEmptyLyrics(t *testing.T) {
	if _, err := parseCloudLyricResponse(200, []byte(`{"code":200,"lrc":""}`)); err == nil {
		t.Fatal("expected empty cloud lyric to return an error")
	}
}

func TestParseCloudLyricResponseRejectsFailedRequest(t *testing.T) {
	if _, err := parseCloudLyricResponse(401, []byte(`{"code":401}`)); err == nil {
		t.Fatal("expected failed cloud lyric request to return an error")
	}
}
