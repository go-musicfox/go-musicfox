package commands

import (
    tea "github.com/anhoder/bubbletea"
    "github.com/gookit/gcli/v2"
    "go-musicfox/configs"
    "go-musicfox/pkg/ui"
)

func NewPlayerCommand() *gcli.Command {
    return &gcli.Command{
        Name: "netease",
        UseFor: "Command line player for Netease Cloud Music",
        Func:   runPlayer,
    }
}

func runPlayer(_ *gcli.Command, _ []string) error {
    neteaseModel := ui.NewNeteaseModel(configs.ConfigRegistry.StartupLoadingDuration)
    program := tea.NewProgram(neteaseModel)
    neteaseModel.BindProgram(program)
    program.EnterAltScreen()
    defer program.ExitAltScreen()
    return program.Start()
}
