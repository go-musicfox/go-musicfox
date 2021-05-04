module go-musicfox

go 1.16

require (
	github.com/StackExchange/wmi v0.0.0-20210224194228-fe8f1750fd46 // indirect
	github.com/anhoder/bubbles v0.7.8
	github.com/anhoder/bubbletea v0.12.8
	github.com/anhoder/netease-music v1.0.0
	github.com/boltdb/bolt v1.3.1
	github.com/buger/jsonparser v1.1.1
	github.com/faiface/beep v1.0.3-0.20210301102329-98afada94bff
	github.com/fogleman/ease v0.0.0-20170301025033-8da417bf1776
	github.com/go-ole/go-ole v1.2.5 // indirect
	github.com/gomodule/redigo v1.8.4 // indirect
	github.com/google/gops v0.3.18 // indirect
	github.com/gookit/gcli/v2 v2.3.4
	github.com/lucasb-eyer/go-colorful v1.2.0
	github.com/mattn/go-runewidth v0.0.10
	github.com/muesli/termenv v0.7.4
	github.com/shirou/gopsutil/v3 v3.21.4 // indirect
	github.com/telanflow/cookiejar v0.0.0-20190719062046-114449e86aa5
	github.com/tklauser/go-sysconf v0.3.5 // indirect
	github.com/xlab/treeprint v1.1.0 // indirect
	golang.org/x/sys v0.0.0-20210503173754-0981d6026fa6 // indirect
)

replace (
	github.com/anhoder/bubbletea => ../bubbletea
	github.com/anhoder/netease-music => ../netease-music
)
