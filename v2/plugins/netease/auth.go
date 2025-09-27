package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/buger/jsonparser"
	"github.com/go-musicfox/netease-music/service"
	"github.com/skip2/go-qrcode"
)

// Login 用户登录
func (p *NeteasePlugin) Login(ctx context.Context, credentials map[string]string) error {
	// 支持多种登录方式
	loginType, exists := credentials["type"]
	if !exists {
		loginType = "phone" // 默认手机号登录
	}

	switch loginType {
	case "phone":
		return p.loginByPhone(ctx, credentials)
	case "email":
		return p.loginByEmail(ctx, credentials)
	case "cookie":
		return p.loginByCookie(ctx, credentials)
	case "qr":
		return p.loginByQR(ctx, credentials)
	default:
		return fmt.Errorf("unsupported login type: %s", loginType)
	}
}

// Logout 用户登出
func (p *NeteasePlugin) Logout(ctx context.Context) error {
	logoutService := &service.LogoutService{}
	code, _, _ := logoutService.Logout()
	if code != 200 {
		return fmt.Errorf("logout failed with code: %f", code)
	}

	// 清除用户信息
	p.mu.Lock()
	p.user = nil
	p.mu.Unlock()

	// 清除cookies
	if p.cookieJar != nil {
		p.cookieJar.SetCookies(
			&url.URL{Scheme: "https", Host: "music.163.com"},
			[]*http.Cookie{},
		)
	}

	return nil
}

// loginByPhone 手机号登录
func (p *NeteasePlugin) loginByPhone(ctx context.Context, credentials map[string]string) error {
	phone, exists := credentials["phone"]
	if !exists {
		return fmt.Errorf("phone number is required")
	}

	password, exists := credentials["password"]
	if !exists {
		return fmt.Errorf("password is required")
	}

	// 处理国家代码
	countryCode := "86"
	if code, exists := credentials["country_code"]; exists {
		countryCode = code
	}

	// 处理带国家代码的手机号格式 (+86 13800138000)
	if phone[0] == '+' {
		parts := strings.Fields(phone)
		if len(parts) == 2 {
			countryCode = strings.TrimPrefix(parts[0], "+")
			phone = parts[1]
		}
	}

	loginService := &service.LoginCellphoneService{
		Phone:       phone,
		Password:    password,
		Countrycode: countryCode,
	}

	code, response, err := loginService.LoginCellphone()
	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}
	if code != 200 {
		return p.handleLoginError(code, response)
	}

	return p.handleLoginSuccess(response)
}

// loginByEmail 邮箱登录
func (p *NeteasePlugin) loginByEmail(ctx context.Context, credentials map[string]string) error {
	email, exists := credentials["email"]
	if !exists {
		return fmt.Errorf("email is required")
	}

	password, exists := credentials["password"]
	if !exists {
		return fmt.Errorf("password is required")
	}

	// 验证邮箱格式
	if !strings.Contains(email, "@") {
		return fmt.Errorf("invalid email format")
	}

	loginService := &service.LoginEmailService{
		Email:    email,
		Password: password,
	}

	code, response := loginService.LoginEmail()
	if code != 200 {
		return p.handleLoginError(code, response)
	}

	return p.handleLoginSuccess(response)
}

// loginByCookie 通过Cookie登录
func (p *NeteasePlugin) loginByCookie(ctx context.Context, credentials map[string]string) error {
	cookieStr, exists := credentials["cookie"]
	if !exists {
		// 尝试从配置文件读取cookie
		if configCookie := p.getCookieFromConfig(); configCookie != "" {
			cookieStr = configCookie
		} else {
			return fmt.Errorf("cookie is required")
		}
	}

	// 解析cookie字符串
	var cookies []*http.Cookie
	if strings.Contains(cookieStr, "=") {
		// 单个cookie格式: name=value
		if cookie, err := http.ParseCookie(cookieStr); err == nil {
			cookies = cookie
		} else {
			return fmt.Errorf("invalid cookie format: %w", err)
		}
	} else {
		// 多个cookie格式，按分号分割
		for _, cookiePart := range strings.Split(cookieStr, ";") {
			cookiePart = strings.TrimSpace(cookiePart)
			if cookiePart == "" {
				continue
			}
			if parts := strings.SplitN(cookiePart, "=", 2); len(parts) == 2 {
				cookies = append(cookies, &http.Cookie{
					Name:  strings.TrimSpace(parts[0]),
					Value: strings.TrimSpace(parts[1]),
				})
			}
		}
	}

	if len(cookies) == 0 {
		return fmt.Errorf("no valid cookies found")
	}

	// 设置cookie到jar
	if p.cookieJar != nil {
		p.cookieJar.SetCookies(
			&url.URL{Scheme: "https", Host: "music.163.com"},
			cookies,
		)
	}

	// 验证登录状态
	return p.verifyLoginStatus()
}

// loginByQR 二维码登录
func (p *NeteasePlugin) loginByQR(ctx context.Context, credentials map[string]string) error {
	// 获取二维码key
	qrService := &service.LoginQRService{}
	code, resp, url, err := qrService.GetKey()
	if err != nil {
		return fmt.Errorf("failed to get QR key: %w", err)
	}
	if code != 200 || url == "" {
		return fmt.Errorf("failed to create QR code: code=%v, resp=%s", code, string(resp))
	}

	// 保存unikey用于后续检查
	p.qrUniKey = qrService.UniKey

	// 生成二维码图片
	if err := p.generateQRCode(url); err != nil {
		return fmt.Errorf("failed to generate QR code image: %w", err)
	}

	// 如果提供了二维码处理回调，调用它
	if qrHandlerInterface, exists := credentials["qr_handler"]; exists {
		// 这里应该是一个接口类型，但由于credentials是map[string]string，暂时跳过
		_ = qrHandlerInterface
	}

	// 轮询检查二维码扫描状态
	return p.pollQRStatus(ctx, qrService.UniKey)
}

// pollQRStatus 轮询二维码登录状态
func (p *NeteasePlugin) pollQRStatus(ctx context.Context, key string) error {
	// 轮询检查登录状态
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	timeout := time.After(5 * time.Minute)

	qrService := &service.LoginQRService{}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout:
			return fmt.Errorf("QR code login timeout")
		case <-ticker.C:
			qrService.UniKey = p.qrUniKey
			code, response, err := qrService.CheckQR()
			if err != nil {
				return fmt.Errorf("failed to check QR status: %w", err)
			}
			
			switch code {
			case 803:
				// 登录成功
				return p.handleLoginSuccess(response)
			case 800:
				// 二维码过期
				return fmt.Errorf("QR code expired")
			case 801:
				// 等待扫描
				continue
			case 802:
				// 等待确认
				continue
			default:
				return fmt.Errorf("QR code check failed: code %f, response: %s", code, string(response))
			}
		}
	}
}

// handleLoginError 处理登录错误
func (p *NeteasePlugin) handleLoginError(code float64, response []byte) error {
	msg := "login failed"
	if message, err := jsonparser.GetString(response, "message"); err == nil {
		msg = message
	} else if message, err := jsonparser.GetString(response, "msg"); err == nil {
		msg = message
	}

	return fmt.Errorf("%s (code: %f)", msg, code)
}

// handleLoginSuccess 处理登录成功
func (p *NeteasePlugin) handleLoginSuccess(response []byte) error {
	// 解析用户信息
	userInfo, err := p.parseUserInfo(response)
	if err != nil {
		return fmt.Errorf("failed to parse user info: %w", err)
	}

	// 保存用户信息
	p.mu.Lock()
	p.user = userInfo
	p.mu.Unlock()

	// 获取用户详细信息
	if err := p.fetchUserDetails(); err != nil {
		// 不影响登录成功，只记录警告
		fmt.Printf("Warning: failed to fetch user details: %v\n", err)
	}

	return nil
}

// parseUserInfo 解析用户信息
func (p *NeteasePlugin) parseUserInfo(data []byte) (*UserInfo, error) {
	userInfo := &UserInfo{}

	// 添加调试日志
	fmt.Printf("[DEBUG] Parsing user info from response: %s\n", string(data))

	// 尝试从不同路径解析用户信息
	userIDPaths := [][]string{
		{"profile", "userId"},
		{"account", "profile", "userId"},
		{"data", "profile", "userId"},
		{"profile", "id"},
		{"account", "id"},
		{"data", "id"},
		{"userId"},
		{"id"},
		{"uid"},
	}

	// 解析用户ID - 支持多种数据类型
	for _, path := range userIDPaths {
		// 尝试解析为int64
		if userID, err := jsonparser.GetInt(data, path...); err == nil {
			userInfo.UserID = userID
			fmt.Printf("[DEBUG] Found user ID as int64: %d at path %v\n", userID, path)
			break
		}
		// 尝试解析为字符串然后转换
		if userIDStr, err := jsonparser.GetString(data, path...); err == nil {
			if userID, parseErr := strconv.ParseInt(userIDStr, 10, 64); parseErr == nil {
				userInfo.UserID = userID
				fmt.Printf("[DEBUG] Found user ID as string: %s (converted to %d) at path %v\n", userIDStr, userID, path)
				break
			}
		}
	}

	// 解析昵称
	nicknamePaths := [][]string{
		{"profile", "nickname"},
		{"account", "profile", "nickname"},
		{"data", "profile", "nickname"},
		{"profile", "name"},
		{"nickname"},
		{"name"},
	}

	for _, path := range nicknamePaths {
		if nickname, err := jsonparser.GetString(data, path...); err == nil {
			userInfo.Nickname = nickname
			fmt.Printf("[DEBUG] Found nickname: %s at path %v\n", nickname, path)
			break
		}
	}

	// 解析头像URL
	avatarPaths := [][]string{
		{"profile", "avatarUrl"},
		{"account", "profile", "avatarUrl"},
		{"data", "profile", "avatarUrl"},
		{"profile", "avatar"},
		{"avatarUrl"},
		{"avatar"},
	}

	for _, path := range avatarPaths {
		if avatarURL, err := jsonparser.GetString(data, path...); err == nil {
			userInfo.Avatar = avatarURL
			fmt.Printf("[DEBUG] Found avatar URL: %s at path %v\n", avatarURL, path)
			break
		}
	}

	if userInfo.UserID == 0 {
		// 打印详细的调试信息
		fmt.Printf("[ERROR] Failed to parse user ID from response. Tried paths: %v\n", userIDPaths)
		fmt.Printf("[ERROR] Response data: %s\n", string(data))
		return nil, fmt.Errorf("failed to parse user ID from response data")
	}

	fmt.Printf("[DEBUG] Successfully parsed user info: ID=%d, Nickname=%s\n", userInfo.UserID, userInfo.Nickname)
	return userInfo, nil
}

// fetchUserDetails 获取用户详细信息
func (p *NeteasePlugin) fetchUserDetails() error {
	p.mu.RLock()
	userID := p.user.UserID
	p.mu.RUnlock()

	if userID == 0 {
		return fmt.Errorf("user ID is empty")
	}

	// 获取用户详细信息
	userDetailService := &service.UserDetailService{
		Uid: strconv.FormatInt(userID, 10),
	}

	code, response := userDetailService.UserDetail()
	if code != 200 {
		return fmt.Errorf("failed to get user details: code %f", code)
	}

	// 更新用户信息
	p.mu.Lock()
	defer p.mu.Unlock()

	// 解析更多用户信息
	if level, err := jsonparser.GetInt(response, "level"); err == nil {
		// 可以扩展UserInfo结构来包含更多信息
		_ = level
	}

	return nil
}

// verifyLoginStatus 验证登录状态
func (p *NeteasePlugin) verifyLoginStatus() error {
	accountService := &service.UserAccountService{}
	code, response := accountService.AccountInfo()
	if code != 200 {
		return fmt.Errorf("failed to verify login status: code %f", code)
	}

	// 解析用户信息
	userInfo, err := p.parseUserInfo(response)
	if err != nil {
		return fmt.Errorf("failed to parse user info: %w", err)
	}

	// 保存用户信息
	p.mu.Lock()
	p.user = userInfo
	p.mu.Unlock()

	return nil
}

// IsLoggedIn 检查是否已登录
func (p *NeteasePlugin) IsLoggedIn() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.user != nil && p.user.UserID != 0
}

// GetCurrentUser 获取当前登录用户信息
func (p *NeteasePlugin) GetCurrentUser() *UserInfo {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.user == nil {
		return nil
	}
	// 返回副本以避免并发问题
	return &UserInfo{
		UserID:   p.user.UserID,
		Nickname: p.user.Nickname,
		Avatar:   p.user.Avatar,
	}
}

// RefreshLogin 刷新登录状态
func (p *NeteasePlugin) RefreshLogin(ctx context.Context) error {
	if !p.IsLoggedIn() {
		return fmt.Errorf("user not logged in")
	}

	refreshService := &service.LoginRefreshService{}
	code, _ := refreshService.LoginRefresh()
	if code != 200 {
		return fmt.Errorf("failed to refresh login: code %f", code)
	}

	return nil
}

// generateQRCode 生成二维码图片
func (p *NeteasePlugin) generateQRCode(url string) error {
	// 创建数据目录
	dataDir := "/tmp/musicfox" // TODO: 从配置获取
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	// 生成二维码
	qrPath := filepath.Join(dataDir, "qrcode.png")
	err := qrcode.WriteFile(url, qrcode.Medium, 256, qrPath)
	if err != nil {
		return fmt.Errorf("failed to generate QR code: %w", err)
	}

	return nil
}

// GetQRCodePath 获取二维码图片路径
func (p *NeteasePlugin) GetQRCodePath() string {
	dataDir := "/tmp/musicfox" // TODO: 从配置获取
	return filepath.Join(dataDir, "qrcode.png")
}

// StartQRLogin 开始二维码登录（异步）
func (p *NeteasePlugin) StartQRLogin(ctx context.Context) (string, string, error) {
	// 获取二维码key
	qrService := &service.LoginQRService{}
	code, resp, url, err := qrService.GetKey()
	if err != nil {
		return "", "", fmt.Errorf("failed to get QR key: %w", err)
	}
	if code != 200 || url == "" {
		return "", "", fmt.Errorf("failed to create QR code: code=%v, resp=%s", code, string(resp))
	}

	// 保存unikey
	p.qrUniKey = qrService.UniKey

	// 生成二维码图片
	if err := p.generateQRCode(url); err != nil {
		return "", "", fmt.Errorf("failed to generate QR code image: %w", err)
	}

	return qrService.UniKey, url, nil
}

// CheckQRLogin 检查二维码登录状态
func (p *NeteasePlugin) CheckQRLogin(ctx context.Context, qrKey string) (string, error) {
	qrService := &service.LoginQRService{}
	qrService.UniKey = qrKey
	code, response, err := qrService.CheckQR()
	if err != nil {
		return "error", fmt.Errorf("failed to check QR status: %w", err)
	}

	switch code {
	case 803:
		// 登录成功
		if err := p.handleLoginSuccess(response); err != nil {
			return "error", err
		}
		return "success", nil
	case 800:
		// 二维码过期
		return "expired", nil
	case 801:
		// 等待扫描
		return "waiting", nil
	case 802:
		// 等待确认
		return "scanned", nil
	default:
		return "error", fmt.Errorf("QR code check failed: code %f", code)
	}
}

// getCookieFromConfig 从配置文件读取cookie
func (p *NeteasePlugin) getCookieFromConfig() string {
	// TODO: 从插件配置中读取cookie
	// 这里应该从插件上下文或配置管理器中获取
	return ""
}

// SaveCookieToConfig 保存cookie到配置文件
func (p *NeteasePlugin) SaveCookieToConfig(cookieStr string) error {
	// TODO: 保存cookie到插件配置
	// 这里应该通过插件上下文或配置管理器保存
	return nil
}

// AutoLogin 自动登录（从配置文件读取cookie）
func (p *NeteasePlugin) AutoLogin(ctx context.Context) error {
	cookieStr := p.getCookieFromConfig()
	if cookieStr == "" {
		return fmt.Errorf("no saved cookie found")
	}

	credentials := map[string]string{
		"type":   "cookie",
		"cookie": cookieStr,
	}

	return p.Login(ctx, credentials)
}

// GetLoginMethods 获取支持的登录方式
func (p *NeteasePlugin) GetLoginMethods() []string {
	return []string{"phone", "email", "cookie", "qr"}
}

// ValidateCredentials 验证登录凭据
func (p *NeteasePlugin) ValidateCredentials(credentials map[string]string) error {
	loginType, exists := credentials["type"]
	if !exists {
		return fmt.Errorf("login type is required")
	}

	switch loginType {
	case "phone":
		if credentials["phone"] == "" {
			return fmt.Errorf("phone number is required")
		}
		if credentials["password"] == "" {
			return fmt.Errorf("password is required")
		}
	case "email":
		if credentials["email"] == "" {
			return fmt.Errorf("email is required")
		}
		if credentials["password"] == "" {
			return fmt.Errorf("password is required")
		}
		if !strings.Contains(credentials["email"], "@") {
			return fmt.Errorf("invalid email format")
		}
	case "cookie":
		if credentials["cookie"] == "" && p.getCookieFromConfig() == "" {
			return fmt.Errorf("cookie is required")
		}
	case "qr":
		// 二维码登录不需要额外验证
	default:
		return fmt.Errorf("unsupported login type: %s", loginType)
	}

	return nil
}