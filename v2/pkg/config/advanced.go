package config

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/knadh/koanf/v2"
)

// AdvancedConfigManager 高级配置管理器接口
type AdvancedConfigManager interface {
	ConfigManager

	// 热更新功能
	EnableHotReload(ctx context.Context) error
	DisableHotReload() error
	IsHotReloadEnabled() bool
	OnConfigChanged(callback ConfigChangeCallback) error
	RemoveConfigChangeCallback(callbackID string) error

	// 版本管理功能
	GetVersion() string
	GetVersionHistory() ([]*ConfigVersion, error)
	CreateSnapshot(description string) (*ConfigVersion, error)
	RollbackToVersion(versionID string) error
	CompareVersions(version1, version2 string) (*ConfigDiff, error)
	DeleteVersion(versionID string) error

	// 模板和继承功能
	LoadTemplate(templatePath string) error
	ApplyTemplate(templateName string, variables map[string]interface{}) error
	SetInheritance(parentConfig string) error
	ResolveInheritance() error
	GetEffectiveConfig() (*koanf.Koanf, error)

	// 加密和安全功能
	SetEncryptionKey(key []byte) error
	EncryptSensitiveData(keys []string) error
	DecryptSensitiveData(keys []string) error
	IsEncrypted(key string) bool
	SetAccessControl(rules *AccessControlRules) error
	CheckAccess(operation string, key string, user string) bool

	// 批量操作和事务支持
	BeginTransaction() (ConfigTransaction, error)
	BatchUpdate(updates map[string]interface{}) error
	ImportConfig(source string, format string) error
	ExportConfig(target string, format string) error

	// 监控和统计
	GetStats() *ConfigStats
	GetChangeLog() ([]*ConfigChange, error)
	ClearChangeLog() error
}

// ConfigChangeCallback 配置变更回调函数
type ConfigChangeCallback func(change *ConfigChange) error

// ConfigVersion 配置版本信息
type ConfigVersion struct {
	ID          string                 `json:"id"`
	Description string                 `json:"description"`
	Timestamp   time.Time              `json:"timestamp"`
	Data        map[string]interface{} `json:"data"`
	Checksum    string                 `json:"checksum"`
	CreatedBy   string                 `json:"created_by"`
	Tags        []string               `json:"tags"`
}

// ConfigDiff 配置差异信息
type ConfigDiff struct {
	Added    map[string]interface{} `json:"added"`
	Modified map[string]DiffValue   `json:"modified"`
	Deleted  map[string]interface{} `json:"deleted"`
}

// DiffValue 差异值
type DiffValue struct {
	Old interface{} `json:"old"`
	New interface{} `json:"new"`
}

// ConfigChange 配置变更记录
type ConfigChange struct {
	ID        string      `json:"id"`
	Timestamp time.Time   `json:"timestamp"`
	Operation string      `json:"operation"` // set, delete, reload, rollback
	Key       string      `json:"key"`
	OldValue  interface{} `json:"old_value"`
	NewValue  interface{} `json:"new_value"`
	User      string      `json:"user"`
	Source    string      `json:"source"`
}

// ConfigTransaction 配置事务接口
type ConfigTransaction interface {
	Set(key string, value interface{}) error
	Delete(key string) error
	Commit() error
	Rollback() error
	IsActive() bool
}

// AccessControlRules 访问控制规则
type AccessControlRules struct {
	DefaultPolicy string                    `json:"default_policy"` // allow, deny
	Rules         []AccessRule              `json:"rules"`
	Roles         map[string][]string       `json:"roles"`
	Users         map[string]UserPermission `json:"users"`
}

// AccessRule 访问规则
type AccessRule struct {
	Pattern     string   `json:"pattern"`     // 配置键模式
	Operations  []string `json:"operations"`  // read, write, delete
	Users       []string `json:"users"`       // 允许的用户
	Roles       []string `json:"roles"`       // 允许的角色
	Policy      string   `json:"policy"`      // allow, deny
	Conditions  []string `json:"conditions"`  // 额外条件
}

// UserPermission 用户权限
type UserPermission struct {
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
	Restricted  []string `json:"restricted"`
}

// ConfigStats 配置统计信息
type ConfigStats struct {
	TotalKeys        int           `json:"total_keys"`
	EncryptedKeys    int           `json:"encrypted_keys"`
	LastModified     time.Time     `json:"last_modified"`
	VersionCount     int           `json:"version_count"`
	ChangeCount      int           `json:"change_count"`
	HotReloadEnabled bool          `json:"hot_reload_enabled"`
	MemoryUsage      int64         `json:"memory_usage"`
	LoadTime         time.Duration `json:"load_time"`
}

// AdvancedManager 高级配置管理器实现
type AdvancedManager struct {
	*Manager

	// 热更新相关
	watcher         *fsnotify.Watcher
	hotReloadCtx    context.Context
	hotReloadCancel context.CancelFunc
	callbacks       map[string]ConfigChangeCallback
	callbackMutex   sync.RWMutex

	// 版本管理相关
	versionDir      string
	currentVersion  string
	versionHistory  []*ConfigVersion
	versionMutex    sync.RWMutex

	// 模板和继承相关
	templates       map[string]*koanf.Koanf
	parentConfig    string
	inheritanceTree map[string][]string
	templateMutex   sync.RWMutex

	// 加密相关
	encryptionKey   []byte
	encryptedKeys   map[string]bool
	encryptionMutex sync.RWMutex

	// 访问控制相关
	accessRules     *AccessControlRules
	accessMutex     sync.RWMutex

	// 统计和监控相关
	stats           *ConfigStats
	changeLog       []*ConfigChange
	statsMutex      sync.RWMutex

	// 事务相关
	activeTransactions map[string]*configTransaction
	transactionMutex   sync.RWMutex
}

// NewAdvancedManager 创建高级配置管理器
func NewAdvancedManager(configDir, configFile string) *AdvancedManager {
	baseManager := NewManager(configDir, configFile)
	
	am := &AdvancedManager{
		Manager:            baseManager,
		callbacks:          make(map[string]ConfigChangeCallback),
		versionDir:         filepath.Join(configDir, "versions"),
		versionHistory:     make([]*ConfigVersion, 0),
		templates:          make(map[string]*koanf.Koanf),
		inheritanceTree:    make(map[string][]string),
		encryptedKeys:      make(map[string]bool),
		activeTransactions: make(map[string]*configTransaction),
		stats: &ConfigStats{
			LastModified:     time.Now(),
			HotReloadEnabled: false,
		},
	}

	// 确保版本目录存在
	if err := os.MkdirAll(am.versionDir, 0755); err != nil {
		// 记录错误但不阻止创建
		fmt.Printf("Warning: failed to create version directory: %v\n", err)
	}

	// 加载版本历史
	if err := am.loadVersionHistory(); err != nil {
		fmt.Printf("Warning: failed to load version history: %v\n", err)
	}

	return am
}

// generateID 生成唯一ID
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// calculateChecksum 计算配置数据的校验和
func calculateChecksum(data map[string]interface{}) string {
	jsonData, _ := json.Marshal(data)
	hash := sha256.Sum256(jsonData)
	return base64.StdEncoding.EncodeToString(hash[:])
}

// encryptData 加密数据
func (am *AdvancedManager) encryptData(data []byte) ([]byte, error) {
	if len(am.encryptionKey) == 0 {
		return nil, fmt.Errorf("encryption key not set")
	}

	block, err := aes.NewCipher(am.encryptionKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

// decryptData 解密数据
func (am *AdvancedManager) decryptData(data []byte) ([]byte, error) {
	if len(am.encryptionKey) == 0 {
		return nil, fmt.Errorf("encryption key not set")
	}

	block, err := aes.NewCipher(am.encryptionKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(data) < gcm.NonceSize() {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := data[:gcm.NonceSize()], data[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// loadVersionHistory 加载版本历史
func (am *AdvancedManager) loadVersionHistory() error {
	am.versionMutex.Lock()
	defer am.versionMutex.Unlock()

	historyFile := filepath.Join(am.versionDir, "history.json")
	if _, err := os.Stat(historyFile); os.IsNotExist(err) {
		return nil // 历史文件不存在是正常的
	}

	data, err := os.ReadFile(historyFile)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &am.versionHistory)
}

// saveVersionHistory 保存版本历史
func (am *AdvancedManager) saveVersionHistory() error {
	am.versionMutex.RLock()
	defer am.versionMutex.RUnlock()

	historyFile := filepath.Join(am.versionDir, "history.json")
	data, err := json.MarshalIndent(am.versionHistory, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(historyFile, data, 0644)
}

// saveVersionHistoryUnsafe 保存版本历史（不加锁版本）
func (am *AdvancedManager) saveVersionHistoryUnsafe() error {
	historyFile := filepath.Join(am.versionDir, "history.json")
	data, err := json.MarshalIndent(am.versionHistory, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(historyFile, data, 0644)
}

// recordChange 记录配置变更
func (am *AdvancedManager) recordChange(operation, key string, oldValue, newValue interface{}, user, source string) {
	am.statsMutex.Lock()
	defer am.statsMutex.Unlock()

	change := &ConfigChange{
		ID:        generateID(),
		Timestamp: time.Now(),
		Operation: operation,
		Key:       key,
		OldValue:  oldValue,
		NewValue:  newValue,
		User:      user,
		Source:    source,
	}

	am.changeLog = append(am.changeLog, change)
	am.stats.ChangeCount++
	am.stats.LastModified = time.Now()

	// 限制变更日志大小
	if len(am.changeLog) > 1000 {
		am.changeLog = am.changeLog[len(am.changeLog)-1000:]
	}
}