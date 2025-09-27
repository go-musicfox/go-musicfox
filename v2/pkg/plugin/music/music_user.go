package plugin

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// UserService 用户服务
type UserService struct {
	sources       map[string]MusicSourcePlugin
	cache         Cache
	repository    Repository
	mu            sync.RWMutex
	defaultTTL    time.Duration
	sessionTTL    time.Duration
	activeSessions map[string]*UserSession
}

// UserSession 用户会话
type UserSession struct {
	UserID      string                 `json:"user_id"`
	SourceName  string                 `json:"source_name"`
	Token       string                 `json:"token"`
	RefreshToken string                `json:"refresh_token"`
	ExpiresAt   time.Time              `json:"expires_at"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
	LastAccess  time.Time              `json:"last_access"`
}

// NewUserService 创建用户服务
func NewUserService() *UserService {
	return &UserService{
		sources:        make(map[string]MusicSourcePlugin),
		defaultTTL:     30 * time.Minute,
		sessionTTL:     24 * time.Hour,
		activeSessions: make(map[string]*UserSession),
	}
}

// RegisterSource 注册音乐源
func (u *UserService) RegisterSource(name string, source MusicSourcePlugin) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.sources[name] = source
}

// SetCache 设置缓存
func (u *UserService) SetCache(cache Cache) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.cache = cache
}

// SetRepository 设置数据仓库
func (u *UserService) SetRepository(repo Repository) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.repository = repo
}

// Login 用户登录
func (u *UserService) Login(ctx context.Context, sourceName string, credentials map[string]string) (*UserSession, error) {
	if len(credentials) == 0 {
		return nil, fmt.Errorf("credentials cannot be empty")
	}

	u.mu.RLock()
	source, exists := u.sources[sourceName]
	u.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("source %s not found", sourceName)
	}

	if baseSrc, ok := source.(*BaseMusicSourcePlugin); ok {
		if !baseSrc.HasFeature(MusicSourceFeatureUser) {
			return nil, fmt.Errorf("source %s does not support user feature", sourceName)
		}
	}

	// 执行登录
	err := source.Login(ctx, credentials)
	if err != nil {
		return nil, err
	}

	// 创建用户会话
	session := &UserSession{
		UserID:     credentials["username"], // 或从响应中获取
		SourceName: sourceName,
		Token:      fmt.Sprintf("token_%d", time.Now().Unix()),
		ExpiresAt:  time.Now().Add(u.sessionTTL),
		Metadata:   make(map[string]interface{}),
		CreatedAt:  time.Now(),
		LastAccess: time.Now(),
	}

	// 保存会话
	u.mu.Lock()
	sessionKey := fmt.Sprintf("%s:%s", sourceName, session.UserID)
	u.activeSessions[sessionKey] = session
	u.mu.Unlock()

	// 缓存会话
	if u.cache != nil {
		cacheKey := fmt.Sprintf("user_session:%s", sessionKey)
		u.cache.Set(ctx, cacheKey, session, u.sessionTTL)
	}

	return session, nil
}

// Logout 用户登出
func (u *UserService) Logout(ctx context.Context, sourceName string, userID string) error {
	u.mu.RLock()
	source, exists := u.sources[sourceName]
	u.mu.RUnlock()

	if !exists {
		return fmt.Errorf("source %s not found", sourceName)
	}

	// 执行登出
	err := source.Logout(ctx)
	if err != nil {
		return err
	}

	// 清除缓存
	u.mu.Lock()
	sessionKey := fmt.Sprintf("%s:%s", sourceName, userID)
	delete(u.activeSessions, sessionKey)
	u.mu.Unlock()

	// 清除缓存
	if u.cache != nil {
		cacheKey := fmt.Sprintf("user_session:%s", sessionKey)
		u.cache.Delete(ctx, cacheKey)
	}

	return nil
}

// GetUserInfo 获取用户信息
func (u *UserService) GetUserInfo(ctx context.Context, sourceName string, userID string) (*UserInfo, error) {
	if userID == "" {
		return nil, fmt.Errorf("user id cannot be empty")
	}

	// 检查缓存
	if u.cache != nil {
		cacheKey := fmt.Sprintf("user_info:%s:%s", sourceName, userID)
		if cached, err := u.cache.Get(ctx, cacheKey); err == nil {
			if userInfo, ok := cached.(*UserInfo); ok {
				return userInfo, nil
			}
		}
	}

	u.mu.RLock()
	source, exists := u.sources[sourceName]
	u.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("source %s not found", sourceName)
	}

	userInfo, err := source.GetUserInfo(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 缓存用户信息
	if u.cache != nil {
		cacheKey := fmt.Sprintf("user_info:%s:%s", sourceName, userID)
		u.cache.Set(ctx, cacheKey, userInfo, u.defaultTTL)
	}

	return userInfo, nil
}

// GetUserPlaylists 获取用户播放列表
func (u *UserService) GetUserPlaylists(ctx context.Context, sourceName string, userID string) ([]*Playlist, error) {
	if userID == "" {
		return nil, fmt.Errorf("user id cannot be empty")
	}

	// 检查缓存
	if u.cache != nil {
		cacheKey := fmt.Sprintf("user_playlists:%s:%s", sourceName, userID)
		if cached, err := u.cache.Get(ctx, cacheKey); err == nil {
			if playlists, ok := cached.([]*Playlist); ok {
				return playlists, nil
			}
		}
	}

	u.mu.RLock()
	source, exists := u.sources[sourceName]
	u.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("source %s not found", sourceName)
	}

	playlists, err := source.GetUserPlaylists(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 缓存播放列表
	if u.cache != nil {
		cacheKey := fmt.Sprintf("user_playlists:%s:%s", sourceName, userID)
		u.cache.Set(ctx, cacheKey, playlists, u.defaultTTL)
	}

	return playlists, nil
}

// GetUserLikedTracks 获取用户喜欢的音轨
func (u *UserService) GetUserLikedTracks(ctx context.Context, sourceName string, userID string) ([]*Track, error) {
	if userID == "" {
		return nil, fmt.Errorf("user id cannot be empty")
	}

	// 检查缓存
	if u.cache != nil {
		cacheKey := fmt.Sprintf("user_liked_tracks:%s:%s", sourceName, userID)
		if cached, err := u.cache.Get(ctx, cacheKey); err == nil {
			if tracks, ok := cached.([]*Track); ok {
				return tracks, nil
			}
		}
	}

	u.mu.RLock()
	source, exists := u.sources[sourceName]
	u.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("source %s not found", sourceName)
	}

	tracks, err := source.GetUserLikedTracks(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 缓存喜欢的音轨
	if u.cache != nil {
		cacheKey := fmt.Sprintf("user_liked_tracks:%s:%s", sourceName, userID)
		u.cache.Set(ctx, cacheKey, tracks, u.defaultTTL)
	}

	return tracks, nil
}

// FollowUser 关注用户
func (u *UserService) FollowUser(ctx context.Context, sourceName string, userID string, targetUserID string) error {
	if userID == "" || targetUserID == "" {
		return fmt.Errorf("user ids cannot be empty")
	}

	u.mu.RLock()
	source, exists := u.sources[sourceName]
	u.mu.RUnlock()

	if !exists {
		return fmt.Errorf("source %s not found", sourceName)
	}

	err := source.FollowUser(ctx, targetUserID)
	if err != nil {
		return err
	}

	// 清除相关缓存
	if u.cache != nil {
		u.invalidateUserCache(ctx, sourceName, userID)
		u.invalidateUserCache(ctx, sourceName, targetUserID)
	}

	return nil
}

// UnfollowUser 取消关注用户
func (u *UserService) UnfollowUser(ctx context.Context, sourceName string, userID string, targetUserID string) error {
	if userID == "" || targetUserID == "" {
		return fmt.Errorf("user ids cannot be empty")
	}

	u.mu.RLock()
	source, exists := u.sources[sourceName]
	u.mu.RUnlock()

	if !exists {
		return fmt.Errorf("source %s not found", sourceName)
	}

	err := source.UnfollowUser(ctx, targetUserID)
	if err != nil {
		return err
	}

	// 清除相关缓存
	if u.cache != nil {
		u.invalidateUserCache(ctx, sourceName, userID)
		u.invalidateUserCache(ctx, sourceName, targetUserID)
	}

	return nil
}

// GetUserFollowing 获取用户关注列表
func (u *UserService) GetUserFollowing(ctx context.Context, sourceName string, userID string) ([]string, error) {
	if userID == "" {
		return nil, fmt.Errorf("user id cannot be empty")
	}

	// 检查缓存
	if u.cache != nil {
		cacheKey := fmt.Sprintf("user_following:%s:%s", sourceName, userID)
		if cached, err := u.cache.Get(ctx, cacheKey); err == nil {
			if following, ok := cached.([]string); ok {
				return following, nil
			}
		}
	}

	u.mu.RLock()
	source, exists := u.sources[sourceName]
	u.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("source %s not found", sourceName)
	}

	following, err := source.GetUserFollowing(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 缓存关注列表
	if u.cache != nil {
		cacheKey := fmt.Sprintf("user_following:%s:%s", sourceName, userID)
		u.cache.Set(ctx, cacheKey, following, u.defaultTTL)
	}

	return following, nil
}

// GetUserFollowers 获取用户粉丝列表
func (u *UserService) GetUserFollowers(ctx context.Context, sourceName string, userID string) ([]string, error) {
	if userID == "" {
		return nil, fmt.Errorf("user id cannot be empty")
	}

	// 检查缓存
	if u.cache != nil {
		cacheKey := fmt.Sprintf("user_followers:%s:%s", sourceName, userID)
		if cached, err := u.cache.Get(ctx, cacheKey); err == nil {
			if followers, ok := cached.([]string); ok {
				return followers, nil
			}
		}
	}

	u.mu.RLock()
	source, exists := u.sources[sourceName]
	u.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("source %s not found", sourceName)
	}

	followers, err := source.GetUserFollowers(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 缓存粉丝列表
	if u.cache != nil {
		cacheKey := fmt.Sprintf("user_followers:%s:%s", sourceName, userID)
		u.cache.Set(ctx, cacheKey, followers, u.defaultTTL)
	}

	return followers, nil
}

// GetUserSession 获取用户会话
func (u *UserService) GetUserSession(sourceName string, userID string) (*UserSession, error) {
	u.mu.RLock()
	defer u.mu.RUnlock()

	sessionKey := fmt.Sprintf("%s:%s", sourceName, userID)
	session, exists := u.activeSessions[sessionKey]
	if !exists {
		return nil, fmt.Errorf("session not found")
	}

	// 检查会话是否过期
	if time.Now().After(session.ExpiresAt) {
		delete(u.activeSessions, sessionKey)
		return nil, fmt.Errorf("session expired")
	}

	// 更新最后访问时间
	session.LastAccess = time.Now()

	return session, nil
}

// RefreshSession 刷新用户会话
func (u *UserService) RefreshSession(ctx context.Context, sourceName string, userID string) (*UserSession, error) {
	session, err := u.GetUserSession(sourceName, userID)
	if err != nil {
		return nil, err
	}

	// 延长会话时间
	session.ExpiresAt = time.Now().Add(u.sessionTTL)
	session.LastAccess = time.Now()

	// 更新缓存
	if u.cache != nil {
		sessionKey := fmt.Sprintf("%s:%s", sourceName, userID)
		cacheKey := fmt.Sprintf("user_session:%s", sessionKey)
		u.cache.Set(ctx, cacheKey, session, u.sessionTTL)
	}

	return session, nil
}

// IsUserLoggedIn 检查用户是否已登录
func (u *UserService) IsUserLoggedIn(sourceName string, userID string) bool {
	_, err := u.GetUserSession(sourceName, userID)
	return err == nil
}

// GetActiveUsers 获取活跃用户列表
func (u *UserService) GetActiveUsers(sourceName string) []string {
	u.mu.RLock()
	defer u.mu.RUnlock()

	users := make([]string, 0)
	for _, session := range u.activeSessions {
		if session.SourceName == sourceName && time.Now().Before(session.ExpiresAt) {
			users = append(users, session.UserID)
		}
	}

	return users
}

// CleanupExpiredSessions 清理过期会话
func (u *UserService) CleanupExpiredSessions(ctx context.Context) {
	u.mu.Lock()
	defer u.mu.Unlock()

	now := time.Now()
	for sessionKey, session := range u.activeSessions {
		if now.After(session.ExpiresAt) {
			delete(u.activeSessions, sessionKey)

			// 清除缓存
			if u.cache != nil {
				cacheKey := fmt.Sprintf("user_session:%s", sessionKey)
				u.cache.Delete(ctx, cacheKey)
			}
		}
	}
}

// invalidateUserCache 清除用户相关缓存
func (u *UserService) invalidateUserCache(ctx context.Context, sourceName string, userID string) {
	if u.cache == nil {
		return
	}

	// 清除用户相关的所有缓存
	cacheKeys := []string{
		fmt.Sprintf("user_info:%s:%s", sourceName, userID),
		fmt.Sprintf("user_playlists:%s:%s", sourceName, userID),
		fmt.Sprintf("user_liked_tracks:%s:%s", sourceName, userID),
		fmt.Sprintf("user_following:%s:%s", sourceName, userID),
		fmt.Sprintf("user_followers:%s:%s", sourceName, userID),
	}

	for _, key := range cacheKeys {
		u.cache.Delete(ctx, key)
	}
}

// GetUserStatistics 获取用户统计信息
func (u *UserService) GetUserStatistics(ctx context.Context, sourceName string, userID string) (map[string]interface{}, error) {
	userInfo, err := u.GetUserInfo(ctx, sourceName, userID)
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"level":        userInfo.Level,
		"follow_count": userInfo.FollowCount,
		"fan_count":    userInfo.FanCount,
		"play_count":   userInfo.PlayCount,
		"created_at":   userInfo.CreatedAt,
		"updated_at":   userInfo.UpdatedAt,
	}

	// 获取播放列表数量
	playlists, err := u.GetUserPlaylists(ctx, sourceName, userID)
	if err == nil {
		stats["playlist_count"] = len(playlists)
	}

	// 获取喜欢的音轨数量
	likedTracks, err := u.GetUserLikedTracks(ctx, sourceName, userID)
	if err == nil {
		stats["liked_tracks_count"] = len(likedTracks)
	}

	return stats, nil
}