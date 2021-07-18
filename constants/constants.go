package constants

import (
	"time"
)

const AppName = "musicfox"
const AppVersion = "2.0.1"
const AppVersionInt = 20001
const AppDescription = "<cyan>Musicfox - 命令行版网易云音乐</>"
const AppShowStartup = true
const AppGithubUrl = "https://github.com/anhoder/go-musicfox"
const AppCheckUpdateUrl = "https://api.github.com/repos/anhoder/go-musicfox/releases/latest"
const ProgressFullChar = "#"
const ProgressEmptyChar = " "
const StartupProgressOutBounce = true
const StartupLoadingSeconds = 2
const StartupTickDuration = time.Millisecond * 16

const AppLocalDataDir = ".go-musicfox"
const AppDBName = "musicfox"
const AppIniFile = "go-musicfox.ini"

const PlayerSongBr = 320000 // 999000

const MainShowTitle = true
const MainLoadingText = " [加载中...]"

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
