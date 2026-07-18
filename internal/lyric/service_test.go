package lyric

import (
	"testing"
	"time"
)

func TestStateIncludesLyricOffset(t *testing.T) {
	service := NewService(nil, false, 250*time.Millisecond, false)
	if got, want := service.State().OffsetMs, int64(250); got != want {
		t.Errorf("initial offset = %dms, want %dms", got, want)
	}

	service.SetOffset(-100 * time.Millisecond)
	if got, want := service.State().OffsetMs, int64(-100); got != want {
		t.Errorf("updated offset = %dms, want %dms", got, want)
	}
}
