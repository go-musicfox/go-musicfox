module go-musicfox

go 1.16

require (
	github.com/anhoder/bubbles v0.7.8 // indirect
	github.com/anhoder/bubbletea v0.12.8
	github.com/anhoder/netease-music v1.0.0
	github.com/buger/jsonparser v1.1.1 // indirect
	github.com/faiface/beep v1.0.3-0.20210301102329-98afada94bff // indirect
	github.com/fogleman/ease v0.0.0-20170301025033-8da417bf1776
	github.com/gomodule/redigo v1.8.4 // indirect
	github.com/google/gops v0.3.17 // indirect
	github.com/gookit/gcli/v2 v2.3.4
	github.com/lucasb-eyer/go-colorful v1.2.0
	github.com/mattn/go-runewidth v0.0.10
	github.com/muesli/termenv v0.7.4
	github.com/telanflow/cookiejar v0.0.0-20190719062046-114449e86aa5
	golang.org/x/sys v0.0.0-20210414055047-fe65e336abe0 // indirect
)

replace (
	github.com/anhoder/bubbletea => ../bubbletea
	github.com/anhoder/netease-music => ../netease-music
)
