package ui

import "time"

type tickerByPlayer struct {
	c chan time.Time
	p *Player
}

func newTickerByPlayer(p *Player) *tickerByPlayer {
	return &tickerByPlayer{
		c: make(chan time.Time),
		p: p,
	}
}

func (*tickerByPlayer) Start() error {
	return nil
}

func (t *tickerByPlayer) Ticker() <-chan time.Time {
	return t.c
}

func (t *tickerByPlayer) PassedTime() time.Duration {
	return t.p.PassedTime()
}

func (tickerByPlayer) Close() error {
	return nil
}
