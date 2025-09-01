// internal/keybindings/keybindings.go
package keybindings

import (
	"fmt"
	"log/slog"
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
	OpOpenSimiSongsOfPlayingSong
	OpLikeSelectedSong
	OpDislikeSelectedSong
	OpTrashSelectedSong
	OpAddSelectedToUserPlaylist
	OpRemoveSelectedFromUserPlaylist
	OpDownloadSelectedSong
	OpDownloadSelectedSongLrc
	OpAlbumOfSelectedSong
	OpArtistOfSelectedSong
	OpOpenSelectedItemInWeb
	OpCollectSelectedPlaylist
	OpDiscollectSelectedPlaylist
	OpOpenSimiSongsOfSelectedSong
	OpSharePlayingItem
	OpShareSelectItem

	OpActionOfSelected
	OpActionOfPlayingSong
)

var opNameToOperateMap = make(map[string]OperateType)

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
	OpOpenSimiSongsOfPlayingSong:     {name: "simiSongsOfPlayingSong", desc: "与播放中歌曲相似的歌曲"},
	OpLikeSelectedSong:               {name: "likeSelectedSong", desc: "喜欢选中歌曲"},
	OpDislikeSelectedSong:            {name: "dislikeSelectedSong", desc: "取消喜欢选中歌曲"},
	OpTrashSelectedSong:              {name: "trashSelectedSong", desc: "标记选中歌曲为不喜欢"},
	OpAddSelectedToUserPlaylist:      {name: "addSelectedSongToUserPlaylist", desc: "将选中歌曲加入歌单"},
	OpRemoveSelectedFromUserPlaylist: {name: "removeSelectedSongFromUserPlaylist", desc: "将选中歌曲从歌单中删除"},
	OpDownloadSelectedSong:           {name: "downloadSelectedSong", desc: "下载选中歌曲"},
	OpDownloadSelectedSongLrc:        {name: "downloadSelectedSongLrc", desc: "下载选中歌曲的歌词"},
	OpAlbumOfSelectedSong:            {name: "openAlbumOfSelectedSong", desc: "选中歌曲的所属专辑"},
	OpArtistOfSelectedSong:           {name: "openArtistOfSelectedSong", desc: "选中歌曲的所属歌手"},
	OpOpenSelectedItemInWeb:          {name: "openSelectedItemInWeb", desc: "网页打开选中歌曲/专辑..."},
	OpCollectSelectedPlaylist:        {name: "collectSelectedPlaylist", desc: "收藏选中歌单"},
	OpDiscollectSelectedPlaylist:     {name: "discollectSelectedPlaylist", desc: "取消收藏选中歌单"},
	OpOpenSimiSongsOfSelectedSong:    {name: "simiSongsOfSelectedSong", desc: "与选中歌曲相似的歌曲"},
	OpSharePlayingItem:               {name: "sharePlayingItem", desc: "分享当前播放"},
	OpShareSelectItem:                {name: "shareSelectItem", desc: "分享当前选中"},

	OpActionOfSelected:    {name: "actionOfSelected", desc: "对于选中项或当前播放的操作"},
	OpActionOfPlayingSong: {name: "actionOfPlayingSong", desc: "对于当前播放的操作"},
}

// 默认操作 -> 快捷键数组映射
var defaultBaseOperateToKeys = map[OperateType][]string{
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
}

var defaultOtherOperateToKeys = map[OperateType][]string{
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
	OpLogout:         {"W"},
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
	OpOpenSimiSongsOfPlayingSong:     {"f"},
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
	OpOpenSimiSongsOfSelectedSong:    {"F"},

	OpActionOfSelected:    {"m"},
	OpActionOfPlayingSong: {"M"},
}

var userOperateToKeys map[OperateType][]string

func UserOperateToKeys() map[OperateType][]string {
	return userOperateToKeys
}

// InitDefaults 生成操作绑定的 map
func InitDefaults(useDefault bool) map[OperateType][]string {
	if !useDefault {
		baseCopy := make(map[OperateType][]string, len(defaultBaseOperateToKeys))
		for op, keys := range defaultBaseOperateToKeys {
			baseCopy[op] = append([]string(nil), keys...)
		}
		userOperateToKeys = baseCopy
		return baseCopy
	}

	mergedMap := make(map[OperateType][]string, len(defaultBaseOperateToKeys)+len(defaultOtherOperateToKeys))
	for op, keys := range defaultBaseOperateToKeys {
		mergedMap[op] = append([]string(nil), keys...)
	}
	for op, keys := range defaultOtherOperateToKeys {
		mergedMap[op] = append([]string(nil), keys...)
	}
	userOperateToKeys = mergedMap
	return mergedMap
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

// BuildEffectiveBindings 构建最终生效的按键绑定
func BuildEffectiveBindings(userKeyBindings map[string]string, useDefault bool) map[OperateType][]string {
	defaultBindings := InitDefaults(useDefault)

	effectiveBindings := make(map[OperateType][]string, len(defaultBindings))
	for op, keys := range defaultBindings {
		effectiveBindings[op] = append([]string(nil), keys...)
	}

	if len(userKeyBindings) == 0 {
		return effectiveBindings
	}

	// 预处理用户配置，移除不存在操作
	preprocessing := make(map[OperateType][]string)
	for opStr, keysStr := range userKeyBindings {
		op, ok := GetOperationFromName(opStr)
		if !ok {
			slog.Warn(fmt.Sprintf("配置文件 [keybindings] 中发现未知操作 '%s'，已忽略", opStr))
			continue
		}

		// FIXME: 对于内置键绑定的操作，考虑使用（新增）而非不处理。或寻求其他方式进行支持
		if op < 0 {
			slog.Warn(fmt.Sprintf("内置操作 '%s'，暂不可更改，已忽略", opStr))
			continue
		}

		if keysStr == "" {
			slog.Info(fmt.Sprintf("解绑操作 '%s (%s)'", opStr, op.Desc()))
			effectiveBindings[op] = []string{}
			continue
		}

		keys := splitKeys(keysStr)
		if len(keys) != 0 {
			preprocessing[op] = keys
		} else {
			slog.Warn(fmt.Sprintf("操作 '%s (%s)' 的用户配置只无有效的绑定，忽略", op, op.Desc()))
		}
	}

	// 构建临时的反向映射，用于冲突检测 (基于当前的 effectiveBindings)
	getKeyToOp := func() map[string]OperateType {
		keyToOp := make(map[string]OperateType)
		for op, keys := range effectiveBindings {
			for _, key := range keys {
				keyToOp[key] = op
			}
		}
		return keyToOp
	}
	keyToOp := getKeyToOp()

	hardcodedKeys := getHardCordKeys()
	conflicts := make(map[string][]OperateType) // 记录最终的冲突信息 Key -> Conflicting Ops
	for op, keys := range preprocessing {
		validKeys := make([]string, 0, len(keys))
		skippedKeys := []string{}
		for _, key := range keys {
			// 跳过对硬编码按键的处理
			if _, isHardcoded := hardcodedKeys[key]; isHardcoded {
				skippedKeys = append(skippedKeys, key)
				continue
			}

			// 处理冲突
			if existingOp, ok := keyToOp[key]; ok && existingOp != op {
				if _, recorded := conflicts[key]; !recorded {
					conflicts[key] = []OperateType{existingOp, op}
				} else {
					conflicts[key] = append(conflicts[key], op)
				}

				// 在记录冲突后，需要移除旧绑定操作对此键的占用（如果旧操作还有其他键）
				if oldKeys, ok := effectiveBindings[existingOp]; ok {
					cleanedOldKeys := make([]string, 0, len(oldKeys)-1)
					for _, oldKey := range oldKeys {
						if oldKey != key {
							cleanedOldKeys = append(cleanedOldKeys, oldKey)
						}
					}
					effectiveBindings[existingOp] = cleanedOldKeys
				}
			}

			validKeys = append(validKeys, key)
		}
		if len(skippedKeys) > 0 {
			slog.Warn(fmt.Sprintf("操作 '%s (%s)' 的用户配置中包含硬编码按键 [%s]，这些特定绑定已被忽略。", op, op.Desc(), strings.Join(skippedKeys, ", ")))
		}
		// 只用有效的非内建键覆盖或设置绑定
		if len(validKeys) > 0 {
			effectiveBindings[op] = validKeys
		} else {
			slog.Warn(fmt.Sprintf("操作 '%s (%s)' 的用户配置只无有效的绑定，忽略", op, op.Desc()))
		}
	}

	// 打印冲突信息
	if len(conflicts) != 0 {
		keyToOp := getKeyToOp()
		for key, ops := range conflicts {
			opStrs := make([]string, len(ops))
			for _, op := range ops {
				opStrs = append(opStrs, fmt.Sprintf("%s (%s)", op, op.Desc()))
			}
			finalOp := keyToOp[key]
			slog.Warn(fmt.Sprintf("按键 '%s' 被分配给多个操作: %s。最终生效的操作是: '%s (%s)'",
				key, strings.Join(opStrs, ", "), finalOp, finalOp.Desc()))
		}
	}

	userOperateToKeys = effectiveBindings // 更新用户设置
	return effectiveBindings
}

// BuildKeyToOperateTypeMap 构建反向查找映射的函数，Key -> OperateType
func BuildKeyToOperateTypeMap(effectiveBindings map[OperateType][]string) map[string]OperateType {
	keyMap := make(map[string]OperateType)
	if effectiveBindings == nil {
		return keyMap
	}

	for op, keys := range userOperateToKeys {
		for _, key := range keys {
			keyMap[key] = op
		}
	}

	return keyMap
}

// 使用操作名称查询 OperateType
func GetOperationFromName(opName string) (OperateType, bool) {
	op, ok := opNameToOperateMap[opName]
	return op, ok
}

// GetHardCordKeys 获取被硬编码在 foxful-cli 的按键
func getHardCordKeys() map[string]struct{} {
	maps := make(map[string]struct{})
	for op := range keyBindingsRegistry {
		if op >= 0 {
			continue
		}
		for _, k := range op.Keys() {
			if k != "" {
				maps[k] = struct{}{}
			}
		}
	}
	return maps
}

// SplitKeys 用于将配置的键字符串解析为键切片
func splitKeys(s string) []string {
	// FIXME: 对于快捷键本身使用 "," 的，需要做额外适配，使用空格作为分隔符？
	keys := strings.Split(s, ",")
	trimmedKeys := make([]string, 0, len(keys))
	for _, key := range keys {
		// FIXME: 可能会错误移除空格（包括Unicode \u3000 : "　"），考虑使用 "space" 替代或其他处理方式(strings.TrimFunc)
		trimmedKey := strings.TrimSpace(key)
		if trimmedKey == "" {
			continue
		}
		// 统一为小写
		if len(trimmedKey) > 1 {
			trimmedKey = strings.ToLower(trimmedKey)
		}
		if trimmedKey == "space" {
			trimmedKey = " "
		}
		trimmedKeys = append(trimmedKeys, trimmedKey)
	}
	return trimmedKeys
}

// 创建一个操作名称到 OperateType 的映射
func init() {
	maps := make(map[string]OperateType, len(keyBindingsRegistry))
	for op := range keyBindingsRegistry {
		if op.Name() != "" {
			maps[op.Name()] = op
		}
	}
	opNameToOperateMap = maps
}
