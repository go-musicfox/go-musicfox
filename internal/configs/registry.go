package configs

import (
	"runtime"
	"strings"
	"time"

	"github.com/anhoder/foxful-cli/model"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-musicfox/netease-music/service"
	"github.com/gookit/ini/v2"

	"github.com/go-musicfox/go-musicfox/internal/types"
)

var ConfigRegistry *Registry

type Registry struct {
	Startup       StartupOptions
	Progress      ProgressOptions
	Main          MainOptions
	AutoPlayer    AutoPlayerOptions
	UNM           UNMOptions
	Player        PlayerOptions
	GlobalHotkeys map[string]string
}

func (r *Registry) FillToModelOpts(opts *model.Options) {
	opts.StartupOptions = r.Startup.StartupOptions
	opts.ProgressOptions = r.Progress.ProgressOptions

	opts.AppName = types.AppName
	opts.WhetherDisplayTitle = r.Main.ShowTitle
	opts.LoadingText = r.Main.LoadingText
	opts.PrimaryColor = r.Main.PrimaryColor
	opts.DualColumn = r.Main.DualColumn

	if r.Main.EnableMouseEvent {
		opts.TeaOptions = append(opts.TeaOptions, tea.WithMouseCellMotion())
	}
	if r.Main.AltScreen {
		opts.TeaOptions = append(opts.TeaOptions, tea.WithAltScreen())
	}
}

func NewRegistryWithDefault() *Registry {
	registry := &Registry{
		Startup: StartupOptions{
			StartupOptions: model.StartupOptions{
				EnableStartup:     true,
				ProgressOutBounce: true,
				TickDuration:      types.StartupTickDuration,
				LoadingDuration:   time.Second * types.StartupLoadingSeconds,
				Welcome:           types.AppName,
			},
			SignIn:      true,
			CheckUpdate: true,
		},
		Progress: ProgressOptions{
			ProgressOptions: model.ProgressOptions{
				EmptyChar:          []rune(types.ProgressEmptyChar)[0],
				EmptyCharWhenFirst: []rune(types.ProgressEmptyChar)[0],
				EmptyCharWhenLast:  []rune(types.ProgressEmptyChar)[0],
				FirstEmptyChar:     []rune(types.ProgressEmptyChar)[0],
				FullChar:           []rune(types.ProgressFullChar)[0],
				FullCharWhenFirst:  []rune(types.ProgressFullChar)[0],
				FullCharWhenLast:   []rune(types.ProgressFullChar)[0],
				LastFullChar:       []rune(types.ProgressFullChar)[0],
			},
		},
		Main: MainOptions{
			ShowTitle:        true,
			LoadingText:      types.MainLoadingText,
			PlayerSongLevel:  service.Higher,
			PrimaryColor:     types.AppPrimaryColor,
			ShowLyric:        true,
			ShowLyricTrans:   true,
			ShowNotify:       true,
			NotifyIcon:       types.DefaultNotifyIcon,
			NotifyAlbumCover: false,
			PProfPort:        types.MainPProfPort,
			AltScreen:        true,
			EnableMouseEvent: true,
			DownloadDir:      "",
			CacheLimit:       0,
			DynamicMenuRows:  false,
		},
		Player: PlayerOptions{
			Engine:         types.BeepPlayer,
			BeepMp3Decoder: types.BeepGoMp3Decoder,
		},
		UNM: UNMOptions{
			Enable:             true,
			Sources:            []string{types.UNMDefaultSources},
			EnableLocalVip:     true,
			UnlockSoundEffects: true,
		},
	}

	switch runtime.GOOS {
	case "darwin":
		registry.Player.Engine = types.OsxPlayer
	case "windows":
		registry.Player.Engine = types.WinMediaPlayer
	}

	return registry
}

func NewRegistryFromIniFile(filepath string) *Registry {
	registry := NewRegistryWithDefault()

	if err := ini.LoadExists(filepath); err != nil {
		return registry
	}

	registry.Startup.EnableStartup = ini.Bool("startup.show", true)
	registry.Startup.ProgressOutBounce = ini.Bool("startup.progressOutBounce", true)
	registry.Startup.LoadingDuration = time.Second * time.Duration(ini.Int("startup.loadingSeconds", types.StartupLoadingSeconds))
	registry.Startup.Welcome = ini.String("startup.welcome", types.AppName)
	registry.Startup.SignIn = ini.Bool("startup.signIn", false)
	registry.Startup.CheckUpdate = ini.Bool("startup.checkUpdate", true)

	emptyChar := ini.String("progress.emptyChar", types.ProgressEmptyChar)
	registry.Progress.EmptyChar = firstCharOrDefault(emptyChar, types.ProgressEmptyChar)
	emptyCharWhenFirst := ini.String("progress.emptyCharWhenFirst", types.ProgressEmptyChar)
	registry.Progress.EmptyCharWhenFirst = firstCharOrDefault(emptyCharWhenFirst, types.ProgressEmptyChar)
	emptyCharWhenLast := ini.String("progress.emptyCharWhenLast", types.ProgressEmptyChar)
	registry.Progress.EmptyCharWhenLast = firstCharOrDefault(emptyCharWhenLast, types.ProgressEmptyChar)
	firstEmptyChar := ini.String("progress.firstEmptyChar", types.ProgressEmptyChar)
	registry.Progress.FirstEmptyChar = firstCharOrDefault(firstEmptyChar, types.ProgressEmptyChar)

	registry.GlobalHotkeys = ini.StringMap("global_hotkey")

	fullChar := ini.String("progress.fullChar", types.ProgressFullChar)
	registry.Progress.FullChar = firstCharOrDefault(fullChar, types.ProgressFullChar)
	fullCharWhenFirst := ini.String("progress.fullCharWhenFirst", types.ProgressFullChar)
	registry.Progress.FullCharWhenFirst = firstCharOrDefault(fullCharWhenFirst, types.ProgressFullChar)
	fullCharWhenLast := ini.String("progress.fullCharWhenLast", types.ProgressFullChar)
	registry.Progress.FullCharWhenLast = firstCharOrDefault(fullCharWhenLast, types.ProgressFullChar)
	lastFullChar := ini.String("progress.lastFullChar", types.ProgressEmptyChar)
	registry.Progress.LastFullChar = firstCharOrDefault(lastFullChar, types.ProgressEmptyChar)

	registry.Main.ShowTitle = ini.Bool("main.showTitle", true)
	registry.Main.LoadingText = ini.String("main.loadingText", types.MainLoadingText)
	songLevel := service.SongQualityLevel(ini.String("main.songLevel", string(service.Higher)))
	if songLevel.IsValid() {
		registry.Main.PlayerSongLevel = songLevel
	}
	primaryColor := ini.String("main.primaryColor", types.AppPrimaryColor)
	if primaryColor != "" {
		registry.Main.PrimaryColor = primaryColor
	} else {
		registry.Main.PrimaryColor = types.AppPrimaryColor
	}
	registry.Main.ShowLyric = ini.Bool("main.showLyric", true)
	registry.Main.LyricOffset = ini.Int("main.lyricOffset", 0)
	registry.Main.ShowLyricTrans = ini.Bool("main.showLyricTrans", true)
	registry.Main.ShowNotify = ini.Bool("main.showNotify", true)
	registry.Main.NotifyIcon = ini.String("main.notifyIcon", types.DefaultNotifyIcon)
	registry.Main.NotifyAlbumCover = ini.Bool("main.notifyAlbumCover", false)
	registry.Main.PProfPort = ini.Int("main.pprofPort", types.MainPProfPort)
	registry.Main.AltScreen = ini.Bool("main.altScreen", true)
	registry.Main.EnableMouseEvent = ini.Bool("main.enableMouseEvent", true)
	registry.Main.DualColumn = ini.Bool("main.doubleColumn", true)
	registry.Main.DownloadDir = ini.String("main.downloadDir", "")
	registry.Main.DownloadLyricDir = ini.String("main.downloadLyricDir", "")
	registry.Main.DownloadFileNameTpl = ini.String("main.downloadFileNameTpl", "")
	registry.Main.ShowAllSongsOfPlaylist = ini.Bool("main.showAllSongsOfPlaylist", false)
	registry.Main.CacheDir = ini.String("main.cacheDir", "")
	registry.Main.CacheLimit = ini.Int64("main.cacheLimit", 0)
	registry.Main.DynamicMenuRows = ini.Bool("main.dynamicMenuRows", false)

	defaultPlayer := types.BeepPlayer
	switch runtime.GOOS {
	case "darwin":
		defaultPlayer = types.OsxPlayer
	case "windows":
		defaultPlayer = types.WinMediaPlayer
	}
	registry.Player.Engine = ini.String("player.engine", defaultPlayer)
	registry.Player.BeepMp3Decoder = ini.String("player.beepMp3Decoder", types.BeepGoMp3Decoder)
	registry.Player.MpdBin = ini.String("player.mpdBin", "")
	registry.Player.MpdConfigFile = ini.String("player.mpdConfigFile", "")
	registry.Player.MpdNetwork = ini.String("player.mpdNetwork", "")
	registry.Player.MpdAddr = ini.String("player.mpdAddr", "")
	registry.Player.MpdAutoStart = ini.Bool("player.mpdAutoStart", true)
	registry.Player.MaxPlayErrCount = ini.Int("player.maxPlayErrCount", types.MaxPlayErrCount)

	// Auto play
	registry.AutoPlayer.Enable = ini.Bool("autoplay.autoPlay", false)
	registry.AutoPlayer.Playlist = AutoPlayerPlaylistFromString(ini.String("autoplay.autoPlayList", string(AutoPlayerPlaylistNo)))
	registry.AutoPlayer.Offset = ini.Int("autoplay.offset", 0)
	registry.AutoPlayer.Mode = PlayerModeFromAutoPlayModeString(ini.String("autoplay.playMode"))

	// UNM
	registry.UNM.Enable = ini.Bool("unm.switch", false)

	sourceStr := ini.String("unm.sources", "kuwo")
	if sourceStr != "" {
		var sources []string
		for _, source := range strings.Split(sourceStr, ",") {
			sources = append(sources, strings.TrimSpace(source))
		}
		registry.UNM.Sources = sources
	}

	registry.UNM.SearchLimit = ini.Int("unm.searchLimit", 0)
	registry.UNM.EnableLocalVip = ini.Bool("unm.enableLocalVip", true)
	registry.UNM.UnlockSoundEffects = ini.Bool("unm.unlockSoundEffects", true)
	registry.UNM.QQCookieFile = ini.String("unm.qqCookieFile", "")

	return registry
}

func firstCharOrDefault(s, defaultStr string) rune {
	if len(s) > 0 {
		return []rune(s)[0]
	}
	return []rune(defaultStr)[0]
}
