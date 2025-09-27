module github.com/go-musicfox/go-musicfox/v2/cmd/musicfox-core

go 1.23.0

require (
	github.com/go-musicfox/go-musicfox/v2/pkg/event v0.0.0
	github.com/go-musicfox/go-musicfox/v2/pkg/kernel v0.0.0
	github.com/go-musicfox/go-musicfox/v2/pkg/model v0.0.0
	github.com/go-musicfox/go-musicfox/v2/pkg/plugin/core v0.0.0
	github.com/go-musicfox/go-musicfox/v2/plugins/audio v0.0.0-00010101000000-000000000000
	github.com/go-musicfox/go-musicfox/v2/plugins/playlist v0.0.0-00010101000000-000000000000
	github.com/google/uuid v1.6.0
	github.com/spf13/pflag v1.0.10
	github.com/stretchr/testify v1.11.1
)

replace github.com/go-musicfox/go-musicfox/v2 => ../../

replace github.com/go-musicfox/go-musicfox/v2/plugins/audio => ../../plugins/audio

replace github.com/go-musicfox/go-musicfox/v2/plugins/playlist => ../../plugins/playlist

replace github.com/go-musicfox/go-musicfox/v2/pkg/event => ../../pkg/event

replace github.com/go-musicfox/go-musicfox/v2/pkg/kernel => ../../pkg/kernel

replace github.com/go-musicfox/go-musicfox/v2/pkg/model => ../../pkg/model

replace github.com/go-musicfox/go-musicfox/v2/pkg/plugin/core => ../../pkg/plugin/core

require (
	github.com/bytecodealliance/wasmtime-go/v14 v14.0.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/ebitengine/purego v0.8.4 // indirect
	github.com/faiface/beep v1.1.0 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/go-musicfox/go-musicfox/v2 v2.0.0 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.0.0-alpha.1 // indirect
	github.com/hajimehoshi/go-mp3 v0.3.0 // indirect
	github.com/hajimehoshi/oto v0.7.1 // indirect
	github.com/icza/bitio v1.0.0 // indirect
	github.com/jfreymuth/oggvorbis v1.0.1 // indirect
	github.com/jfreymuth/vorbis v1.0.0 // indirect
	github.com/knadh/koanf/maps v0.1.2 // indirect
	github.com/knadh/koanf/parsers/json v1.0.0 // indirect
	github.com/knadh/koanf/parsers/toml v0.1.0 // indirect
	github.com/knadh/koanf/parsers/yaml v1.1.0 // indirect
	github.com/knadh/koanf/providers/env v1.1.0 // indirect
	github.com/knadh/koanf/providers/file v1.2.0 // indirect
	github.com/knadh/koanf/v2 v2.1.1 // indirect
	github.com/mewkiz/flac v1.0.7 // indirect
	github.com/mewkiz/pkg v0.0.0-20190919212034-518ade7978e2 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	go.uber.org/dig v1.17.1 // indirect
	go.yaml.in/yaml/v3 v3.0.3 // indirect
	golang.org/x/exp v0.0.0-20190306152737-a1d7652674e8 // indirect
	golang.org/x/image v0.0.0-20190227222117-0694c2d4d067 // indirect
	golang.org/x/mobile v0.0.0-20190415191353-3e0bab5405d6 // indirect
	golang.org/x/sys v0.32.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
