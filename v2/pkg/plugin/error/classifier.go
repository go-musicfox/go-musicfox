package plugin

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

// ErrorClassifier 错误分类器接口
type ErrorClassifier interface {
	// ClassifyError 分类错误
	ClassifyError(ctx context.Context, err error, pluginID string) (*ErrorClassification, error)
	// TrainClassifier 训练分类器
	TrainClassifier(trainingData []TrainingExample) error
	// AddRule 添加分类规则
	AddRule(rule ClassificationRule) error
	// RemoveRule 移除分类规则
	RemoveRule(ruleID string) error
	// GetRules 获取所有规则
	GetRules() []ClassificationRule
	// UpdateWeights 更新权重
	UpdateWeights(feedback []ClassificationFeedback) error
	// GetAccuracy 获取分类准确率
	GetAccuracy() float64
	// Reset 重置分类器
	Reset() error
}

// ErrorClassification 错误分类结果
type ErrorClassification struct {
	ErrorCode     ErrorCode     `json:"error_code"`     // 错误代码
	ErrorType     ErrorType     `json:"error_type"`     // 错误类型
	Severity      ErrorSeverity `json:"severity"`       // 严重程度
	Category      string        `json:"category"`       // 错误类别
	Subcategory   string        `json:"subcategory"`    // 错误子类别
	Confidence    float64       `json:"confidence"`     // 置信度
	Reason        string        `json:"reason"`         // 分类原因
	Suggestions   []string      `json:"suggestions"`    // 建议
	Tags          []string      `json:"tags"`           // 标签
	Metadata      map[string]interface{} `json:"metadata"` // 元数据
	Timestamp     time.Time     `json:"timestamp"`      // 时间戳
}

// TrainingExample 训练样本
type TrainingExample struct {
	ErrorMessage  string                 `json:"error_message"`  // 错误消息
	PluginID      string                 `json:"plugin_id"`      // 插件ID
	ExpectedCode  ErrorCode              `json:"expected_code"`  // 期望错误代码
	ExpectedType  ErrorType              `json:"expected_type"`  // 期望错误类型
	ExpectedSeverity ErrorSeverity       `json:"expected_severity"` // 期望严重程度
	Context       map[string]interface{} `json:"context"`        // 上下文
	Weight        float64                `json:"weight"`         // 权重
}

// ClassificationRule 分类规则
type ClassificationRule struct {
	ID          string                 `json:"id"`          // 规则ID
	Name        string                 `json:"name"`        // 规则名称
	Description string                 `json:"description"` // 描述
	Conditions  []RuleCondition        `json:"conditions"`  // 条件
	Action      ClassificationAction   `json:"action"`      // 动作
	Priority    int                    `json:"priority"`    // 优先级
	Enabled     bool                   `json:"enabled"`     // 是否启用
	Weight      float64                `json:"weight"`      // 权重
	Metadata    map[string]interface{} `json:"metadata"`    // 元数据
	CreatedAt   time.Time              `json:"created_at"`  // 创建时间
	UpdatedAt   time.Time              `json:"updated_at"`  // 更新时间
}

// RuleCondition 规则条件
type RuleCondition struct {
	Field    string      `json:"field"`    // 字段名
	Operator string      `json:"operator"` // 操作符
	Value    interface{} `json:"value"`    // 值
	Regex    string      `json:"regex"`    // 正则表达式
}

// ClassificationAction 分类动作
type ClassificationAction struct {
	ErrorCode   ErrorCode     `json:"error_code"`   // 错误代码
	ErrorType   ErrorType     `json:"error_type"`   // 错误类型
	Severity    ErrorSeverity `json:"severity"`     // 严重程度
	Category    string        `json:"category"`     // 类别
	Subcategory string        `json:"subcategory"`  // 子类别
	Tags        []string      `json:"tags"`         // 标签
	Suggestions []string      `json:"suggestions"`  // 建议
}

// ClassificationFeedback 分类反馈
type ClassificationFeedback struct {
	ErrorMessage     string        `json:"error_message"`     // 错误消息
	PluginID         string        `json:"plugin_id"`         // 插件ID
	PredictedCode    ErrorCode     `json:"predicted_code"`    // 预测错误代码
	ActualCode       ErrorCode     `json:"actual_code"`       // 实际错误代码
	PredictedType    ErrorType     `json:"predicted_type"`    // 预测错误类型
	ActualType       ErrorType     `json:"actual_type"`       // 实际错误类型
	PredictedSeverity ErrorSeverity `json:"predicted_severity"` // 预测严重程度
	ActualSeverity   ErrorSeverity `json:"actual_severity"`   // 实际严重程度
	Correct          bool          `json:"correct"`           // 是否正确
	Confidence       float64       `json:"confidence"`        // 置信度
	Timestamp        time.Time     `json:"timestamp"`         // 时间戳
}

// HybridErrorClassifier 混合错误分类器
type HybridErrorClassifier struct {
	rules           []ClassificationRule
	patterns        map[string]*regexp.Regexp
	featureWeights  map[string]float64
	accuracyHistory []float64
	mutex           sync.RWMutex
	logger          Logger
	metrics         MetricsCollector
}

// NewErrorClassifier 创建新的错误分类器
func NewErrorClassifier(logger Logger, metrics MetricsCollector) ErrorClassifier {
	return &HybridErrorClassifier{
		rules:          make([]ClassificationRule, 0),
		patterns:       make(map[string]*regexp.Regexp),
		featureWeights: make(map[string]float64),
		logger:         logger,
		metrics:        metrics,
	}
}

// ClassifyError 分类错误
func (c *HybridErrorClassifier) ClassifyError(ctx context.Context, err error, pluginID string) (*ErrorClassification, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	errorMessage := err.Error()
	
	// 提取特征
	features := c.extractFeatures(errorMessage, pluginID)
	
	// 基于规则的分类
	ruleResult := c.classifyByRules(errorMessage, pluginID, features)
	
	// 基于模式的分类
	patternResult := c.classifyByPatterns(errorMessage, pluginID)
	
	// 基于特征权重的分类
	weightResult := c.classifyByWeights(features)
	
	// 合并结果
	finalResult := c.mergeClassificationResults(ruleResult, patternResult, weightResult)
	
	// 记录指标
	if c.metrics != nil {
		c.metrics.IncrementCounter("error_classification_total", map[string]string{
			"plugin_id": pluginID,
			"error_type": finalResult.ErrorType.String(),
		})
		c.metrics.RecordHistogram("classification_confidence", finalResult.Confidence, map[string]string{
			"plugin_id": pluginID,
		})
	}
	
	// 记录日志
	if c.logger != nil {
		c.logger.Debug("Error classified", map[string]interface{}{
			"plugin_id":   pluginID,
			"error_code":  finalResult.ErrorCode.String(),
			"error_type":  finalResult.ErrorType.String(),
			"severity":    finalResult.Severity.String(),
			"confidence":  finalResult.Confidence,
			"category":    finalResult.Category,
		})
	}
	
	return finalResult, nil
}

// TrainClassifier 训练分类器
func (c *HybridErrorClassifier) TrainClassifier(trainingData []TrainingExample) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	if len(trainingData) == 0 {
		return fmt.Errorf("training data is empty")
	}
	
	// 更新特征权重
	featureFreq := make(map[string]map[ErrorCode]int)
	totalSamples := len(trainingData)
	
	for _, example := range trainingData {
		features := c.extractFeatures(example.ErrorMessage, example.PluginID)
		
		for feature := range features {
			if _, exists := featureFreq[feature]; !exists {
				featureFreq[feature] = make(map[ErrorCode]int)
			}
			featureFreq[feature][example.ExpectedCode]++
		}
	}
	
	// 计算特征权重（基于信息增益）
	for feature, codeFreq := range featureFreq {
		entropy := c.calculateEntropy(codeFreq, totalSamples)
		c.featureWeights[feature] = entropy
	}
	
	// 生成模式规则
	c.generatePatternRules(trainingData)
	
	if c.logger != nil {
		c.logger.Info("Classifier trained", map[string]interface{}{
			"training_samples": totalSamples,
			"features_count":   len(c.featureWeights),
			"patterns_count":   len(c.patterns),
		})
	}
	
	return nil
}

// AddRule 添加分类规则
func (c *HybridErrorClassifier) AddRule(rule ClassificationRule) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	// 检查规则ID是否已存在
	for _, existingRule := range c.rules {
		if existingRule.ID == rule.ID {
			return fmt.Errorf("rule with ID %s already exists", rule.ID)
		}
	}
	
	// 编译正则表达式
	for _, condition := range rule.Conditions {
		if condition.Regex != "" {
			if _, err := regexp.Compile(condition.Regex); err != nil {
				return fmt.Errorf("invalid regex in condition: %w", err)
			}
		}
	}
	
	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()
	
	// 按优先级插入
	inserted := false
	for i, existingRule := range c.rules {
		if rule.Priority > existingRule.Priority {
			c.rules = append(c.rules[:i], append([]ClassificationRule{rule}, c.rules[i:]...)...)
			inserted = true
			break
		}
	}
	
	if !inserted {
		c.rules = append(c.rules, rule)
	}
	
	if c.logger != nil {
		c.logger.Info("Classification rule added", map[string]interface{}{
			"rule_id":   rule.ID,
			"rule_name": rule.Name,
			"priority":  rule.Priority,
		})
	}
	
	return nil
}

// RemoveRule 移除分类规则
func (c *HybridErrorClassifier) RemoveRule(ruleID string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	for i, rule := range c.rules {
		if rule.ID == ruleID {
			c.rules = append(c.rules[:i], c.rules[i+1:]...)
			
			if c.logger != nil {
				c.logger.Info("Classification rule removed", map[string]interface{}{
					"rule_id": ruleID,
				})
			}
			
			return nil
		}
	}
	
	return fmt.Errorf("rule with ID %s not found", ruleID)
}

// GetRules 获取所有规则
func (c *HybridErrorClassifier) GetRules() []ClassificationRule {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	rules := make([]ClassificationRule, len(c.rules))
	copy(rules, c.rules)
	return rules
}

// UpdateWeights 更新权重
func (c *HybridErrorClassifier) UpdateWeights(feedback []ClassificationFeedback) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	if len(feedback) == 0 {
		return nil
	}
	
	// 计算准确率
	correctCount := 0
	for _, fb := range feedback {
		if fb.Correct {
			correctCount++
		}
	}
	
	accuracy := float64(correctCount) / float64(len(feedback))
	c.accuracyHistory = append(c.accuracyHistory, accuracy)
	
	// 保持历史记录在合理范围内
	if len(c.accuracyHistory) > 100 {
		c.accuracyHistory = c.accuracyHistory[1:]
	}
	
	// 基于反馈调整特征权重
	for _, fb := range feedback {
		features := c.extractFeatures(fb.ErrorMessage, fb.PluginID)
		
		for feature := range features {
			if fb.Correct {
				// 正确分类，增加权重
				c.featureWeights[feature] *= 1.1
			} else {
				// 错误分类，减少权重
				c.featureWeights[feature] *= 0.9
			}
			
			// 限制权重范围
			if c.featureWeights[feature] > 10.0 {
				c.featureWeights[feature] = 10.0
			} else if c.featureWeights[feature] < 0.1 {
				c.featureWeights[feature] = 0.1
			}
		}
	}
	
	if c.logger != nil {
		c.logger.Info("Classifier weights updated", map[string]interface{}{
			"feedback_count": len(feedback),
			"accuracy":       accuracy,
			"correct_count":  correctCount,
		})
	}
	
	return nil
}

// GetAccuracy 获取分类准确率
func (c *HybridErrorClassifier) GetAccuracy() float64 {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	if len(c.accuracyHistory) == 0 {
		return 0.0
	}
	
	// 返回最近的准确率
	return c.accuracyHistory[len(c.accuracyHistory)-1]
}

// Reset 重置分类器
func (c *HybridErrorClassifier) Reset() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	c.rules = make([]ClassificationRule, 0)
	c.patterns = make(map[string]*regexp.Regexp)
	c.featureWeights = make(map[string]float64)
	c.accuracyHistory = make([]float64, 0)
	
	if c.logger != nil {
		c.logger.Info("Classifier reset", nil)
	}
	
	return nil
}

// extractFeatures 提取特征
func (c *HybridErrorClassifier) extractFeatures(errorMessage, pluginID string) map[string]float64 {
	features := make(map[string]float64)
	
	// 插件ID特征
	features["plugin_"+pluginID] = 1.0
	
	// 长度特征
	msgLen := len(errorMessage)
	if msgLen < 50 {
		features["length_short"] = 1.0
	} else if msgLen < 200 {
		features["length_medium"] = 1.0
	} else {
		features["length_long"] = 1.0
	}
	
	// 关键词特征
	keywords := []string{
		"timeout", "connection", "network", "permission", "denied",
		"not found", "invalid", "failed", "error", "exception",
		"null", "undefined", "memory", "cpu", "disk",
		"authentication", "authorization", "config", "format",
	}
	
	lowerMsg := strings.ToLower(errorMessage)
	for _, keyword := range keywords {
		if strings.Contains(lowerMsg, keyword) {
			features["keyword_"+keyword] = 1.0
		}
	}
	
	// 数字特征
	numberRegex := regexp.MustCompile(`\d+`)
	numbers := numberRegex.FindAllString(errorMessage, -1)
	if len(numbers) > 0 {
		features["has_numbers"] = float64(len(numbers))
	}
	
	// 路径特征
	pathRegex := regexp.MustCompile(`[/\\][\w\-\.]+`)
	if pathRegex.MatchString(errorMessage) {
		features["has_path"] = 1.0
	}
	
	// URL特征
	urlRegex := regexp.MustCompile(`https?://[\w\-\./]+`)
	if urlRegex.MatchString(errorMessage) {
		features["has_url"] = 1.0
	}
	
	return features
}

// classifyByRules 基于规则分类
func (c *HybridErrorClassifier) classifyByRules(errorMessage, pluginID string, features map[string]float64) *ErrorClassification {
	for _, rule := range c.rules {
		if !rule.Enabled {
			continue
		}
		
		if c.matchesRule(rule, errorMessage, pluginID, features) {
			return &ErrorClassification{
				ErrorCode:   rule.Action.ErrorCode,
				ErrorType:   rule.Action.ErrorType,
				Severity:    rule.Action.Severity,
				Category:    rule.Action.Category,
				Subcategory: rule.Action.Subcategory,
				Confidence:  rule.Weight,
				Reason:      fmt.Sprintf("Matched rule: %s", rule.Name),
				Suggestions: rule.Action.Suggestions,
				Tags:        rule.Action.Tags,
				Timestamp:   time.Now(),
			}
		}
	}
	
	return c.getDefaultClassification("No matching rules")
}

// classifyByPatterns 基于模式分类
func (c *HybridErrorClassifier) classifyByPatterns(errorMessage, pluginID string) *ErrorClassification {
	for pattern, regex := range c.patterns {
		if regex.MatchString(errorMessage) {
			// 基于模式名称推断分类
			return c.inferClassificationFromPattern(pattern)
		}
	}
	
	return c.getDefaultClassification("No matching patterns")
}

// classifyByWeights 基于权重分类
func (c *HybridErrorClassifier) classifyByWeights(features map[string]float64) *ErrorClassification {
	scores := make(map[ErrorCode]float64)
	
	for feature, value := range features {
		if weight, exists := c.featureWeights[feature]; exists {
			// 简化的评分计算
			for code := ErrorCodeUnknown; code <= ErrorCodeUIResourceNotFound; code++ {
				scores[code] += value * weight
			}
		}
	}
	
	// 找到最高分的错误代码
	var bestCode ErrorCode
	var bestScore float64
	for code, score := range scores {
		if score > bestScore {
			bestCode = code
			bestScore = score
		}
	}
	
	if bestScore > 0 {
		return &ErrorClassification{
			ErrorCode:  bestCode,
			ErrorType:  getErrorTypeByCode(bestCode),
			Severity:   getErrorSeverityByCode(bestCode),
			Confidence: math.Min(bestScore/10.0, 1.0), // 归一化到0-1
			Reason:     "Weight-based classification",
			Timestamp:  time.Now(),
		}
	}
	
	return c.getDefaultClassification("Insufficient feature weights")
}

// mergeClassificationResults 合并分类结果
func (c *HybridErrorClassifier) mergeClassificationResults(results ...*ErrorClassification) *ErrorClassification {
	// 按置信度排序
	sort.Slice(results, func(i, j int) bool {
		return results[i].Confidence > results[j].Confidence
	})
	
	// 返回置信度最高的结果
	if len(results) > 0 && results[0].Confidence > 0 {
		return results[0]
	}
	
	return c.getDefaultClassification("No confident classification")
}

// matchesRule 检查是否匹配规则
func (c *HybridErrorClassifier) matchesRule(rule ClassificationRule, errorMessage, pluginID string, features map[string]float64) bool {
	for _, condition := range rule.Conditions {
		if !c.evaluateCondition(condition, errorMessage, pluginID, features) {
			return false
		}
	}
	return true
}

// evaluateCondition 评估条件
func (c *HybridErrorClassifier) evaluateCondition(condition RuleCondition, errorMessage, pluginID string, features map[string]float64) bool {
	switch condition.Field {
	case "message":
		return c.evaluateStringCondition(condition, errorMessage)
	case "plugin_id":
		return c.evaluateStringCondition(condition, pluginID)
	case "feature":
		if featureName, ok := condition.Value.(string); ok {
			_, exists := features[featureName]
			return exists
		}
	default:
		return false
	}
	return false
}

// evaluateStringCondition 评估字符串条件
func (c *HybridErrorClassifier) evaluateStringCondition(condition RuleCondition, value string) bool {
	switch condition.Operator {
	case "contains":
		if str, ok := condition.Value.(string); ok {
			return strings.Contains(strings.ToLower(value), strings.ToLower(str))
		}
	case "equals":
		if str, ok := condition.Value.(string); ok {
			return strings.EqualFold(value, str)
		}
	case "regex":
		if condition.Regex != "" {
			if regex, err := regexp.Compile(condition.Regex); err == nil {
				return regex.MatchString(value)
			}
		}
	case "starts_with":
		if str, ok := condition.Value.(string); ok {
			return strings.HasPrefix(strings.ToLower(value), strings.ToLower(str))
		}
	case "ends_with":
		if str, ok := condition.Value.(string); ok {
			return strings.HasSuffix(strings.ToLower(value), strings.ToLower(str))
		}
	}
	return false
}

// calculateEntropy 计算熵
func (c *HybridErrorClassifier) calculateEntropy(codeFreq map[ErrorCode]int, totalSamples int) float64 {
	if totalSamples == 0 {
		return 0.0
	}
	
	entropy := 0.0
	for _, freq := range codeFreq {
		if freq > 0 {
			p := float64(freq) / float64(totalSamples)
			entropy -= p * math.Log2(p)
		}
	}
	
	return entropy
}

// generatePatternRules 生成模式规则
func (c *HybridErrorClassifier) generatePatternRules(trainingData []TrainingExample) {
	// 简化的模式生成
	patternMap := make(map[string][]TrainingExample)
	
	for _, example := range trainingData {
		// 提取常见模式
		patterns := c.extractPatterns(example.ErrorMessage)
		for _, pattern := range patterns {
			patternMap[pattern] = append(patternMap[pattern], example)
		}
	}
	
	// 为频繁出现的模式创建正则表达式
	for pattern, examples := range patternMap {
		if len(examples) >= 3 { // 至少出现3次
			if regex, err := regexp.Compile(pattern); err == nil {
				c.patterns[pattern] = regex
			}
		}
	}
}

// extractPatterns 提取模式
func (c *HybridErrorClassifier) extractPatterns(errorMessage string) []string {
	var patterns []string
	
	// 数字模式
	numberPattern := regexp.MustCompile(`\d+`)
	generalizedMsg := numberPattern.ReplaceAllString(errorMessage, "\\d+")
	patterns = append(patterns, generalizedMsg)
	
	// 路径模式
	pathPattern := regexp.MustCompile(`[/\\][\w\-\.]+`)
	generalizedMsg = pathPattern.ReplaceAllString(generalizedMsg, "[/\\\\][\\w\\-\\.]+")
	patterns = append(patterns, generalizedMsg)
	
	return patterns
}

// inferClassificationFromPattern 从模式推断分类
func (c *HybridErrorClassifier) inferClassificationFromPattern(pattern string) *ErrorClassification {
	// 基于模式内容推断错误类型
	lowerPattern := strings.ToLower(pattern)
	
	if strings.Contains(lowerPattern, "timeout") {
		return &ErrorClassification{
			ErrorCode:  ErrorCodePluginTimeout,
			ErrorType:  ErrorTypeTimeout,
			Severity:   ErrorSeverityError,
			Category:   "timeout",
			Confidence: 0.7,
			Reason:     "Pattern-based timeout detection",
			Timestamp:  time.Now(),
		}
	} else if strings.Contains(lowerPattern, "network") || strings.Contains(lowerPattern, "connection") {
		return &ErrorClassification{
			ErrorCode:  ErrorCodePluginNetworkError,
			ErrorType:  ErrorTypeNetwork,
			Severity:   ErrorSeverityError,
			Category:   "network",
			Confidence: 0.7,
			Reason:     "Pattern-based network detection",
			Timestamp:  time.Now(),
		}
	}
	
	return c.getDefaultClassification("Pattern-based fallback")
}

// getDefaultClassification 获取默认分类
func (c *HybridErrorClassifier) getDefaultClassification(reason string) *ErrorClassification {
	return &ErrorClassification{
		ErrorCode:  ErrorCodeUnknown,
		ErrorType:  ErrorTypeSystem,
		Severity:   ErrorSeverityError,
		Category:   "unknown",
		Confidence: 0.1,
		Reason:     reason,
		Timestamp:  time.Now(),
	}
}