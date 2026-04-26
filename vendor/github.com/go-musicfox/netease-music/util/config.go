package util

import (
	"sync"
	"time"

	"github.com/cnsilvan/UnblockNeteaseMusic/common"
	"github.com/cnsilvan/UnblockNeteaseMusic/config"
	"github.com/cnsilvan/UnblockNeteaseMusic/provider"
)

var (
	UNMSwitch          bool
	Sources            []string
	ForceBestQuality   bool
	SearchLimit        int
	EnableLocalVip     bool
	UnlockSoundEffects bool
	QQCookieFile       string
	UNMProxyURL        string
	HTTPClientTimeout  time.Duration

	providerInited = sync.Once{}
)

func ConfigInit() {
	providerInited.Do(func() {
		ConfigReload()
	})
}

func ConfigReload() {
	common.Source = Sources
	config.ForceBestQuality = &ForceBestQuality
	config.SearchLimit = &SearchLimit
	config.EnableLocalVip = &EnableLocalVip
	config.UnlockSoundEffects = &UnlockSoundEffects
	config.QQCookieFile = &QQCookieFile

	if len(UNMProxyURL) > 0 {
		UNMSwitch = false
	}

	provider.Init()
}
