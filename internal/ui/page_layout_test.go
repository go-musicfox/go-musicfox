package ui

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/anhoder/foxful-cli/model"
	"github.com/charmbracelet/x/ansi"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/structs"
)

func TestStandalonePageMenuTitleMatchesMainPageRow(t *testing.T) {
	const (
		width           = 80
		height          = 24
		mainTitle       = "MAIN_PAGE_TITLE"
		standaloneTitle = "STANDALONE_PAGE_TITLE"
	)

	previousConfig := configs.AppConfig
	configs.AppConfig = &configs.Config{}
	configs.AppConfig.Theme.ShowTitle = true
	t.Cleanup(func() { configs.AppConfig = previousConfig })

	opts := model.DefaultOptions()
	opts.EnableStartup = false
	opts.WhetherDisplayTitle = true
	opts.MainMenu = &pageLayoutTestMenu{}
	opts.MainMenuTitle = &model.MenuItem{Title: mainTitle}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	opts.TeaOptions = []tea.ProgramOption{
		tea.WithContext(ctx),
		tea.WithInput(nil),
		tea.WithOutput(io.Discard),
	}

	app := model.NewApp(opts)
	_ = app.Run()
	if app.Main() == nil {
		t.Fatal("main page was not initialized")
	}
	_, _ = app.Update(tea.WindowSizeMsg{Width: width, Height: height})

	wantRow := visibleRowContaining(t, app.Main().View(app), mainTitle, height)
	netease := &Netease{App: app}
	pages := map[string]model.Page{
		"search": NewSearchPage(netease),
		"login":  NewLoginPage(netease),
	}

	for name, page := range pages {
		t.Run(name, func(t *testing.T) {
			switch page := page.(type) {
			case *SearchPage:
				page.menuTitle = &model.MenuItem{Title: standaloneTitle}
			case *LoginPage:
				page.menuTitle = &model.MenuItem{Title: standaloneTitle}
			}

			view := page.View(app)
			gotHeight := lipgloss.Height(view)
			if gotHeight > height {
				t.Errorf("rendered height = %d, terminal height = %d", gotHeight, height)
			}
			if gotRow := visibleRowContaining(t, view, standaloneTitle, height); gotRow != wantRow {
				t.Errorf("visible menu title row = %d, main page row = %d", gotRow, wantRow)
			}
		})
	}
}

func TestTopStatusBarKeepsProgressBarOnLastRow(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		height int
	}{
		{name: "compact", width: 80, height: 24},
		{name: "runtime size", width: 120, height: 30},
		{name: "tall", width: 80, height: 40},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if row := renderTopStatusProgressBarRow(t, tt.width, tt.height); row != tt.height-1 {
				t.Fatalf("progress bar row = %d, want last terminal row %d", row, tt.height-1)
			}
		})
	}
}

func renderTopStatusProgressBarRow(t *testing.T, width, height int) int {
	t.Helper()

	previousConfig := configs.AppConfig
	configs.AppConfig = &configs.Config{}
	t.Cleanup(func() { configs.AppConfig = previousConfig })

	opts := model.DefaultOptions()
	opts.EnableStartup = false
	opts.WhetherDisplayTitle = true
	opts.StatusBar = &model.DefaultStatusBar{}
	opts.StatusBarPosition = model.StatusBarTop
	opts.DualColumn = true
	opts.MainMenu = &pageLayoutTestMenu{itemCount: 16}
	opts.MainMenuTitle = &model.MenuItem{Title: "Main"}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	opts.TeaOptions = []tea.ProgramOption{
		tea.WithContext(ctx),
		tea.WithInput(nil),
		tea.WithOutput(io.Discard),
	}

	netease := &Netease{}
	state := songInfoTestState{song: structs.Song{Id: 1, Name: "Layout Song", Duration: time.Minute}}
	opts.Components = []model.Component{
		&LyricRenderer{netease: netease},
		NewSongInfoRenderer(netease, state),
		NewProgressRenderer(netease, state),
	}
	app := model.NewApp(opts)
	netease.App = app
	_ = app.Run()
	_, _ = app.Update(tea.WindowSizeMsg{Width: width, Height: height})
	view := app.Main().View(app)
	return visibleRowContaining(t, view, "00:00/01:00", height)
}

func visibleRowContaining(t *testing.T, view, marker string, height int) int {
	t.Helper()
	lines := strings.Split(ansi.Strip(view), "\n")
	firstVisible := max(0, len(lines)-height)
	for row, line := range lines[firstVisible:] {
		if strings.Contains(line, marker) {
			return row
		}
	}
	t.Fatalf("visible view does not contain %q", marker)
	return -1
}

type pageLayoutTestMenu struct {
	model.DefaultMenu
	itemCount int
}

func (pageLayoutTestMenu) GetMenuKey() string {
	return "page-layout-test"
}

func (pageLayoutTestMenu) HelpHints() []model.HelpHint {
	return nil
}

func (m pageLayoutTestMenu) MenuViews() []model.MenuItem {
	count := max(1, m.itemCount)
	items := make([]model.MenuItem, count)
	for i := range items {
		items[i] = model.MenuItem{Title: "item"}
	}
	return items
}
