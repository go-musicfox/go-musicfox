package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// RPCClient RPC客户端
type RPCClient struct {
	httpClient    *http.Client
	baseURL       string
	apiKey        string
	headers       map[string]string
	retryConfig   *RetryConfig
	rateLimit     *RateLimit
	mu            sync.RWMutex
	interceptors  []RequestInterceptor
}

// RetryConfig 重试配置
type RetryConfig struct {
	MaxRetries    int           `json:"max_retries"`
	InitialDelay  time.Duration `json:"initial_delay"`
	MaxDelay      time.Duration `json:"max_delay"`
	BackoffFactor float64       `json:"backoff_factor"`
	RetryableErrors []int       `json:"retryable_errors"`
}

// RateLimit 限流配置
type RateLimit struct {
	RequestsPerSecond int           `json:"requests_per_second"`
	BurstSize         int           `json:"burst_size"`
	lastRequest       time.Time
	tokens            int
	mu                sync.Mutex
}

// RequestInterceptor 请求拦截器
type RequestInterceptor func(*http.Request) error

// RPCRequest RPC请求
type RPCRequest struct {
	Method   string                 `json:"method"`
	URL      string                 `json:"url"`
	Headers  map[string]string      `json:"headers"`
	Body     interface{}            `json:"body"`
	Params   map[string]string      `json:"params"`
	Timeout  time.Duration          `json:"timeout"`
	Metadata map[string]interface{} `json:"metadata"`
}

// RPCResponse RPC响应
type RPCResponse struct {
	StatusCode int                    `json:"status_code"`
	Headers    map[string][]string    `json:"headers"`
	Body       []byte                 `json:"body"`
	Error      string                 `json:"error"`
	Duration   time.Duration          `json:"duration"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// NewRPCClient 创建RPC客户端
func NewRPCClient(baseURL string, apiKey string) *RPCClient {
	return &RPCClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		baseURL: baseURL,
		apiKey:  apiKey,
		headers: make(map[string]string),
		retryConfig: &RetryConfig{
			MaxRetries:      3,
			InitialDelay:    1 * time.Second,
			MaxDelay:        30 * time.Second,
			BackoffFactor:   2.0,
			RetryableErrors: []int{500, 502, 503, 504, 429},
		},
		rateLimit: &RateLimit{
			RequestsPerSecond: 10,
			BurstSize:         20,
			tokens:            20,
		},
		interceptors: make([]RequestInterceptor, 0),
	}
}

// SetTimeout 设置超时时间
func (c *RPCClient) SetTimeout(timeout time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.httpClient.Timeout = timeout
}

// SetHeader 设置请求头
func (c *RPCClient) SetHeader(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.headers[key] = value
}

// AddInterceptor 添加请求拦截器
func (c *RPCClient) AddInterceptor(interceptor RequestInterceptor) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.interceptors = append(c.interceptors, interceptor)
}

// SetRetryConfig 设置重试配置
func (c *RPCClient) SetRetryConfig(config *RetryConfig) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.retryConfig = config
}

// SetRateLimit 设置限流配置
func (c *RPCClient) SetRateLimit(rateLimit *RateLimit) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.rateLimit = rateLimit
}

// Get 发送GET请求
func (c *RPCClient) Get(ctx context.Context, path string, params map[string]string) (*RPCResponse, error) {
	request := &RPCRequest{
		Method:  "GET",
		URL:     c.buildURL(path),
		Params:  params,
		Timeout: c.httpClient.Timeout,
	}
	return c.Do(ctx, request)
}

// Post 发送POST请求
func (c *RPCClient) Post(ctx context.Context, path string, body interface{}) (*RPCResponse, error) {
	request := &RPCRequest{
		Method:  "POST",
		URL:     c.buildURL(path),
		Body:    body,
		Timeout: c.httpClient.Timeout,
	}
	return c.Do(ctx, request)
}

// Put 发送PUT请求
func (c *RPCClient) Put(ctx context.Context, path string, body interface{}) (*RPCResponse, error) {
	request := &RPCRequest{
		Method:  "PUT",
		URL:     c.buildURL(path),
		Body:    body,
		Timeout: c.httpClient.Timeout,
	}
	return c.Do(ctx, request)
}

// Delete 发送DELETE请求
func (c *RPCClient) Delete(ctx context.Context, path string) (*RPCResponse, error) {
	request := &RPCRequest{
		Method:  "DELETE",
		URL:     c.buildURL(path),
		Timeout: c.httpClient.Timeout,
	}
	return c.Do(ctx, request)
}

// Do 执行RPC请求
func (c *RPCClient) Do(ctx context.Context, rpcReq *RPCRequest) (*RPCResponse, error) {
	start := time.Now()

	// 应用限流
	if err := c.applyRateLimit(); err != nil {
		return nil, err
	}

	// 构建HTTP请求
	req, err := c.buildHTTPRequest(ctx, rpcReq)
	if err != nil {
		return nil, err
	}

	// 应用拦截器
	for _, interceptor := range c.interceptors {
		if err := interceptor(req); err != nil {
			return nil, err
		}
	}

	// 执行请求（带重试）
	resp, err := c.doWithRetry(req)
	if err != nil {
		return &RPCResponse{
			Error:    err.Error(),
			Duration: time.Since(start),
		}, err
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &RPCResponse{
			StatusCode: resp.StatusCode,
			Headers:    resp.Header,
			Error:      err.Error(),
			Duration:   time.Since(start),
		}, err
	}

	return &RPCResponse{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       body,
		Duration:   time.Since(start),
	}, nil
}

// buildURL 构建完整URL
func (c *RPCClient) buildURL(path string) string {
	if c.baseURL == "" {
		return path
	}
	return c.baseURL + path
}

// buildHTTPRequest 构建HTTP请求
func (c *RPCClient) buildHTTPRequest(ctx context.Context, rpcReq *RPCRequest) (*http.Request, error) {
	var body io.Reader

	// 处理请求体
	if rpcReq.Body != nil {
		bodyBytes, err := json.Marshal(rpcReq.Body)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(bodyBytes)
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, rpcReq.Method, rpcReq.URL, body)
	if err != nil {
		return nil, err
	}

	// 设置默认头部
	c.mu.RLock()
	for key, value := range c.headers {
		req.Header.Set(key, value)
	}
	c.mu.RUnlock()

	// 设置API密钥
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	// 设置Content-Type
	if rpcReq.Body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// 设置自定义头部
	for key, value := range rpcReq.Headers {
		req.Header.Set(key, value)
	}

	// 添加查询参数
	if len(rpcReq.Params) > 0 {
		q := req.URL.Query()
		for key, value := range rpcReq.Params {
			q.Add(key, value)
		}
		req.URL.RawQuery = q.Encode()
	}

	return req, nil
}

// doWithRetry 带重试的请求执行
func (c *RPCClient) doWithRetry(req *http.Request) (*http.Response, error) {
	c.mu.RLock()
	retryConfig := c.retryConfig
	c.mu.RUnlock()

	var lastErr error
	delay := retryConfig.InitialDelay

	for attempt := 0; attempt <= retryConfig.MaxRetries; attempt++ {
		if attempt > 0 {
			// 等待重试延迟
			time.Sleep(delay)
			// 指数退避
			delay = time.Duration(float64(delay) * retryConfig.BackoffFactor)
			if delay > retryConfig.MaxDelay {
				delay = retryConfig.MaxDelay
			}
		}

		// 克隆请求（因为Body可能被消费）
		reqClone := c.cloneRequest(req)

		// 执行请求
		resp, err := c.httpClient.Do(reqClone)
		if err != nil {
			lastErr = err
			continue
		}

		// 检查是否需要重试
		if !c.shouldRetry(resp.StatusCode) {
			return resp, nil
		}

		resp.Body.Close()
		lastErr = fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	return nil, lastErr
}

// shouldRetry 判断是否应该重试
func (c *RPCClient) shouldRetry(statusCode int) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, code := range c.retryConfig.RetryableErrors {
		if statusCode == code {
			return true
		}
	}
	return false
}

// cloneRequest 克隆HTTP请求
func (c *RPCClient) cloneRequest(req *http.Request) *http.Request {
	reqClone := req.Clone(req.Context())

	// 如果有Body，需要重新设置
	if req.Body != nil {
		// 这里简化处理，实际应用中可能需要更复杂的Body克隆逻辑
		reqClone.Body = req.Body
	}

	return reqClone
}

// applyRateLimit 应用限流
func (c *RPCClient) applyRateLimit() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.rateLimit == nil {
		return nil
	}

	c.rateLimit.mu.Lock()
	defer c.rateLimit.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(c.rateLimit.lastRequest)

	// 补充令牌
	tokensToAdd := int(elapsed.Seconds() * float64(c.rateLimit.RequestsPerSecond))
	c.rateLimit.tokens += tokensToAdd
	if c.rateLimit.tokens > c.rateLimit.BurstSize {
		c.rateLimit.tokens = c.rateLimit.BurstSize
	}

	// 检查是否有可用令牌
	if c.rateLimit.tokens <= 0 {
		return fmt.Errorf("rate limit exceeded")
	}

	// 消费令牌
	c.rateLimit.tokens--
	c.rateLimit.lastRequest = now

	return nil
}

// MusicSourceRPCAdapter 音乐源RPC适配器
type MusicSourceRPCAdapter struct {
	client     *RPCClient
	sourceName string
	baseURL    string
	apiKey     string
}

// NewMusicSourceRPCAdapter 创建音乐源RPC适配器
func NewMusicSourceRPCAdapter(sourceName, baseURL, apiKey string) *MusicSourceRPCAdapter {
	return &MusicSourceRPCAdapter{
		client:     NewRPCClient(baseURL, apiKey),
		sourceName: sourceName,
		baseURL:    baseURL,
		apiKey:     apiKey,
	}
}

// Search 搜索音乐
func (a *MusicSourceRPCAdapter) Search(ctx context.Context, query string, options SearchOptions) (*SearchResult, error) {
	params := map[string]string{
		"q":      query,
		"type":   options.Type.String(),
		"limit":  fmt.Sprintf("%d", options.Limit),
		"offset": fmt.Sprintf("%d", options.Offset),
	}

	resp, err := a.client.Get(ctx, "/search", params)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("search failed: %s", resp.Error)
	}

	var result SearchResult
	err = json.Unmarshal(resp.Body, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetTrackURL 获取音轨URL
func (a *MusicSourceRPCAdapter) GetTrackURL(ctx context.Context, trackID string, quality AudioQuality) (string, error) {
	params := map[string]string{
		"track_id": trackID,
		"quality":  quality.String(),
	}

	resp, err := a.client.Get(ctx, "/track/url", params)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("get track url failed: %s", resp.Error)
	}

	var result struct {
		URL string `json:"url"`
	}
	err = json.Unmarshal(resp.Body, &result)
	if err != nil {
		return "", err
	}

	return result.URL, nil
}

// GetPlaylist 获取播放列表
func (a *MusicSourceRPCAdapter) GetPlaylist(ctx context.Context, id string) (*Playlist, error) {
	resp, err := a.client.Get(ctx, fmt.Sprintf("/playlist/%s", id), nil)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("get playlist failed: %s", resp.Error)
	}

	var playlist Playlist
	err = json.Unmarshal(resp.Body, &playlist)
	if err != nil {
		return nil, err
	}

	return &playlist, nil
}

// Login 用户登录
func (a *MusicSourceRPCAdapter) Login(ctx context.Context, credentials map[string]string) error {
	resp, err := a.client.Post(ctx, "/auth/login", credentials)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("login failed: %s", resp.Error)
	}

	// 处理登录响应，可能需要保存token等
	var loginResp struct {
		Token string `json:"token"`
	}
	err = json.Unmarshal(resp.Body, &loginResp)
	if err != nil {
		return err
	}

	// 设置认证头
	a.client.SetHeader("Authorization", "Bearer "+loginResp.Token)

	return nil
}

// GetClientStats 获取客户端统计信息
func (c *RPCClient) GetClientStats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return map[string]interface{}{
		"base_url":         c.baseURL,
		"timeout":          c.httpClient.Timeout.String(),
		"retry_config":     c.retryConfig,
		"rate_limit":       c.rateLimit,
		"interceptor_count": len(c.interceptors),
		"header_count":     len(c.headers),
	}
}

// Close 关闭客户端
func (c *RPCClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 关闭HTTP客户端的空闲连接
	if transport, ok := c.httpClient.Transport.(*http.Transport); ok {
		transport.CloseIdleConnections()
	}

	return nil
}