package core

import (
	"log/slog"

	neteaseutil "github.com/go-musicfox/netease-music/util"
	cookiejar "github.com/juju/persistent-cookiejar"

	"github.com/go-musicfox/go-musicfox/utils/slogx"
)

// CompleteQRLogin 完成扫码登录的后半程：替换 app 级 cookie jar、同步全局
// jar、持久化并刷新用户资料。jar 为 nil 时跳过 jar 操作（nil-safe），
// 仍执行 LoginCallback。
func (e *Engine) CompleteQRLogin(jar *cookiejar.Jar) error {
	if jar != nil {
		SetAppCookieJar(jar)
		neteaseutil.SetGlobalCookieJar(jar)
		if err := jar.Save(); err != nil {
			slog.Warn("持久化 Cookie 失败", slogx.Error(err))
		}
	}
	return e.LoginCallback()
}
