package util

import (
	"sync"

	"github.com/cnsilvan/UnblockNeteaseMusic/common"
	"github.com/cnsilvan/UnblockNeteaseMusic/config"
	"github.com/cnsilvan/UnblockNeteaseMusic/network"
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
	ProxyUrl           string

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
	config.ProxyUrl = &ProxyUrl
	network.GlobalProxyUrl = ProxyUrl

	provider.Init()
}
