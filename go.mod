module github.com/go-musicfox/go-musicfox

go 1.22

require (
	github.com/anhoder/foxful-cli v0.4.2
	github.com/bogem/id3v2/v2 v2.1.4
	github.com/buger/jsonparser v1.1.1
	github.com/charmbracelet/bubbles v0.16.1
	github.com/charmbracelet/bubbletea v0.25.0
	github.com/charmbracelet/lipgloss v0.8.0
	github.com/ebitengine/purego v0.7.0
	github.com/fhs/gompd/v2 v2.3.0
	github.com/frolovo22/tag v0.0.2
	github.com/gen2brain/beeep v0.0.0-20240516210008-9c006672e7f4
	github.com/go-flac/flacpicture v0.3.0
	github.com/go-musicfox/netease-music v1.4.6
	github.com/go-musicfox/notificator v0.1.2
	github.com/go-ole/go-ole v1.3.0
	github.com/godbus/dbus/v5 v5.1.0
	github.com/gookit/gcli/v2 v2.3.4
	github.com/gookit/ini/v2 v2.2.2
	github.com/gopxl/beep v1.4.0
	github.com/markthree/go-get-folder-size v0.5.0
	github.com/mattn/go-runewidth v0.0.15
	github.com/muesli/termenv v0.15.2
	github.com/pkg/errors v0.9.1
	github.com/robotn/gohook v0.41.0
	github.com/saltosystems/winrt-go v0.0.0-20240320184339-289d313a74b7
	github.com/shkh/lastfm-go v0.0.0-20191215035245-89a801c244e0
	github.com/skip2/go-qrcode v0.0.0-20200617195104-da1b6568686e
	github.com/skratchdot/open-golang v0.0.0-20200116055534-eef842397966
	github.com/telanflow/cookiejar v0.0.0-20190719062046-114449e86aa5
	github.com/tosone/minimp3 v1.0.2
	go.etcd.io/bbolt v1.3.7
	golang.org/x/sys v0.30.0
)

require (
	github.com/atotto/clipboard v0.1.4 // indirect
	github.com/aymanbagabas/go-osc52/v2 v2.0.1 // indirect
	github.com/cnsilvan/UnblockNeteaseMusic v0.0.0-20230310083816-92b59c95a366 // indirect
	github.com/cocoonlife/goflac v0.0.0-20170210142907-50ea06ed5a9d // indirect
	github.com/ebitengine/oto/v3 v3.1.0 // indirect
	github.com/erikgeiser/coninput v0.0.0-20211004153227-1c3628e74d0f // indirect
	github.com/fogleman/ease v0.0.0-20170301025033-8da417bf1776 // indirect
	github.com/forgoer/openssl v1.6.0 // indirect
	github.com/go-flac/go-flac v1.0.0 // indirect
	github.com/go-musicfox/requests v0.2.3 // indirect
	github.com/go-toast/toast v0.0.0-20190211030409-01e6764cf0a4 // indirect
	github.com/gomodule/redigo v1.8.9 // indirect
	github.com/gookit/color v1.5.3 // indirect
	github.com/gookit/goutil v0.6.10 // indirect
	github.com/hajimehoshi/go-mp3 v0.3.4 // indirect
	github.com/icza/bitio v1.1.0 // indirect
	github.com/jfreymuth/oggvorbis v1.0.5 // indirect
	github.com/jfreymuth/vorbis v1.0.2 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/lucasb-eyer/go-colorful v1.2.0 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/mattn/go-localereader v0.0.1 // indirect
	github.com/mewkiz/flac v1.0.8 // indirect
	github.com/mewkiz/pkg v0.0.0-20230226050401-4010bf0fec14 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/muesli/ansi v0.0.0-20230316100256-276c6243b2f6 // indirect
	github.com/muesli/cancelreader v0.2.2 // indirect
	github.com/muesli/reflow v0.3.0 // indirect
	github.com/nu7hatch/gouuid v0.0.0-20131221200532-179d4d0c4d8d // indirect
	github.com/rivo/uniseg v0.4.6 // indirect
	github.com/sahilm/fuzzy v0.1.0 // indirect
	github.com/stretchr/testify v1.9.0 // indirect
	github.com/tadvi/systray v0.0.0-20190226123456-11a2b8fa57af // indirect
	github.com/tidwall/gjson v1.17.1 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/vcaesar/keycode v0.10.1 // indirect
	github.com/xo/terminfo v0.0.0-20220910002029-abceb7e1c41e // indirect
	golang.org/x/crypto v0.31.0 // indirect
	golang.org/x/exp v0.0.0-20230905200255-921286631fa9 // indirect
	golang.org/x/sync v0.10.0 // indirect
	golang.org/x/term v0.27.0 // indirect
	golang.org/x/text v0.21.0 // indirect
)

replace (
	github.com/charmbracelet/bubbletea v0.25.0 => github.com/go-musicfox/bubbletea v0.25.0-foxful
	github.com/cnsilvan/UnblockNeteaseMusic => github.com/go-musicfox/UnblockNeteaseMusic v0.1.6
	// github.com/cnsilvan/UnblockNeteaseMusic => ../UnblockNeteaseMusic
	github.com/cocoonlife/goflac v0.0.0-20170210142907-50ea06ed5a9d => github.com/go-musicfox/goflac v0.1.5
	github.com/frolovo22/tag v0.0.2 => github.com/go-musicfox/tag v1.0.2
	// github.com/go-musicfox/netease-music => ../netease-music
	// github.com/go-musicfox/requests => ../requests
	github.com/gookit/gcli/v2 v2.3.4 => github.com/anhoder/gcli/v2 v2.3.5
	github.com/gopxl/beep v1.4.0 => github.com/go-musicfox/beep v1.4.1
	github.com/hajimehoshi/go-mp3 v0.3.4 => github.com/go-musicfox/go-mp3 v0.3.3
	github.com/robotn/gohook v0.41.0 => github.com/go-musicfox/gohook v0.41.1
	github.com/saltosystems/winrt-go => github.com/go-musicfox/winrt-go v0.1.4
	github.com/shkh/lastfm-go => github.com/go-musicfox/lastfm-go v0.0.2
)
