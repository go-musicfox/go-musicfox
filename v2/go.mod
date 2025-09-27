module github.com/go-musicfox/go-musicfox/v2

go 1.23.0

require (
	github.com/bytecodealliance/wasmtime-go/v14 v14.0.0
	github.com/ebitengine/purego v0.8.4
	github.com/fsnotify/fsnotify v1.9.0
	github.com/go-musicfox/go-musicfox/v2/pkg/event v0.0.0
	github.com/go-musicfox/go-musicfox/v2/pkg/kernel v0.0.0
	github.com/go-musicfox/go-musicfox/v2/pkg/model v0.0.0
	github.com/go-musicfox/go-musicfox/v2/pkg/plugin/core v0.0.0
	github.com/google/uuid v1.6.0
	github.com/knadh/koanf/parsers/json v1.0.0
	github.com/knadh/koanf/parsers/toml v0.1.0
	github.com/knadh/koanf/parsers/yaml v1.1.0
	github.com/knadh/koanf/providers/env v1.1.0
	github.com/knadh/koanf/providers/file v1.2.0
	github.com/knadh/koanf/providers/posflag v1.0.1
	github.com/knadh/koanf/v2 v2.1.1
	github.com/mattn/go-sqlite3 v1.14.32
	github.com/spf13/pflag v1.0.10
	github.com/stretchr/testify v1.11.1
	gopkg.in/yaml.v3 v3.0.1
)

replace github.com/charmbracelet/bubbletea v0.25.0 => github.com/go-musicfox/bubbletea v0.25.0-foxful

replace github.com/go-musicfox/go-musicfox/v2/pkg/event => ./pkg/event

replace github.com/go-musicfox/go-musicfox/v2/pkg/kernel => ./pkg/kernel

replace github.com/go-musicfox/go-musicfox/v2/pkg/model => ./pkg/model

replace github.com/go-musicfox/go-musicfox/v2/pkg/plugin/core => ./pkg/plugin/core

replace github.com/go-musicfox/go-musicfox/v2/plugins/storage => ./plugins/storage

require (
	github.com/anhoder/foxful-cli v0.5.0 // indirect
	github.com/atotto/clipboard v0.1.4 // indirect
	github.com/aymanbagabas/go-osc52/v2 v2.0.1 // indirect
	github.com/buger/jsonparser v1.1.1 // indirect
	github.com/charmbracelet/bubbles v0.16.1 // indirect
	github.com/charmbracelet/bubbletea v0.25.0 // indirect
	github.com/charmbracelet/lipgloss v0.8.0 // indirect
	github.com/cnsilvan/UnblockNeteaseMusic v0.0.0-20230310083816-92b59c95a366 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/erikgeiser/coninput v0.0.0-20211004153227-1c3628e74d0f // indirect
	github.com/faiface/beep v1.1.0 // indirect
	github.com/fogleman/ease v0.0.0-20170301025033-8da417bf1776 // indirect
	github.com/forgoer/openssl v1.6.0 // indirect
	github.com/go-musicfox/netease-music v1.4.9 // indirect
	github.com/go-musicfox/requests v0.2.3 // indirect
	github.com/go-viper/mapstructure/v2 v2.0.0-alpha.1 // indirect
	github.com/gomodule/redigo v1.9.2 // indirect
	github.com/hajimehoshi/go-mp3 v0.3.0 // indirect
	github.com/hajimehoshi/oto v0.7.1 // indirect
	github.com/icza/bitio v1.0.0 // indirect
	github.com/jfreymuth/oggvorbis v1.0.1 // indirect
	github.com/jfreymuth/vorbis v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/knadh/koanf/maps v0.1.2 // indirect
	github.com/lucasb-eyer/go-colorful v1.2.0 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/mattn/go-localereader v0.0.1 // indirect
	github.com/mattn/go-runewidth v0.0.15 // indirect
	github.com/mewkiz/flac v1.0.7 // indirect
	github.com/mewkiz/pkg v0.0.0-20190919212034-518ade7978e2 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180228061459-e0a39a4cb421 // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/muesli/ansi v0.0.0-20230316100256-276c6243b2f6 // indirect
	github.com/muesli/cancelreader v0.2.2 // indirect
	github.com/muesli/reflow v0.3.0 // indirect
	github.com/muesli/termenv v0.15.2 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rivo/uniseg v0.4.6 // indirect
	github.com/robotn/gohook v0.41.0 // indirect
	github.com/sahilm/fuzzy v0.1.0 // indirect
	github.com/skip2/go-qrcode v0.0.0-20200617195104-da1b6568686e // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/telanflow/cookiejar v0.0.0-20190719062046-114449e86aa5 // indirect
	github.com/vcaesar/keycode v0.10.1 // indirect
	go.uber.org/dig v1.17.1 // indirect
	go.yaml.in/yaml/v3 v3.0.3 // indirect
	golang.org/x/exp v0.0.0-20190306152737-a1d7652674e8 // indirect
	golang.org/x/image v0.0.0-20190227222117-0694c2d4d067 // indirect
	golang.org/x/mobile v0.0.0-20190415191353-3e0bab5405d6 // indirect
	golang.org/x/sync v0.6.0 // indirect
	golang.org/x/sys v0.32.0 // indirect
	golang.org/x/term v0.17.0 // indirect
	golang.org/x/text v0.13.0 // indirect
)
