package handlers

import (
	"fmt"
	"strings"
)

// KeyMapping 键盘映射
type KeyMapping struct {
	Key         string   `json:"key"`
	Modifiers   []string `json:"modifiers"`
	Action      string   `json:"action"`
	Context     string   `json:"context"`     // 上下文：global, main, player, search, playlist
	Description string   `json:"description"`
}

// KeyMappingManager 键盘映射管理器
type KeyMappingManager struct {
	mappings map[string][]KeyMapping // context -> mappings
}

// NewKeyMappingManager 创建键盘映射管理器
func NewKeyMappingManager() *KeyMappingManager {
	manager := &KeyMappingManager{
		mappings: make(map[string][]KeyMapping),
	}
	
	// 加载默认键盘映射
	manager.loadDefaultMappings()
	
	return manager
}

// loadDefaultMappings 加载默认键盘映射
func (km *KeyMappingManager) loadDefaultMappings() {
	// 全局快捷键
	globalMappings := []KeyMapping{
		{Key: "q", Action: "quit", Context: "global", Description: "退出程序"},
		{Key: "c", Modifiers: []string{"ctrl"}, Action: "quit", Context: "global", Description: "强制退出"},
		{Key: "h", Action: "help", Context: "global", Description: "显示帮助"},
		{Key: "?", Action: "help", Context: "global", Description: "显示帮助"},
		{Key: "r", Action: "refresh", Context: "global", Description: "刷新界面"},
		{Key: "F5", Action: "refresh", Context: "global", Description: "刷新界面"},
		{Key: "t", Action: "toggle_theme", Context: "global", Description: "切换主题"},
		{Key: "F11", Action: "toggle_fullscreen", Context: "global", Description: "切换全屏"},
		
		// 导航快捷键
		{Key: "k", Action: "navigate_up", Context: "global", Description: "向上导航"},
		{Key: "up", Action: "navigate_up", Context: "global", Description: "向上导航"},
		{Key: "j", Action: "navigate_down", Context: "global", Description: "向下导航"},
		{Key: "down", Action: "navigate_down", Context: "global", Description: "向下导航"},
		{Key: "h", Action: "navigate_left", Context: "global", Description: "向左导航"},
		{Key: "left", Action: "navigate_left", Context: "global", Description: "向左导航"},
		{Key: "l", Action: "navigate_right", Context: "global", Description: "向右导航"},
		{Key: "right", Action: "navigate_right", Context: "global", Description: "向右导航"},
		{Key: "ctrl+u", Action: "page_up", Context: "global", Description: "向上翻页"},
		{Key: "page_up", Action: "page_up", Context: "global", Description: "向上翻页"},
		{Key: "ctrl+d", Action: "page_down", Context: "global", Description: "向下翻页"},
		{Key: "page_down", Action: "page_down", Context: "global", Description: "向下翻页"},
		{Key: "g", Action: "home", Context: "global", Description: "跳到开头"},
		{Key: "home", Action: "home", Context: "global", Description: "跳到开头"},
		{Key: "G", Action: "end", Context: "global", Description: "跳到结尾"},
		{Key: "end", Action: "end", Context: "global", Description: "跳到结尾"},
		{Key: "enter", Action: "select", Context: "global", Description: "选择/确认"},
		{Key: "return", Action: "select", Context: "global", Description: "选择/确认"},
		{Key: "esc", Action: "back", Context: "global", Description: "返回"},
		{Key: "backspace", Action: "back", Context: "global", Description: "返回"},
		
		// 播放控制快捷键
		{Key: " ", Action: "toggle_play", Context: "global", Description: "播放/暂停"},
		{Key: "p", Action: "toggle_play", Context: "global", Description: "播放/暂停"},
		{Key: "s", Action: "stop", Context: "global", Description: "停止播放"},
		{Key: "n", Action: "next", Context: "global", Description: "下一首"},
		{Key: ">", Action: "next", Context: "global", Description: "下一首"},
		{Key: "N", Action: "previous", Context: "global", Description: "上一首"},
		{Key: "<", Action: "previous", Context: "global", Description: "上一首"},
		{Key: "z", Action: "shuffle", Context: "global", Description: "随机播放"},
		{Key: "R", Action: "repeat", Context: "global", Description: "重复播放"},
		{Key: "m", Action: "mute", Context: "global", Description: "静音"},
		
		// 音量控制
		{Key: "+", Action: "volume_up", Context: "global", Description: "音量+"},
		{Key: "=", Action: "volume_up", Context: "global", Description: "音量+"},
		{Key: "-", Action: "volume_down", Context: "global", Description: "音量-"},
		{Key: "_", Action: "volume_down", Context: "global", Description: "音量-"},
		
		// 进度控制
		{Key: "f", Action: "seek_forward", Context: "global", Description: "快进"},
		{Key: "→", Action: "seek_forward", Context: "global", Description: "快进"},
		{Key: "b", Action: "seek_backward", Context: "global", Description: "快退"},
		{Key: "←", Action: "seek_backward", Context: "global", Description: "快退"},
		
		// 视图切换
		{Key: "1", Action: "view_main", Context: "global", Description: "主菜单"},
		{Key: "2", Action: "view_player", Context: "global", Description: "播放器"},
		{Key: "3", Action: "view_playlist", Context: "global", Description: "播放列表"},
		{Key: "4", Action: "view_search", Context: "global", Description: "搜索"},
		{Key: "5", Action: "view_lyrics", Context: "global", Description: "歌词"},
		{Key: "6", Action: "view_help", Context: "global", Description: "帮助"},
	}
	
	// 主菜单快捷键
	mainMappings := []KeyMapping{
		{Key: "/", Action: "search", Context: "main", Description: "搜索"},
		{Key: "ctrl+f", Action: "search", Context: "main", Description: "搜索"},
		{Key: "u", Action: "user_login", Context: "main", Description: "用户登录"},
		{Key: "U", Action: "user_logout", Context: "main", Description: "用户登出"},
		{Key: "c", Action: "clear_cache", Context: "main", Description: "清除缓存"},
		{Key: "i", Action: "show_info", Context: "main", Description: "显示信息"},
	}
	
	// 播放器快捷键
	playerMappings := []KeyMapping{
		{Key: "L", Action: "toggle_lyrics", Context: "player", Description: "显示/隐藏歌词"},
		{Key: "v", Action: "toggle_visualizer", Context: "player", Description: "显示/隐藏可视化"},
		{Key: "i", Action: "song_info", Context: "player", Description: "歌曲信息"},
		{Key: "d", Action: "download", Context: "player", Description: "下载歌曲"},
		{Key: "a", Action: "add_to_playlist", Context: "player", Description: "添加到歌单"},
		{Key: "x", Action: "remove_from_playlist", Context: "player", Description: "从歌单移除"},
		{Key: "ctrl+l", Action: "locate_current", Context: "player", Description: "定位当前歌曲"},
	}
	
	// 搜索快捷键
	searchMappings := []KeyMapping{
		{Key: "tab", Action: "search_type_next", Context: "search", Description: "切换搜索类型"},
		{Key: "shift+tab", Action: "search_type_prev", Context: "search", Description: "切换搜索类型"},
		{Key: "ctrl+a", Action: "select_all", Context: "search", Description: "全选"},
		{Key: "ctrl+c", Action: "copy_selected", Context: "search", Description: "复制选中"},
		{Key: "delete", Action: "clear_input", Context: "search", Description: "清空输入"},
		{Key: "F3", Action: "search_next", Context: "search", Description: "下一个结果"},
		{Key: "shift+F3", Action: "search_prev", Context: "search", Description: "上一个结果"},
	}
	
	// 播放列表快捷键
	playlistMappings := []KeyMapping{
		{Key: "d", Action: "remove_song", Context: "playlist", Description: "删除歌曲"},
		{Key: "D", Action: "clear_playlist", Context: "playlist", Description: "清空列表"},
		{Key: "s", Action: "save_playlist", Context: "playlist", Description: "保存列表"},
		{Key: "o", Action: "load_playlist", Context: "playlist", Description: "加载列表"},
		{Key: "r", Action: "shuffle_playlist", Context: "playlist", Description: "随机排序"},
		{Key: "ctrl+s", Action: "sort_playlist", Context: "playlist", Description: "排序列表"},
		{Key: "ctrl+r", Action: "reverse_playlist", Context: "playlist", Description: "反转列表"},
		{Key: "a", Action: "add_all_to_queue", Context: "playlist", Description: "全部加入队列"},
	}
	
	// 注册映射
	km.mappings["global"] = globalMappings
	km.mappings["main"] = mainMappings
	km.mappings["player"] = playerMappings
	km.mappings["search"] = searchMappings
	km.mappings["playlist"] = playlistMappings
}

// FindAction 根据键盘输入查找对应的动作
func (km *KeyMappingManager) FindAction(key string, modifiers []string, context string) (string, bool) {
	// 标准化键名
	normalizedKey := km.normalizeKey(key)
	normalizedModifiers := km.normalizeModifiers(modifiers)
	
	// 首先在特定上下文中查找
	if mappings, exists := km.mappings[context]; exists {
		for _, mapping := range mappings {
			if km.matchesMapping(normalizedKey, normalizedModifiers, mapping) {
				return mapping.Action, true
			}
		}
	}
	
	// 然后在全局上下文中查找
	if context != "global" {
		if globalMappings, exists := km.mappings["global"]; exists {
			for _, mapping := range globalMappings {
				if km.matchesMapping(normalizedKey, normalizedModifiers, mapping) {
					return mapping.Action, true
				}
			}
		}
	}
	
	return "", false
}

// matchesMapping 检查键盘输入是否匹配映射
func (km *KeyMappingManager) matchesMapping(key string, modifiers []string, mapping KeyMapping) bool {
	// 检查键是否匹配
	if km.normalizeKey(mapping.Key) != key {
		return false
	}
	
	// 检查修饰键是否匹配
	mappingModifiers := km.normalizeModifiers(mapping.Modifiers)
	
	// 如果修饰键数量不同，不匹配
	if len(modifiers) != len(mappingModifiers) {
		return false
	}
	
	// 检查每个修饰键
	for _, mod := range modifiers {
		found := false
		for _, mappingMod := range mappingModifiers {
			if mod == mappingMod {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	
	return true
}

// normalizeKey 标准化键名
func (km *KeyMappingManager) normalizeKey(key string) string {
	// 转换为小写
	normalized := strings.ToLower(key)
	
	// 处理特殊键名映射
	keyMappings := map[string]string{
		"space":     " ",
		"return":    "enter",
		"escape":    "esc",
		"delete":    "del",
		"backspace": "bs",
		"tab":       "\t",
		"up":        "↑",
		"down":      "↓",
		"left":      "←",
		"right":     "→",
		"page_up":   "pgup",
		"page_down": "pgdn",
		"home":      "home",
		"end":       "end",
	}
	
	if mapped, exists := keyMappings[normalized]; exists {
		return mapped
	}
	
	return normalized
}

// normalizeModifiers 标准化修饰键
func (km *KeyMappingManager) normalizeModifiers(modifiers []string) []string {
	normalized := make([]string, len(modifiers))
	for i, mod := range modifiers {
		normalized[i] = strings.ToLower(mod)
	}
	return normalized
}

// GetMappings 获取指定上下文的键盘映射
func (km *KeyMappingManager) GetMappings(context string) []KeyMapping {
	if mappings, exists := km.mappings[context]; exists {
		return mappings
	}
	return []KeyMapping{}
}

// GetAllMappings 获取所有键盘映射
func (km *KeyMappingManager) GetAllMappings() map[string][]KeyMapping {
	return km.mappings
}

// AddMapping 添加键盘映射
func (km *KeyMappingManager) AddMapping(mapping KeyMapping) {
	if km.mappings[mapping.Context] == nil {
		km.mappings[mapping.Context] = []KeyMapping{}
	}
	km.mappings[mapping.Context] = append(km.mappings[mapping.Context], mapping)
}

// RemoveMapping 移除键盘映射
func (km *KeyMappingManager) RemoveMapping(context, key string, modifiers []string) bool {
	mappings, exists := km.mappings[context]
	if !exists {
		return false
	}
	
	normalizedKey := km.normalizeKey(key)
	normalizedModifiers := km.normalizeModifiers(modifiers)
	
	for i, mapping := range mappings {
		if km.matchesMapping(normalizedKey, normalizedModifiers, mapping) {
			// 移除找到的映射
			km.mappings[context] = append(mappings[:i], mappings[i+1:]...)
			return true
		}
	}
	
	return false
}

// UpdateMapping 更新键盘映射
func (km *KeyMappingManager) UpdateMapping(context, key string, modifiers []string, newAction, newDescription string) bool {
	mappings, exists := km.mappings[context]
	if !exists {
		return false
	}
	
	normalizedKey := km.normalizeKey(key)
	normalizedModifiers := km.normalizeModifiers(modifiers)
	
	for i, mapping := range mappings {
		if km.matchesMapping(normalizedKey, normalizedModifiers, mapping) {
			// 更新找到的映射
			km.mappings[context][i].Action = newAction
			km.mappings[context][i].Description = newDescription
			return true
		}
	}
	
	return false
}

// LoadMappingsFromConfig 从配置加载键盘映射
func (km *KeyMappingManager) LoadMappingsFromConfig(config map[string]interface{}) error {
	// 清空现有映射
	km.mappings = make(map[string][]KeyMapping)
	
	// 加载默认映射
	km.loadDefaultMappings()
	
	// 从配置中覆盖映射
	for context, contextConfig := range config {
		if mappingsConfig, ok := contextConfig.([]interface{}); ok {
			mappings := []KeyMapping{}
			for _, mappingConfig := range mappingsConfig {
				if mappingMap, ok := mappingConfig.(map[string]interface{}); ok {
					mapping := KeyMapping{
						Context: context,
					}
					
					if key, ok := mappingMap["key"].(string); ok {
						mapping.Key = key
					}
					
					if action, ok := mappingMap["action"].(string); ok {
						mapping.Action = action
					}
					
					if description, ok := mappingMap["description"].(string); ok {
						mapping.Description = description
					}
					
					if modifiers, ok := mappingMap["modifiers"].([]interface{}); ok {
						for _, mod := range modifiers {
							if modStr, ok := mod.(string); ok {
								mapping.Modifiers = append(mapping.Modifiers, modStr)
							}
						}
					}
					
					mappings = append(mappings, mapping)
				}
			}
			km.mappings[context] = mappings
		}
	}
	
	return nil
}

// ExportMappingsToConfig 导出键盘映射到配置
func (km *KeyMappingManager) ExportMappingsToConfig() map[string]interface{} {
	config := make(map[string]interface{})
	
	for context, mappings := range km.mappings {
		mappingsConfig := make([]interface{}, len(mappings))
		for i, mapping := range mappings {
			mappingConfig := map[string]interface{}{
				"key":         mapping.Key,
				"action":      mapping.Action,
				"description": mapping.Description,
			}
			
			if len(mapping.Modifiers) > 0 {
				mappingConfig["modifiers"] = mapping.Modifiers
			}
			
			mappingsConfig[i] = mappingConfig
		}
		config[context] = mappingsConfig
	}
	
	return config
}

// GetHelpText 获取帮助文本
func (km *KeyMappingManager) GetHelpText(context string) []string {
	lines := []string{}
	
	// 添加上下文标题
	contextNames := map[string]string{
		"global":   "全局快捷键",
		"main":     "主菜单快捷键",
		"player":   "播放器快捷键",
		"search":   "搜索快捷键",
		"playlist": "播放列表快捷键",
	}
	
	if name, exists := contextNames[context]; exists {
		lines = append(lines, fmt.Sprintf("=== %s ===", name))
	} else {
		lines = append(lines, fmt.Sprintf("=== %s ===", context))
	}
	lines = append(lines, "")
	
	// 添加映射列表
	if mappings, exists := km.mappings[context]; exists {
		for _, mapping := range mappings {
			keyText := mapping.Key
			if len(mapping.Modifiers) > 0 {
				keyText = strings.Join(mapping.Modifiers, "+") + "+" + keyText
			}
			
			lines = append(lines, fmt.Sprintf("  %-15s %s", keyText, mapping.Description))
		}
	}
	
	return lines
}

// GetAllHelpText 获取所有上下文的帮助文本
func (km *KeyMappingManager) GetAllHelpText() []string {
	lines := []string{}
	
	contexts := []string{"global", "main", "player", "search", "playlist"}
	
	for i, context := range contexts {
		if i > 0 {
			lines = append(lines, "", "")
		}
		lines = append(lines, km.GetHelpText(context)...)
	}
	
	return lines
}

// ValidateMapping 验证键盘映射
func (km *KeyMappingManager) ValidateMapping(mapping KeyMapping) error {
	if mapping.Key == "" {
		return fmt.Errorf("key cannot be empty")
	}
	
	if mapping.Action == "" {
		return fmt.Errorf("action cannot be empty")
	}
	
	if mapping.Context == "" {
		return fmt.Errorf("context cannot be empty")
	}
	
	// 只检查同一上下文中的冲突
	if mappings, exists := km.mappings[mapping.Context]; exists {
		normalizedKey := km.normalizeKey(mapping.Key)
		normalizedModifiers := km.normalizeModifiers(mapping.Modifiers)
		
		for _, existingMapping := range mappings {
			if km.matchesMapping(normalizedKey, normalizedModifiers, existingMapping) {
				if existingMapping.Action != mapping.Action {
					return fmt.Errorf("key mapping conflict: %s is already mapped to %s", mapping.Key, existingMapping.Action)
				}
			}
		}
	}
	
	return nil
}