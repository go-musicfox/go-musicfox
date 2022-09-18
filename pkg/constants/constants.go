package constants

import (
	"time"
)

var (
	// AppVersion Inject by -ldflags
	AppVersion = "v2.2.1"
)

const AppName = "musicfox"
const AppDescription = "<cyan>Musicfox - 命令行版网易云音乐</>"
const AppShowStartup = true
const AppGithubUrl = "https://github.com/anhoder/go-musicfox"
const AppLatestReleases = "https://github.com/anhoder/go-musicfox/releases/latest"
const AppCheckUpdateUrl = "https://api.github.com/repos/anhoder/go-musicfox/releases/latest"
const ProgressFullChar = "#"
const ProgressEmptyChar = " "
const StartupProgressOutBounce = true
const StartupLoadingSeconds = 2
const StartupTickDuration = time.Millisecond * 16
const StartupSignIn = true
const StartupCheckUpdate = true

const AppLocalDataDir = ".go-musicfox"
const AppDBName = "musicfox"
const AppIniFile = "go-musicfox.ini"
const AppPrimaryRandom = "random"
const AppPrimaryColor = "#f90022"

const PlayerSongBr = 320000 // 999000

const MainShowTitle = true
const MainLoadingText = "[加载中...]"
const MainShowLyric = true
const MainShowNotify = true
const MainNotifySender = "com.netease.163music"
const MainPProfPort = 9876
const MainAltScreen = true

const PlayerEngine = "beep" // beep、mpd

const AppHelpTemplate = `%s

{{.Description}} (Version: <info>{{.Version}}</>)

<comment>Usage:</>
  {$binName} [Global Options...] <info>{command}</> [--option ...] [argument ...]

<comment>Global Options:</>
{{.GOpts}}
<comment>Available Commands:</>{{range $module, $cs := .Cs}}{{if $module}}
<comment> {{ $module }}</>{{end}}{{ range $cs }}
  <info>{{.Name | paddingName }}</> {{.UseFor}}{{if .Aliases}} (alias: <cyan>{{ join .Aliases ","}}</>){{end}}{{end}}{{end}}

  <info>{{ paddingName "help" }}</> Display help information

Use "<cyan>{$binName} {COMMAND} -h</>" for more information about a command
`
