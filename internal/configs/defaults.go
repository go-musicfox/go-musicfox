package configs

import (
	"runtime"

	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/netease-music/service"
)

// NewDefaultConfig 创建并返回一个包含所有默认值的Config实例。
// 这是所有配置加载流程的起点和最低优先级的回退。
func NewDefaultConfig() *Config {
	// 基础的主题和UI默认值
	theme := ThemeConfig{
		ShowTitle:        true,
		LoadingText:      types.MainLoadingText,
		DoubleColumn:     true,
		DynamicMenuRows:  false,
		CenterEverything: false,
		PrimaryColor:     types.AppPrimaryColor,
		Progress: ProgressConfig{
			FullChar:           types.ProgressFullChar,
			EmptyChar:          types.ProgressEmptyChar,
			FullCharWhenFirst:  types.ProgressFullChar,
			FullCharWhenLast:   types.ProgressFullChar,
			LastFullChar:       types.ProgressFullChar,
			EmptyCharWhenFirst: types.ProgressEmptyChar,
			EmptyCharWhenLast:  types.ProgressEmptyChar,
			FirstEmptyChar:     types.ProgressEmptyChar,
		},
	}

	// 播放器引擎的平台特定默认值
	defaultPlayerEngine := types.BeepPlayer
	switch runtime.GOOS {
	case "darwin":
		defaultPlayerEngine = types.OsxPlayer
	case "windows":
		defaultPlayerEngine = types.WinMediaPlayer
	}

	return &Config{
		Startup: StartupConfig{
			Enable:            true,
			ProgressOutBounce: true,
			LoadingSeconds:    types.StartupLoadingSeconds,
			Welcome:           types.AppName,
			SignIn:            false,
			CheckUpdate:       true,
		},
		Main: MainConfig{
			AltScreen:        true,
			EnableMouseEvent: true,
			Debug:            false,
			Notification: NotificationConfig{
				Enable:     true,
				Icon:       types.DefaultNotifyIcon,
				AlbumCover: false,
			},
			Lyric: LyricConfig{
				Show:            true,
				ShowTranslation: true,
				Offset:          0,
			},
			Pprof: PprofConfig{
				Port: types.MainPProfPort,
			},
			Account: AccountConfig{
				NeteaseCookie: "",
			},
		},
		Theme: theme,
		Storage: StorageConfig{
			DownloadDir:           "",
			LyricDir:              "",
			DownloadSongWithLyric: false,
			FileNameTpl:           "{{.SongName}}-{{.SongArtists}}.{{.FileExt}}",
			Cache: CacheConfig{
				Dir:   "",
				Limit: 0,
			},
		},
		Player: PlayerConfig{
			Engine:                 defaultPlayerEngine,
			MaxPlayErrCount:        types.MaxPlayErrCount,
			SongLevel:              service.Higher,
			ShowAllSongsOfPlaylist: false,
			Beep: BeepConfig{
				Mp3Decoder: types.BeepGoMp3Decoder,
			},
			Mpd: MpdConfig{
				Bin:        "mpd",
				ConfigFile: "",
				Network:    "unix",
				Addr:       "",
				AutoStart:  true,
			},
			Mpv: MpvConfig{
				Bin: "mpv",
			},
		},
		Autoplay: AutoplayConfig{
			Enable:   false,
			Playlist: AutoPlayerPlaylistDailyReco,
			Offset:   0,
			Mode:     types.PmUnknown,
		},
		UNM: UNMConfig{
			Enable:             false,
			Sources:            []string{types.UNMDefaultSources},
			SearchLimit:        0,
			EnableLocalVip:     true,
			UnlockSoundEffects: true,
			QQCookieFile:       "",
			SkipInvalidTracks:  false,
		},
		Reporter: ReporterConfig{
			Netease: NeteaseReporterConfig{
				Enable: false,
			},
			Lastfm: LastfmReporterConfig{
				Enable:          false,
				Key:             "",
				Secret:          "",
				ScrobblePoint:   50,
				OnlyFirstArtist: false,
				SkipDjRadio:     false,
			},
		},
		Keybindings: KeybindingsConfig{
			UseDefaultKeyBindings: true,
			Global:                make(map[string]string),
			App:                   make(map[string][]string),
		},
		Share: make(map[string]string),
	}
}
