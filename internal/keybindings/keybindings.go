// internal/keybindings/keybindings.go
package keybindings

import (
	"strings"
)

type OperateType int

func (op OperateType) String() string {
	return op.Name()
}

func (op OperateType) Name() string {
	if info, ok := keyBindingsRegistry[op]; ok {
		return info.name
	}
	return ""
}

func (op OperateType) Desc() string {
	if info, ok := keyBindingsRegistry[op]; ok {
		return info.desc
	}
	return ""
}

func (op OperateType) Keys() []string {
	if keys, ok := UserOperateToKeys()[op]; ok {
		return keys
	}
	return []string{}
}

type OperationInfo struct {
	name string // 操作的唯一标识符
	desc string // 用户友好的描述
}

// Operate (Managed by foxful-cli)
const (
	OpRerenderUI OperateType = -(iota + 1)
	OpMoveLeft
	OpMoveRight
	OpMoveUp
	OpMoveDown
	OpMoveToTop
	OpMoveToBottom
	OpEnter
	OpGoBack
	OpSearch
	OpQuit
)

// Operate (Safe to Customize)
const (
	OpHelp OperateType = iota
	OpPageUp
	OpPageDown
	OpPlayOrToggle
	OpToggle
	OpPrevious
	OpNext
	OpSeekBackward5s
	OpSeekBackward1s
	OpSeekForward5s
	OpSeekForward10s
	OpVolumeDown
	OpVolumeUp
	OpSwitchPlayMode
	OpIntelligence
	OpClearSongCache
	OpLogout
	OpCurPlaylist
	OpAppendSongsToNext
	OpAppendSongsToEnd
	OpDeleteSongFromPlaylist
	OpLikePlayingSong
	OpDislikePlayingSong
	OpTrashPlayingSong
	OpAddPlayingToUserPlaylist
	OpRemovePlayingFromUserPlaylist
	OpDownloadPlayingSong
	OpDownloadPlayingSongLrc
	OpAlbumOfPlayingSong
	OpArtistOfPlayingSong
	OpOpenPlayingSongInWeb
	OpLikeSelectedSong
	OpDislikeSelectedSong
	OpTrashSelectedSong
	OpAddSelectedToUserPlaylist
	OpRemoveSelectedFromUserPlaylist
	OpDownloadSelectedSong
	OpAlbumOfSelectedSong
	OpArtistOfSelectedSong
	OpOpenSelectedItemInWeb
	OpCollectSelectedPlaylist
	OpDiscollectSelectedPlaylist
)

// 操作信息
var keyBindingsRegistry = map[OperateType]OperationInfo{
	OpRerenderUI:   {name: "rerender", desc: "重新渲染UI"},
	OpMoveLeft:     {name: "moveLeft", desc: "左"},
	OpMoveRight:    {name: "moveRight", desc: "右"},
	OpMoveUp:       {name: "moveUp", desc: "上"},
	OpMoveDown:     {name: "moveDown", desc: "下"},
	OpMoveToTop:    {name: "moveToTop", desc: "上移到顶部"},
	OpMoveToBottom: {name: "moveToBottom", desc: "下移到底部"},
	OpEnter:        {name: "enter", desc: "进入"},
	OpGoBack:       {name: "goBack", desc: "返回上一级菜单"},
	OpSearch:       {name: "search", desc: "搜索当前列表"},
	OpQuit:         {name: "quit", desc: "退出"},

	OpHelp:                           {name: "help", desc: "帮助信息"},
	OpPageUp:                         {name: "pageUp", desc: "上一页"},
	OpPageDown:                       {name: "pageDown", desc: "下一页"},
	OpPlayOrToggle:                   {name: "playortoggle", desc: "播放/暂停"},
	OpToggle:                         {name: "toggle", desc: "切换播放状态"},
	OpPrevious:                       {name: "previous", desc: "上一首"},
	OpNext:                           {name: "next", desc: "下一首"},
	OpSeekBackward5s:                 {name: "backwardFiveSec", desc: "快退5秒"},
	OpSeekBackward1s:                 {name: "backwardOneSec", desc: "快退1秒"},
	OpSeekForward5s:                  {name: "forwardFiveSec", desc: "快进5秒"},
	OpSeekForward10s:                 {name: "forwardTenSec", desc: "快进10秒"},
	OpVolumeDown:                     {name: "downVolume", desc: "减小音量"},
	OpVolumeUp:                       {name: "upVolume", desc: "加大音量"},
	OpSwitchPlayMode:                 {name: "switchPlayMode", desc: "切换播放模式"},
	OpIntelligence:                   {name: "intelligence", desc: "心动模式"},
	OpClearSongCache:                 {name: "clearSongCache", desc: "清除音乐缓存"},
	OpLogout:                         {name: "logout", desc: "注销并退出"},
	OpCurPlaylist:                    {name: "curPlaylist", desc: "显示当前播放列表"},
	OpAppendSongsToNext:              {name: "appendSongsToNext", desc: "添加为下一曲播放"},
	OpAppendSongsToEnd:               {name: "appendSongsAfterCurPlaylist", desc: "添加到播放列表末尾"},
	OpDeleteSongFromPlaylist:         {name: "delSongFromCurPlaylist", desc: "从播放列表删除选中歌曲"},
	OpLikePlayingSong:                {name: "likePlayingSong", desc: "喜欢播放中歌曲"},
	OpDislikePlayingSong:             {name: "dislikePlayingSong", desc: "取消喜欢播放中歌曲"},
	OpTrashPlayingSong:               {name: "trashPlayingSong", desc: "标记播放中歌曲为不喜欢"},
	OpAddPlayingToUserPlaylist:       {name: "addPlayingSongToUserPlaylist", desc: "将播放中歌曲加入歌单"},
	OpRemovePlayingFromUserPlaylist:  {name: "removePlayingSongFromUserPlaylist", desc: "将播放歌曲从歌单中删除"},
	OpDownloadPlayingSong:            {name: "downloadPlayingSong", desc: "下载播放中歌曲"},
	OpDownloadPlayingSongLrc:         {name: "downloadPlayingSongLrc", desc: "下载当前播放音乐歌词"},
	OpAlbumOfPlayingSong:             {name: "openAlbumOfPlayingSong", desc: "播放中歌曲的所属专辑"},
	OpArtistOfPlayingSong:            {name: "openArtistOfPlayingSong", desc: "播放中歌曲的所属歌手"},
	OpOpenPlayingSongInWeb:           {name: "openPlayingSongInWeb", desc: "网页打开播放中歌曲"},
	OpLikeSelectedSong:               {name: "likeSelectedSong", desc: "喜欢选中歌曲"},
	OpDislikeSelectedSong:            {name: "dislikeSelectedSong", desc: "取消喜欢选中歌曲"},
	OpTrashSelectedSong:              {name: "trashSelectedSong", desc: "标记选中歌曲为不喜欢"},
	OpAddSelectedToUserPlaylist:      {name: "addSelectedSongToUserPlaylist", desc: "将选中歌曲加入歌单"},
	OpRemoveSelectedFromUserPlaylist: {name: "removeSelectedSongFromUserPlaylist", desc: "将选中歌曲从歌单中删除"},
	OpDownloadSelectedSong:           {name: "downloadSelectedSong", desc: "下载选中歌曲"},
	OpAlbumOfSelectedSong:            {name: "openAlbumOfSelectedSong", desc: "选中歌曲的所属专辑"},
	OpArtistOfSelectedSong:           {name: "openArtistOfSelectedSong", desc: "选中歌曲的所属歌手"},
	OpOpenSelectedItemInWeb:          {name: "openSelectedItemInWeb", desc: "网页打开选中歌曲/专辑..."},
	OpCollectSelectedPlaylist:        {name: "collectSelectedPlaylist", desc: "收藏选中歌单"},
	OpDiscollectSelectedPlaylist:     {name: "discollectSelectedPlaylist", desc: "取消收藏选中歌单"},
}

// 默认操作 -> 快捷键数组映射
var defaultOperationKeys = map[OperateType][]string{
	OpRerenderUI:   {"r", "R"},
	OpMoveLeft:     {"h", "H", "left"},
	OpMoveRight:    {"l", "L", "right"},
	OpMoveUp:       {"k", "K", "up"},
	OpMoveDown:     {"j", "J", "down"},
	OpMoveToTop:    {"g"},
	OpMoveToBottom: {"G"},
	OpEnter:        {"n", "N", "enter"},
	OpGoBack:       {"b", "B", "esc"},
	OpSearch:       {"/", "／", "、"},
	OpQuit:         {"q", "Q"},

	OpHelp:           {"?", "？"},
	OpPageUp:         {"ctrl+u", "pgup"},
	OpPageDown:       {"ctrl+d", "pgdown"},
	OpPlayOrToggle:   {"space", " ", "　"},
	OpToggle:         {},
	OpPrevious:       {"[", "【"},
	OpNext:           {"]", "】"},
	OpSeekBackward5s: {"X"},
	OpSeekBackward1s: {"x"},
	OpSeekForward5s:  {"v"},
	OpSeekForward10s: {"V"},
	OpVolumeDown:     {"-", "−", "ー"},
	OpVolumeUp:       {"=", "＝"},
	OpSwitchPlayMode: {"p"},
	OpIntelligence:   {"P"},
	OpClearSongCache: {"u", "U"},
	OpLogout:         {"w", "W"},
	OpCurPlaylist:    {"c", "C"},

	OpAppendSongsToNext:              {"e"},
	OpAppendSongsToEnd:               {"E"},
	OpDeleteSongFromPlaylist:         {"\\", "、"},
	OpLikePlayingSong:                {",", "，"},
	OpDislikePlayingSong:             {".", "。"},
	OpTrashPlayingSong:               {"t"},
	OpAddPlayingToUserPlaylist:       {"`"},
	OpRemovePlayingFromUserPlaylist:  {"~", "～"},
	OpDownloadPlayingSong:            {"d"},
	OpDownloadPlayingSongLrc:         {"ctrl+l"},
	OpAlbumOfPlayingSong:             {"a"},
	OpArtistOfPlayingSong:            {"s"},
	OpOpenPlayingSongInWeb:           {"o"},
	OpLikeSelectedSong:               {"<", "〈", "＜", "《", "«"},
	OpDislikeSelectedSong:            {">", "〉", "＞", "》", "»"},
	OpTrashSelectedSong:              {"T"},
	OpAddSelectedToUserPlaylist:      {"tab"},
	OpRemoveSelectedFromUserPlaylist: {"shift+tab"},
	OpDownloadSelectedSong:           {"D"},
	OpAlbumOfSelectedSong:            {"A"},
	OpArtistOfSelectedSong:           {"S"},
	OpOpenSelectedItemInWeb:          {"O"},
	OpCollectSelectedPlaylist:        {";", ":", "：", "；"},
	OpDiscollectSelectedPlaylist:     {"'", "\""},
}

func UserOperateToKeys() map[OperateType][]string {
	return defaultOperationKeys
}

var specialKeyDisplayMap = map[string]string{
	"pgup":      "PageUp",
	"pgdown":    "PageDown",
	"home":      "Home",
	"end":       "End",
	"delete":    "Delete",
	"insert":    "Insert",
	"backspace": "Backspace",
	"enter":     "Enter",
	"esc":       "Esc",
	"tab":       "Tab",
	"left":      "Left",
	"right":     "Right",
	"up":        "Up",
	"down":      "Down",
	"space":     "Space",

	" ": "Space",
	"　": "Space",
	"／": "/",
	"＝": "=",
	"？": "?",
	"，": ",",
	"；": ";",
	"：": ":",
	"’": "'",
	"、": "\\",
	"。": ".",
	"【": "[",
	"】": "]",
	"“": "\"",
	"”": "\"",
	"～": "~",
	"－": "-",
	"−": "-",
	"ー": "-",

	"＜": "<",
	"〈": "<",
	"《": "<",
	"«": "<",
	"＞": ">",
	"〉": ">",
	"》": ">",
	"»": ">",
}

// modifierMap defines how modifier str should be displayed.
var modifierMap = map[string]string{
	"ctrl+":  "Ctrl+",
	"shift+": "Shift+",
	"alt+":   "Alt+",
	"+tab":   "+Tab",
}

// FormatKeyForDisplay 格式化按键用于显示
func FormatKeyForDisplay(key string) string {
	if key == "" {
		return ""
	}

	if k, ok := specialKeyDisplayMap[key]; ok {
		return k
	}

	keyStr := key
	for modLower, modDisplay := range modifierMap {
		keyStr = strings.Replace(keyStr, modLower, modDisplay, -1)
	}
	if keyStr != key {
		return keyStr
	}

	// 对特殊按键名称进行标准化显示
	switch key {
	case "f1", "f2", "f3", "f4", "f5", "f6", "f7", "f8", "f9", "f10", "f11", "f12":
		return strings.ToUpper(key)
	}
	return key
}

// BuildKeyToOperateTypeMap 构建反向查找映射的函数，Key -> OperateType
func BuildKeyToOperateTypeMap() map[string]OperateType {
	effectiveBindings := UserOperateToKeys()
	keyMap := make(map[string]OperateType)
	for op, keys := range effectiveBindings {
		for _, key := range keys {
			keyMap[key] = op
		}
	}

	return keyMap
}
