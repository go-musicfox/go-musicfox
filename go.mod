module go-musicfox

go 1.16

require (
	github.com/anhoder/bubbles v0.7.8
	github.com/anhoder/bubbletea v0.12.10
	github.com/anhoder/netease-music v1.1.1
	github.com/anhoder/notificator v0.0.0-20220906123738-8410351970b5
	github.com/asmcos/requests v0.0.0-20210319030608-c839e8ae4946 // indirect
	github.com/atotto/clipboard v0.1.4 // indirect
	github.com/boltdb/bolt v1.3.1
	github.com/buger/jsonparser v1.1.1
	github.com/containerd/console v1.0.3 // indirect
	github.com/faiface/beep v1.1.0
	github.com/fogleman/ease v0.0.0-20170301025033-8da417bf1776
	github.com/forgoer/openssl v1.2.1 // indirect
	github.com/gomodule/redigo v1.8.8 // indirect
	github.com/gookit/gcli/v2 v2.3.4
	github.com/gookit/goutil v0.5.2 // indirect
	github.com/gookit/ini/v2 v2.1.0
	github.com/hajimehoshi/oto v1.0.1 // indirect
	github.com/icza/bitio v1.1.0 // indirect
	github.com/jfreymuth/oggvorbis v1.0.3 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/lucasb-eyer/go-colorful v1.2.0
	github.com/mattn/go-runewidth v0.0.13
	github.com/mewkiz/pkg v0.0.0-20211102230744-16a6ce8f1b77 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/muesli/termenv v0.11.0
	github.com/progrium/macdriver v0.2.0
	github.com/telanflow/cookiejar v0.0.0-20190719062046-114449e86aa5
	github.com/tosone/minimp3 v1.0.1
)

replace (
	github.com/faiface/beep v1.1.0 => github.com/anhoder/beep v1.1.1
	github.com/progrium/macdriver v0.2.0 => github.com/anhoder/macdriver v0.2.4
	github.com/tosone/minimp3 v1.0.1 => github.com/anhoder/minimp3 v1.0.2
)
