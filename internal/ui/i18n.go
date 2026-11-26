package ui

import (
	"github.com/anhoder/foxful-cli/model"
)

// go-musicfox 应用级 MessageID。
const (
	MsgSearchPageTitle       model.MessageID = "search.page.title"
	MsgSearchPlaceholder     model.MessageID = "search.placeholder"
	MsgSearchKeywordRequired model.MessageID = "search.keyword_required"
	MsgSearchUnknownError    model.MessageID = "search.unknown_error"
	MsgSearchNetworkError    model.MessageID = "search.network_error"

	MsgLoginPageTitle           model.MessageID = "login.page.title"
	MsgLoginPageSubtitle        model.MessageID = "login.page.subtitle"
	MsgLoginAccountPlaceholder  model.MessageID = "login.account.placeholder"
	MsgLoginPasswordPlaceholder model.MessageID = "login.password.placeholder"
	MsgLoginCookiePlaceholder   model.MessageID = "login.cookie.placeholder"
	MsgLoginAccountTab          model.MessageID = "login.tab.account"
	MsgLoginCookieTab           model.MessageID = "login.tab.cookie"
	MsgLoginQRCodeButton        model.MessageID = "login.button.qrcode"
	MsgLoginQRCodeContinue      model.MessageID = "login.button.qrcode_continue"
	MsgLoginCredentialRequired  model.MessageID = "login.error.credential_required"
	MsgLoginFailed              model.MessageID = "login.error.failed"
	MsgLoginUnknownError        model.MessageID = "login.error.unknown"
	MsgLoginNetworkError        model.MessageID = "login.error.network"
	MsgLoginTooManyRequests     model.MessageID = "login.error.too_many_requests"
	MsgLoginInvalidCredentials  model.MessageID = "login.error.invalid_credentials"
	MsgLoginVersionTooOld       model.MessageID = "login.error.version_too_old"
	MsgLogin2FANotSupported     model.MessageID = "login.error.2fa_not_supported"
	MsgLoginCookieInvalid       model.MessageID = "login.error.cookie_invalid"
	MsgLoginCookieVerifying     model.MessageID = "login.cookie.verifying"
	MsgLoginCookieRequired      model.MessageID = "login.error.cookie_required"

	MsgLoginWebviewButton      model.MessageID = "login.button.webview"
	MsgLoginWebviewPageTitle   model.MessageID = "login.webview.page.title"
	MsgLoginWebviewWaiting     model.MessageID = "login.webview.waiting"
	MsgLoginWebviewVerifying   model.MessageID = "login.webview.verifying"
	MsgLoginWebviewSuccess     model.MessageID = "login.webview.success"
	MsgLoginWebviewCancelled   model.MessageID = "login.webview.cancelled"
	MsgLoginWebviewFailed      model.MessageID = "login.webview.failed"
	MsgLoginWebviewUnsupported model.MessageID = "login.webview.unsupported"
	MsgLoginWebviewUnavailable model.MessageID = "login.webview.unavailable"

	MsgQRLoginPageTitle       model.MessageID = "qr_login.page.title"
	MsgQRLoginGenerating      model.MessageID = "qr_login.generating"
	MsgQRLoginScanPrompt      model.MessageID = "qr_login.scan_prompt"
	MsgQRLoginSuccess         model.MessageID = "qr_login.success"
	MsgQRLoginExpired         model.MessageID = "qr_login.expired"
	MsgQRLoginExpiredAction   model.MessageID = "qr_login.expired_action"
	MsgQRLoginWaitingScan     model.MessageID = "qr_login.waiting_scan"
	MsgQRLoginWaitingConfirm  model.MessageID = "qr_login.waiting_confirm"
	MsgQRLoginUnknownStatus   model.MessageID = "qr_login.unknown_status"
	MsgQRLoginError           model.MessageID = "qr_login.error"
	MsgQRLoginOpenImageFailed model.MessageID = "qr_login.open_image_failed"
	MsgQRLoginGetCodeFailed   model.MessageID = "qr_login.get_code_failed"

	MsgOperationFailed           model.MessageID = "operation.failed"
	MsgOperationLogoutSuccess    model.MessageID = "operation.logout.success"
	MsgOperationLogoutCleared    model.MessageID = "operation.logout.cleared"
	MsgOperationLikeFailed       model.MessageID = "operation.like.failed"
	MsgOperationLikeAdded        model.MessageID = "operation.like.added"
	MsgOperationLikeRemoved      model.MessageID = "operation.like.removed"
	MsgOperationFMDisliked       model.MessageID = "operation.fm.disliked"
	MsgOperationDownloading      model.MessageID = "operation.download.downloading"
	MsgOperationDownloadSuccess  model.MessageID = "operation.download.success"
	MsgOperationDownloadExists   model.MessageID = "operation.download.exists"
	MsgOperationDownloadFailed   model.MessageID = "operation.download.failed"
	MsgOperationShareSuccess     model.MessageID = "operation.share.success"
	MsgOperationShareFailed      model.MessageID = "operation.share.failed"
	MsgOperationCacheCleared     model.MessageID = "operation.cache.cleared"
	MsgOperationCacheClearFailed model.MessageID = "operation.cache.clear_failed"

	MsgMenuCurrentPlaylist    model.MessageID = "menu.current_playlist"
	MsgMenuSimilarSongs       model.MessageID = "menu.similar_songs"
	MsgMenuMyPlaylists        model.MessageID = "menu.my_playlists"
	MsgMenuSearchResult       model.MessageID = "menu.search_result"
	MsgMenuOperationTarget    model.MessageID = "menu.operation.target"
	MsgMenuOperationNoPlaying model.MessageID = "menu.operation.no_playing"
	MsgMenuOperationUnknown   model.MessageID = "menu.operation.unknown"

	MsgErrorNoAlbum     model.MessageID = "error.no_album"
	MsgErrorNoArtist    model.MessageID = "error.no_artist"
	MsgPromptConfirm    model.MessageID = "prompt.confirm"
	MsgPromptClearCache model.MessageID = "prompt.clear_cache"
)

// zhMessages 覆盖 foxful-cli 内置组件（帮助提示、表格、文件选择器、表单、弹窗等）
// 使用的可翻译文案，使其与 go-musicfox 的中文界面保持一致。
var zhMessages = map[model.MessageID]string{
	model.MsgLoading:        "加载中...",
	model.MsgHintNavigate:   "导航",
	model.MsgHintConfirm:    "确认",
	model.MsgHintBack:       "返回",
	model.MsgHintQuit:       "退出",
	model.MsgHintSearch:     "搜索",
	model.MsgNoData:         "暂无数据",
	model.MsgNoColumns:      "暂无列",
	model.MsgEmptyDirectory: "(空目录)",
	model.MsgReadError:      "错误: %s",
	model.MsgYes:            "是",
	model.MsgNo:             "否",
	model.MsgConfirm:        "确认",
	model.MsgCancel:         "取消",
	model.MsgFieldRequired:  "此项为必填",

	MsgSearchPageTitle:       "搜索",
	MsgSearchPlaceholder:     " 输入关键词",
	MsgSearchKeywordRequired: "关键词不得为空",
	MsgSearchUnknownError:    "未知错误，请稍后再试~",
	MsgSearchNetworkError:    "网络异常，请稍后再试~",

	MsgLoginPageTitle:           "用户登录",
	MsgLoginPageSubtitle:        "手机号或邮箱",
	MsgLoginAccountPlaceholder:  "手机号或邮箱",
	MsgLoginPasswordPlaceholder: "密码",
	MsgLoginCookiePlaceholder:   "请输入 Cookie",
	MsgLoginAccountTab:          "手机号/邮箱登录",
	MsgLoginCookieTab:           "Cookie 登录",
	MsgLoginQRCodeButton:        "扫码登录",
	MsgLoginQRCodeContinue:      "已扫码登录，继续",
	MsgLoginCredentialRequired:  "请输入账号或密码",
	MsgLoginFailed:              "使用账号密码登录失败：",
	MsgLoginUnknownError:        "未知错误，请稍后再试！code: %d",
	MsgLoginNetworkError:        "网络异常，请检查后重试",
	MsgLoginTooManyRequests:     "请求过于频繁，请稍后再试~",
	MsgLoginInvalidCredentials:  "账号或密码错误，请重试",
	MsgLoginVersionTooOld:       "客户端版本过低，请升级到最新版本后重试",
	MsgLogin2FANotSupported:     "账号需要二阶段验证，目前暂不支持该功能",
	MsgLoginCookieInvalid:       "Cookie 格式错误: %w",
	MsgLoginCookieVerifying:     "正在验证 Cookie...",
	MsgLoginCookieRequired:      "请输入 Cookie",

	MsgLoginWebviewButton:      "网页登录",
	MsgLoginWebviewPageTitle:   "网页登录",
	MsgLoginWebviewWaiting:     "请在打开的窗口中完成登录...",
	MsgLoginWebviewVerifying:   "正在验证登录信息...",
	MsgLoginWebviewSuccess:     "登录成功！",
	MsgLoginWebviewCancelled:   "登录已取消",
	MsgLoginWebviewFailed:      "网页登录失败: ",
	MsgLoginWebviewUnsupported: "当前平台不支持网页登录",
	MsgLoginWebviewUnavailable: "网页登录不可用: ",

	MsgQRLoginPageTitle:       "二维码登录",
	MsgQRLoginGenerating:      "正在生成二维码，请稍候...",
	MsgQRLoginScanPrompt:      "请使用网易云音乐APP扫码",
	MsgQRLoginSuccess:         "登录成功！",
	MsgQRLoginExpired:         "二维码已失效",
	MsgQRLoginExpiredAction:   "二维码已失效，请按 'b' 或 'esc' 返回",
	MsgQRLoginWaitingScan:     "等待扫码...",
	MsgQRLoginWaitingConfirm:  "已扫码，请在手机上确认登录",
	MsgQRLoginUnknownStatus:   "未知状态: %d，请返回重试",
	MsgQRLoginError:           "发生错误: ",
	MsgQRLoginOpenImageFailed: "打开二维码失败: ",
	MsgQRLoginGetCodeFailed:   "无法获取二维码, code: %.0f",

	MsgOperationFailed:           "操作失败",
	MsgOperationLogoutSuccess:    "登出成功",
	MsgOperationLogoutCleared:    "已清理用户信息",
	MsgOperationLikeFailed:       "加入或移出歌单失败",
	MsgOperationLikeAdded:        "已添加到我喜欢的歌曲",
	MsgOperationLikeRemoved:      "已从我喜欢的歌曲移除",
	MsgOperationFMDisliked:       "已标记为不喜欢",
	MsgOperationDownloading:      "正在下载，请稍候...",
	MsgOperationDownloadSuccess:  "下载完成",
	MsgOperationDownloadExists:   "文件已存在",
	MsgOperationDownloadFailed:   "下载失败",
	MsgOperationShareSuccess:     "已复制到剪贴板",
	MsgOperationShareFailed:      "分享失败",
	MsgOperationCacheCleared:     "缓存已清除",
	MsgOperationCacheClearFailed: "清除缓存失败",

	MsgMenuCurrentPlaylist:    "当前播放列表",
	MsgMenuSimilarSongs:       "相似歌曲",
	MsgMenuMyPlaylists:        "我的歌单",
	MsgMenuSearchResult:       "搜索结果",
	MsgMenuOperationTarget:    "操作：",
	MsgMenuOperationNoPlaying: "当前无播放",
	MsgMenuOperationUnknown:   "未知操作对象",

	MsgErrorNoAlbum:     "歌曲没有专辑信息",
	MsgErrorNoArtist:    "歌曲没有歌手信息",
	MsgPromptConfirm:    "确定",
	MsgPromptClearCache: "清除缓存",
}

// SetupI18n 将中文文案注册到 foxful-cli 的默认 Catalog，并按 locale 选择界面语言。
// locale 为空时默认使用中文（zh），以匹配 go-musicfox 的中文界面。
func SetupI18n(locale string) {
	model.DefaultCatalog().Register("zh", zhMessages)
	if locale == "" {
		locale = "zh"
	}
	model.SetLocale(locale)
}
