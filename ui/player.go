package ui

import (
    "github.com/anhoder/go-musicfox/player"
    "strings"
    "time"
)

type Player struct {
    model         *neteaseModel

    playlist      []player.Music
    curMusic      player.Music

    lyric         map[float64]string
    showLyric     bool
    lyricQueue    *[]string
    lyricStartRow int
    lyricLines    int

    timer         *time.Timer
    startPlayTime time.Time
}

func NewPlayer(model *neteaseModel) *Player {
    return &Player{
        model: model,
        timer: time.NewTimer(time.Second),
    }
}

func (p *Player) playerView() string {
    var playerBuilder strings.Builder

    playerBuilder.WriteString(p.LyricView())

    return playerBuilder.String()

}

func (p *Player) LyricView() string {
    startRow := p.model.menuBottomRow + 1
    endRow := p.model.WindowHeight - 4

    if len(p.lyric) == 0 || !p.showLyric {
        return strings.Repeat("\n", endRow - startRow)
    }

    var lyricBuilder strings.Builder
    lyricBuilder.WriteString(strings.Repeat("\n", p.lyricStartRow - p.model.menuBottomRow))

    return lyricBuilder.String()
}

func (p *Player) SingView() string {
    return ""
}

func (p *Player) ProgressView() string {
    return ""
}