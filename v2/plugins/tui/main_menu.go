package tui

import (
	"github.com/anhoder/foxful-cli/model"
)

// MainMenu 主菜单实现，基于foxful-cli
type MainMenu struct {
	model.DefaultMenu
	plugin   *TUIPlugin
	menus    []model.MenuItem
	menuList []model.Menu
}

// NewMainMenu 创建主菜单
func NewMainMenu(plugin *TUIPlugin) *MainMenu {
	mainMenu := &MainMenu{
		plugin: plugin,
		menus: []model.MenuItem{
			{Title: "每日推荐歌曲"},
			{Title: "每日推荐歌单"},
			{Title: "我的歌单"},
			{Title: "我的收藏"},
			{Title: "私人FM"},
			{Title: "专辑列表"},
			{Title: "搜索"},
			{Title: "排行榜"},
			{Title: "精选歌单"},
			{Title: "热门歌手"},
			{Title: "最近播放歌曲"},
			{Title: "云盘"},
			{Title: "主播电台"},
			{Title: "LastFM"},
			{Title: "帮助"},
			{Title: "检查更新"},
		},
		menuList: []model.Menu{
			// 这里应该创建具体的子菜单实例
			// 为了简化，暂时使用nil占位
			nil, // 每日推荐歌曲
			nil, // 每日推荐歌单
			nil, // 我的歌单
			nil, // 我的收藏
			nil, // 私人FM
			nil, // 专辑列表
			NewSearchMenu(plugin), // 搜索
			nil, // 排行榜
			nil, // 精选歌单
			nil, // 热门歌手
			nil, // 最近播放歌曲
			nil, // 云盘
			nil, // 主播电台
			nil, // LastFM
			NewHelpMenu(plugin), // 帮助
			nil, // 检查更新
		},
	}
	return mainMenu
}

// FormatMenuItem 格式化菜单项
func (m *MainMenu) FormatMenuItem(item *model.MenuItem) {
	subtitle := "[未登录]"
	// 从插件获取用户信息
	if m.plugin.user != nil && m.plugin.user.Nickname != "" {
		subtitle = "[" + m.plugin.user.Nickname + "]"
	}
	item.Subtitle = subtitle
}

// GetMenuKey 获取菜单键
func (m *MainMenu) GetMenuKey() string {
	return "main_menu"
}

// MenuViews 获取菜单视图
func (m *MainMenu) MenuViews() []model.MenuItem {
	for i := range m.menus {
		m.FormatMenuItem(&m.menus[i])
	}
	return m.menus
}

// SubMenu 获取子菜单
func (m *MainMenu) SubMenu(_ *model.App, index int) model.Menu {
	if index >= len(m.menuList) {
		return nil
	}
	return m.menuList[index]
}

// IsSearchable 是否可搜索
func (m *MainMenu) IsSearchable() bool {
	return true
}

// RealDataIndex 真实数据索引
func (m *MainMenu) RealDataIndex(index int) int {
	return index
}

// SearchMenu 搜索菜单实现
type SearchMenu struct {
	model.DefaultMenu
	plugin *TUIPlugin
}

// NewSearchMenu 创建搜索菜单
func NewSearchMenu(plugin *TUIPlugin) *SearchMenu {
	return &SearchMenu{
		plugin: plugin,
	}
}

// GetMenuKey 获取菜单键
func (s *SearchMenu) GetMenuKey() string {
	return "search_menu"
}

// MenuViews 获取菜单视图
func (s *SearchMenu) MenuViews() []model.MenuItem {
	return []model.MenuItem{
		{Title: "搜索歌曲"},
		{Title: "搜索专辑"},
		{Title: "搜索歌手"},
		{Title: "搜索歌单"},
		{Title: "搜索用户"},
		{Title: "搜索电台"},
	}
}

// SubMenu 获取子菜单
func (s *SearchMenu) SubMenu(_ *model.App, index int) model.Menu {
	// 根据索引返回对应的搜索结果菜单
	switch index {
	case 0: // 搜索歌曲
		return NewSearchResultMenu(s.plugin, "song", "歌曲")
	case 1: // 搜索专辑
		return NewSearchResultMenu(s.plugin, "album", "专辑")
	case 2: // 搜索歌手
		return NewSearchResultMenu(s.plugin, "artist", "歌手")
	case 3: // 搜索歌单
		return NewSearchResultMenu(s.plugin, "playlist", "歌单")
	case 4: // 搜索用户
		return NewSearchResultMenu(s.plugin, "user", "用户")
	case 5: // 搜索电台
		return NewSearchResultMenu(s.plugin, "radio", "电台")
	default:
		return nil
	}
}

// HelpMenu 帮助菜单实现
type HelpMenu struct {
	model.DefaultMenu
	plugin *TUIPlugin
}

// NewHelpMenu 创建帮助菜单
func NewHelpMenu(plugin *TUIPlugin) *HelpMenu {
	return &HelpMenu{
		plugin: plugin,
	}
}

// GetMenuKey 获取菜单键
func (h *HelpMenu) GetMenuKey() string {
	return "help_menu"
}

// MenuViews 获取菜单视图
func (h *HelpMenu) MenuViews() []model.MenuItem {
	return []model.MenuItem{
		{Title: "快捷键说明", Subtitle: "查看所有快捷键"},
		{Title: "使用指南", Subtitle: "如何使用go-musicfox"},
		{Title: "常见问题", Subtitle: "FAQ"},
		{Title: "关于", Subtitle: "关于go-musicfox"},
	}
}

// SubMenu 获取子菜单
func (h *HelpMenu) SubMenu(_ *model.App, index int) model.Menu {
	// 根据索引返回对应的帮助内容菜单
	switch index {
	case 0: // 快捷键说明
		return NewHelpContentMenu(h.plugin, "shortcuts", "快捷键说明")
	case 1: // 使用指南
		return NewHelpContentMenu(h.plugin, "guide", "使用指南")
	case 2: // 常见问题
		return NewHelpContentMenu(h.plugin, "faq", "常见问题")
	case 3: // 关于
		return NewHelpContentMenu(h.plugin, "about", "关于")
	default:
		return nil
	}
}

// SearchResultMenu 搜索结果菜单
type SearchResultMenu struct {
	model.DefaultMenu
	plugin     *TUIPlugin
	searchType string
	typeName   string
	results    []model.MenuItem
}

// NewSearchResultMenu 创建搜索结果菜单
func NewSearchResultMenu(plugin *TUIPlugin, searchType, typeName string) *SearchResultMenu {
	return &SearchResultMenu{
		plugin:     plugin,
		searchType: searchType,
		typeName:   typeName,
		results:    []model.MenuItem{},
	}
}

// GetMenuKey 获取菜单键
func (s *SearchResultMenu) GetMenuKey() string {
	return "search_result_" + s.searchType
}

// MenuViews 获取菜单视图
func (s *SearchResultMenu) MenuViews() []model.MenuItem {
	if len(s.results) == 0 {
		return []model.MenuItem{
			{Title: "暂无搜索结果", Subtitle: "请输入关键词进行搜索"},
		}
	}
	return s.results
}

// SubMenu 获取子菜单
func (s *SearchResultMenu) SubMenu(_ *model.App, index int) model.Menu {
	// 搜索结果通常不需要子菜单
	return nil
}

// SetResults 设置搜索结果
func (s *SearchResultMenu) SetResults(results []model.MenuItem) {
	s.results = results
}

// HelpContentMenu 帮助内容菜单
type HelpContentMenu struct {
	model.DefaultMenu
	plugin      *TUIPlugin
	contentType string
	title       string
	content     []model.MenuItem
}

// NewHelpContentMenu 创建帮助内容菜单
func NewHelpContentMenu(plugin *TUIPlugin, contentType, title string) *HelpContentMenu {
	menu := &HelpContentMenu{
		plugin:      plugin,
		contentType: contentType,
		title:       title,
	}
	menu.loadContent()
	return menu
}

// GetMenuKey 获取菜单键
func (h *HelpContentMenu) GetMenuKey() string {
	return "help_" + h.contentType
}

// MenuViews 获取菜单视图
func (h *HelpContentMenu) MenuViews() []model.MenuItem {
	return h.content
}

// SubMenu 获取子菜单
func (h *HelpContentMenu) SubMenu(_ *model.App, index int) model.Menu {
	return nil
}

// loadContent 加载帮助内容
func (h *HelpContentMenu) loadContent() {
	switch h.contentType {
	case "shortcuts":
		h.content = []model.MenuItem{
			{Title: "基本操作", Subtitle: "方向键: 导航, 回车: 选择, ESC: 返回"},
			{Title: "播放控制", Subtitle: "空格: 播放/暂停, []: 上/下一首, +/-: 音量"},
			{Title: "搜索功能", Subtitle: "/: 搜索, n/N: 下/上一个结果"},
			{Title: "其他功能", Subtitle: "l: 歌词, s: 收藏, d: 下载, i: 信息"},
		}
	case "guide":
		h.content = []model.MenuItem{
			{Title: "1. 登录账号", Subtitle: "首次使用需要登录网易云音乐账号"},
			{Title: "2. 浏览音乐", Subtitle: "通过主菜单浏览推荐、歌单等内容"},
			{Title: "3. 搜索音乐", Subtitle: "使用搜索功能查找喜欢的音乐"},
			{Title: "4. 播放控制", Subtitle: "使用快捷键控制音乐播放"},
		}
	case "faq":
		h.content = []model.MenuItem{
			{Title: "Q: 如何登录?", Subtitle: "A: 在主界面选择登录选项，扫码或输入账号密码"},
			{Title: "Q: 无法播放音乐?", Subtitle: "A: 检查网络连接和音频设备设置"},
			{Title: "Q: 快捷键不生效?", Subtitle: "A: 确保焦点在应用窗口内"},
			{Title: "Q: 如何下载音乐?", Subtitle: "A: 选择歌曲后按'd'键（需要会员）"},
		}
	case "about":
		h.content = []model.MenuItem{
			{Title: "go-musicfox v2", Subtitle: "基于Go语言的网易云音乐命令行客户端"},
			{Title: "开源项目", Subtitle: "GitHub: github.com/go-musicfox/go-musicfox"},
			{Title: "技术栈", Subtitle: "Go + bubbletea + foxful-cli"},
			{Title: "许可证", Subtitle: "MIT License"},
		}
	default:
		h.content = []model.MenuItem{
			{Title: "暂无内容", Subtitle: "该帮助内容暂未实现"},
		}
	}
}