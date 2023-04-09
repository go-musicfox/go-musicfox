package utils

import (
	"sync"

	"github.com/mattn/go-runewidth"
)

type XScrollBar struct {
	_increment int

	lastContent string
	l           sync.Mutex
}

func NewXScrollBar() *XScrollBar {
	return &XScrollBar{
		_increment: 0,
	}
}

func (b *XScrollBar) increase(reset bool) int {
	b.l.Lock()
	defer b.l.Unlock()
	b._increment++
	if b._increment > 1000 || reset {
		b._increment = 1
	}
	return b._increment
}

func (b *XScrollBar) Tick(width int, content string) string {
	var i = b.increase(false)
	if b.lastContent != content {
		i = b.increase(true)
		b.lastContent = content
	}

	var tmp string
	length := runewidth.StringWidth(content)

	// 歌词首末补偿，歌词开头结尾等待3*200ms
	a := i%(length+3) - 3
	if length < width || a < 1 {
		tmp = runewidth.TruncateLeft(b.lastContent, 0, "")
	} else if a+width <= length {
		tmp = runewidth.TruncateLeft(b.lastContent, a, "")
	} else {
		tmp = runewidth.TruncateLeft(b.lastContent, length-width, "")
	}
	return runewidth.Truncate(runewidth.FillRight(tmp, width), width, "")
}
