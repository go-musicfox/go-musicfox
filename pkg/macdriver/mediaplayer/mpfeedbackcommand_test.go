//go:build darwin

package mediaplayer

import (
	"testing"

	"github.com/go-musicfox/go-musicfox/pkg/macdriver/core"
)

func TestMPFeedbackCommand(t *testing.T) {
	center := MPRemoteCommandCenter_sharedCommandCenter()
	cmd := center.LikeCommand()
	if cmd.ID == 0 {
		panic("get like command failed")
	}

	cmd.SetActive(true)
	if cmd.IsActive() != true {
		panic("active is wrong")
	}

	cmd.SetLocalizedTitle(core.String("aaaa"))
	title := cmd.LocalizedTitle()
	if title.ID == 0 || title.String() != "aaaa" {
		panic("title is wrong")
	}

	cmd.SetLocalizedShortTitle(core.String("bbbb"))
	shortTitle := cmd.LocalizedShortTitle()
	if shortTitle.ID == 0 || shortTitle.String() != "bbbb" {
		panic("short title is wrong")
	}
}
