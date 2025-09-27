package config

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// SetEncryptionKey 设置加密密钥
func (am *AdvancedManager) SetEncryptionKey(key []byte) error {
	am.encryptionMutex.Lock()
	defer am.encryptionMutex.Unlock()

	if len(key) != 32 {
		return fmt.Errorf("encryption key must be 32 bytes long")
	}

	am.encryptionKey = make([]byte, len(key))
	copy(am.encryptionKey, key)

	return nil
}

// GenerateEncryptionKey 生成新的加密密钥
func (am *AdvancedManager) GenerateEncryptionKey() ([]byte, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("failed to generate encryption key: %w", err)
	}
	return key, nil
}

// EncryptSensitiveData 加密敏感数据
func (am *AdvancedManager) EncryptSensitiveData(keys []string) error {
	am.encryptionMutex.Lock()
	defer am.encryptionMutex.Unlock()

	if len(am.encryptionKey) == 0 {
		return fmt.Errorf("encryption key not set")
	}

	for _, key := range keys {
		value := am.k.Get(key)
		if value == nil {
			continue
		}

		// 如果已经加密，跳过
		if am.encryptedKeys[key] {
			continue
		}

		// 序列化值
		valueBytes, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal value for key %s: %w", key, err)
		}

		// 加密数据
		encryptedData, err := am.encryptData(valueBytes)
		if err != nil {
			return fmt.Errorf("failed to encrypt data for key %s: %w", key, err)
		}

		// 编码为base64
		encryptedValue := base64.StdEncoding.EncodeToString(encryptedData)

		// 添加加密标记前缀
		encryptedValueWithPrefix := fmt.Sprintf("__ENCRYPTED__%s", encryptedValue)

		// 更新配置
		am.k.Set(key, encryptedValueWithPrefix)
		am.encryptedKeys[key] = true

		// 记录变更
		am.recordChange("encrypt", key, value, "[ENCRYPTED]", "system", "encryption")
	}

	return nil
}

// DecryptSensitiveData 解密敏感数据
func (am *AdvancedManager) DecryptSensitiveData(keys []string) error {
	am.encryptionMutex.Lock()
	defer am.encryptionMutex.Unlock()

	if len(am.encryptionKey) == 0 {
		return fmt.Errorf("encryption key not set")
	}

	for _, key := range keys {
		value := am.k.Get(key)
		if value == nil {
			continue
		}

		// 检查是否是加密数据
		valueStr, ok := value.(string)
		if !ok || !strings.HasPrefix(valueStr, "__ENCRYPTED__") {
			continue
		}

		// 移除加密标记前缀
		encryptedValue := strings.TrimPrefix(valueStr, "__ENCRYPTED__")

		// 解码base64
		encryptedData, err := base64.StdEncoding.DecodeString(encryptedValue)
		if err != nil {
			return fmt.Errorf("failed to decode encrypted data for key %s: %w", key, err)
		}

		// 解密数据
		decryptedData, err := am.decryptData(encryptedData)
		if err != nil {
			return fmt.Errorf("failed to decrypt data for key %s: %w", key, err)
		}

		// 反序列化值
		var decryptedValue interface{}
		if err := json.Unmarshal(decryptedData, &decryptedValue); err != nil {
			return fmt.Errorf("failed to unmarshal decrypted value for key %s: %w", key, err)
		}

		// 更新配置
		am.k.Set(key, decryptedValue)
		am.encryptedKeys[key] = false

		// 记录变更
		am.recordChange("decrypt", key, "[ENCRYPTED]", decryptedValue, "system", "decryption")
	}

	return nil
}

// IsEncrypted 检查配置项是否已加密
func (am *AdvancedManager) IsEncrypted(key string) bool {
	am.encryptionMutex.RLock()
	defer am.encryptionMutex.RUnlock()

	return am.encryptedKeys[key]
}

// SetAccessControl 设置访问控制规则
func (am *AdvancedManager) SetAccessControl(rules *AccessControlRules) error {
	am.accessMutex.Lock()
	defer am.accessMutex.Unlock()

	// 验证访问控制规则
	if err := am.validateAccessRules(rules); err != nil {
		return fmt.Errorf("invalid access control rules: %w", err)
	}

	am.accessRules = rules
	return nil
}

// CheckAccess 检查访问权限
func (am *AdvancedManager) CheckAccess(operation string, key string, user string) bool {
	am.accessMutex.RLock()
	defer am.accessMutex.RUnlock()

	if am.accessRules == nil {
		return true // 没有设置访问控制规则，默认允许
	}

	// 获取用户权限
	userPermission, userExists := am.accessRules.Users[user]

	// 检查具体规则
	for _, rule := range am.accessRules.Rules {
		if am.matchesPattern(key, rule.Pattern) {
			// 检查操作是否匹配
			if !am.containsString(rule.Operations, operation) {
				continue
			}

			// 检查用户是否在允许列表中
			if am.containsString(rule.Users, user) {
				return rule.Policy == "allow"
			}

			// 检查用户角色是否在允许列表中
			if userExists {
				for _, userRole := range userPermission.Roles {
					if am.containsString(rule.Roles, userRole) {
						return rule.Policy == "allow"
					}
				}
			}

			// 如果规则匹配但用户不在允许列表中
			if rule.Policy == "deny" {
				return false
			}
		}
	}

	// 应用默认策略
	return am.accessRules.DefaultPolicy == "allow"
}

// validateAccessRules 验证访问控制规则
func (am *AdvancedManager) validateAccessRules(rules *AccessControlRules) error {
	if rules.DefaultPolicy != "allow" && rules.DefaultPolicy != "deny" {
		return fmt.Errorf("default policy must be 'allow' or 'deny'")
	}

	for i, rule := range rules.Rules {
		if rule.Pattern == "" {
			return fmt.Errorf("rule %d: pattern cannot be empty", i)
		}

		if rule.Policy != "allow" && rule.Policy != "deny" {
			return fmt.Errorf("rule %d: policy must be 'allow' or 'deny'", i)
		}

		validOperations := []string{"read", "write", "delete"}
		for _, op := range rule.Operations {
			if !am.containsString(validOperations, op) {
				return fmt.Errorf("rule %d: invalid operation '%s'", i, op)
			}
		}
	}

	return nil
}

// matchesPattern 检查键是否匹配模式
func (am *AdvancedManager) matchesPattern(key, pattern string) bool {
	// 支持通配符模式
	if pattern == "*" {
		return true
	}

	// 支持前缀匹配
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(key, prefix)
	}

	// 支持后缀匹配
	if strings.HasPrefix(pattern, "*") {
		suffix := strings.TrimPrefix(pattern, "*")
		return strings.HasSuffix(key, suffix)
	}

	// 支持正则表达式
	if strings.HasPrefix(pattern, "regex:") {
		regexPattern := strings.TrimPrefix(pattern, "regex:")
		matched, err := regexp.MatchString(regexPattern, key)
		if err != nil {
			return false
		}
		return matched
	}

	// 精确匹配
	return key == pattern
}

// containsString 检查字符串切片是否包含指定字符串
func (am *AdvancedManager) containsString(slice []string, str string) bool {
	for _, s := range slice {
		if s == str || s == "*" {
			return true
		}
	}
	return false
}

// GetSecurityInfo 获取安全信息
func (am *AdvancedManager) GetSecurityInfo() *SecurityInfo {
	am.encryptionMutex.RLock()
	am.accessMutex.RLock()
	defer am.encryptionMutex.RUnlock()
	defer am.accessMutex.RUnlock()

	encryptedCount := 0
	for _, encrypted := range am.encryptedKeys {
		if encrypted {
			encryptedCount++
		}
	}

	info := &SecurityInfo{
		EncryptionEnabled:  len(am.encryptionKey) > 0,
		EncryptedKeyCount:  encryptedCount,
		AccessControlEnabled: am.accessRules != nil,
		LastSecurityUpdate:   time.Now(),
	}

	if am.accessRules != nil {
		info.AccessRuleCount = len(am.accessRules.Rules)
		info.UserCount = len(am.accessRules.Users)
		info.RoleCount = len(am.accessRules.Roles)
	}

	return info
}

// SecurityInfo 安全信息
type SecurityInfo struct {
	EncryptionEnabled     bool      `json:"encryption_enabled"`
	EncryptedKeyCount     int       `json:"encrypted_key_count"`
	AccessControlEnabled  bool      `json:"access_control_enabled"`
	AccessRuleCount       int       `json:"access_rule_count"`
	UserCount             int       `json:"user_count"`
	RoleCount             int       `json:"role_count"`
	LastSecurityUpdate    time.Time `json:"last_security_update"`
}

// decryptSensitiveDataUnsafe 解密敏感数据（不加锁版本）
func (am *AdvancedManager) decryptSensitiveDataUnsafe(keys []string) error {
	if len(am.encryptionKey) == 0 {
		return fmt.Errorf("encryption key not set")
	}

	for _, key := range keys {
		value := am.k.Get(key)
		if value == nil {
			continue
		}

		// 检查是否是加密数据
		valueStr, ok := value.(string)
		if !ok || !strings.HasPrefix(valueStr, "__ENCRYPTED__") {
			continue
		}

		// 解码base64
		encryptedData := strings.TrimPrefix(valueStr, "__ENCRYPTED__")
		ciphertext, err := base64.StdEncoding.DecodeString(encryptedData)
		if err != nil {
			return fmt.Errorf("failed to decode encrypted data for key %s: %w", key, err)
		}

		// 解密数据
		plaintext, err := am.decryptData(ciphertext)
		if err != nil {
			return fmt.Errorf("failed to decrypt data for key %s: %w", key, err)
		}

		// 设置解密后的值
		am.k.Set(key, string(plaintext))
		am.encryptedKeys[key] = false
	}

	return nil
}

// encryptSensitiveDataUnsafe 加密敏感数据（不加锁版本）
func (am *AdvancedManager) encryptSensitiveDataUnsafe(keys []string) error {
	if len(am.encryptionKey) == 0 {
		return fmt.Errorf("encryption key not set")
	}

	for _, key := range keys {
		value := am.k.Get(key)
		if value == nil {
			continue
		}

		// 检查是否已经加密
		if am.encryptedKeys[key] {
			continue
		}

		// 转换为字符串
		valueStr := fmt.Sprintf("%v", value)

		// 加密数据
		ciphertext, err := am.encryptData([]byte(valueStr))
		if err != nil {
			return fmt.Errorf("failed to encrypt data for key %s: %w", key, err)
		}

		// 编码为base64并添加前缀
		encryptedValue := "__ENCRYPTED__" + base64.StdEncoding.EncodeToString(ciphertext)
		am.k.Set(key, encryptedValue)
		am.encryptedKeys[key] = true
	}

	return nil
}

// RotateEncryptionKey 轮换加密密钥
func (am *AdvancedManager) RotateEncryptionKey() error {
	am.encryptionMutex.Lock()
	defer am.encryptionMutex.Unlock()

	if len(am.encryptionKey) == 0 {
		return fmt.Errorf("no encryption key set")
	}

	// 生成新密钥
	newKey, err := am.GenerateEncryptionKey()
	if err != nil {
		return fmt.Errorf("failed to generate new encryption key: %w", err)
	}

	// 收集所有加密的键
	encryptedKeys := make([]string, 0)
	for key, encrypted := range am.encryptedKeys {
		if encrypted {
			encryptedKeys = append(encryptedKeys, key)
		}
	}

	// 使用旧密钥解密所有数据（使用不加锁版本）
	if err := am.decryptSensitiveDataUnsafe(encryptedKeys); err != nil {
		return fmt.Errorf("failed to decrypt data with old key: %w", err)
	}

	// 设置新密钥
	am.encryptionKey = newKey

	// 使用新密钥重新加密所有数据（使用不加锁版本）
	if err := am.encryptSensitiveDataUnsafe(encryptedKeys); err != nil {
		return fmt.Errorf("failed to encrypt data with new key: %w", err)
	}

	// 记录密钥轮换
	am.recordChange("key_rotation", "*", "[OLD_KEY]", "[NEW_KEY]", "system", "key_rotation")

	return nil
}

// CreateSecurityAudit 创建安全审计报告
func (am *AdvancedManager) CreateSecurityAudit() *SecurityAudit {
	am.statsMutex.RLock()
	am.encryptionMutex.RLock()
	am.accessMutex.RLock()
	defer am.statsMutex.RUnlock()
	defer am.encryptionMutex.RUnlock()
	defer am.accessMutex.RUnlock()

	audit := &SecurityAudit{
		Timestamp:     time.Now(),
		SecurityInfo:  am.GetSecurityInfo(),
		Vulnerabilities: make([]string, 0),
		Recommendations: make([]string, 0),
	}

	// 检查安全漏洞
	if !audit.SecurityInfo.EncryptionEnabled {
		audit.Vulnerabilities = append(audit.Vulnerabilities, "Encryption is not enabled")
		audit.Recommendations = append(audit.Recommendations, "Enable encryption for sensitive data")
	}

	if !audit.SecurityInfo.AccessControlEnabled {
		audit.Vulnerabilities = append(audit.Vulnerabilities, "Access control is not configured")
		audit.Recommendations = append(audit.Recommendations, "Configure access control rules")
	}

	// 检查敏感配置项
	sensitivePatterns := []string{
		"password", "secret", "key", "token", "credential",
	}

	allKeys := am.k.Keys()
	for _, key := range allKeys {
		for _, pattern := range sensitivePatterns {
			if strings.Contains(strings.ToLower(key), pattern) {
				if !am.encryptedKeys[key] {
					audit.Vulnerabilities = append(audit.Vulnerabilities, 
						fmt.Sprintf("Sensitive key '%s' is not encrypted", key))
					audit.Recommendations = append(audit.Recommendations, 
						fmt.Sprintf("Encrypt sensitive key '%s'", key))
				}
				break
			}
		}
	}

	// 计算安全评分
	audit.SecurityScore = am.calculateSecurityScore(audit)

	return audit
}

// SecurityAudit 安全审计报告
type SecurityAudit struct {
	Timestamp       time.Time     `json:"timestamp"`
	SecurityInfo    *SecurityInfo `json:"security_info"`
	Vulnerabilities []string      `json:"vulnerabilities"`
	Recommendations []string      `json:"recommendations"`
	SecurityScore   int           `json:"security_score"` // 0-100
}

// calculateSecurityScore 计算安全评分
func (am *AdvancedManager) calculateSecurityScore(audit *SecurityAudit) int {
	score := 100

	// 每个漏洞扣分
	score -= len(audit.Vulnerabilities) * 10

	// 没有启用加密扣分
	if !audit.SecurityInfo.EncryptionEnabled {
		score -= 20
	}

	// 没有启用访问控制扣分
	if !audit.SecurityInfo.AccessControlEnabled {
		score -= 15
	}

	// 确保分数不低于0
	if score < 0 {
		score = 0
	}

	return score
}