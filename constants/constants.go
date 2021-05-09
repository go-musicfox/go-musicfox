package constants

import (
	"time"
)

const AppName = "musicfox"
const AppVersion = "2.0.0"
const AppVersionInt = 20000
const AppDescription = "<cyan>Musicfox - 命令行版网易云音乐</>"
const AppShowStartup = true
const ProgressFullChar = '#'
const ProgressEmptyChar = ' '
const StartupProgressOutBounce = true
const StartupLoadingDuration = time.Second * 2
const StartupTickDuration = time.Millisecond * 16

const AppLocalDataDir = ".go-musicfox"
const AppDBName = "musicfox"

const PlayerSongBr = "320000" // "999000"

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
