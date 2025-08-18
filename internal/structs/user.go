package structs

import (
	"github.com/buger/jsonparser"
	"github.com/pkg/errors"
)

type User struct {
	UserId           int64  `json:"user_id"`
	MyLikePlaylistID int64  `json:"my_like_playlist_id"`
	Nickname         string `json:"nickname"`
	AvatarUrl        string `json:"avatar_url"`
	AccountId        int64  `json:"account_id"`
}

func NewUserFromLocalJson(json []byte) (User, error) {
	var user User
	if len(json) == 0 {
		return user, errors.New("json is empty")
	}

	userId, err := jsonparser.GetInt(json, "user_id")
	if err != nil {
		return user, err
	}
	user.UserId = userId

	if playlistId, err := jsonparser.GetInt(json, "my_like_playlist_id"); err == nil {
		user.MyLikePlaylistID = playlistId
	}

	if nickname, err := jsonparser.GetString(json, "nickname"); err == nil {
		user.Nickname = nickname
	}

	if avatarUrl, err := jsonparser.GetString(json, "avatar_url"); err == nil {
		user.AvatarUrl = avatarUrl
	}

	if accountId, err := jsonparser.GetInt(json, "account_id"); err == nil {
		user.AccountId = accountId
	}

	return user, nil
}

func NewUserFromJson(json []byte, keys ...string) (User, error) {
	var user User
	if len(json) == 0 {
		return user, errors.New("json is empty")
	}

	targetData := json
	if len(keys) > 0 {
		extractedData, _, _, err := jsonparser.Get(json, keys...)
		if err != nil {
			return user, err
		}
		targetData = extractedData
	}

	userId, err := jsonparser.GetInt(targetData, "userId")
	if err != nil {
		return user, err
	}
	user.UserId = userId

	if nickname, err := jsonparser.GetString(targetData, "nickname"); err == nil {
		user.Nickname = nickname
	}

	if avatarUrl, err := jsonparser.GetString(targetData, "avatarUrl"); err == nil {
		user.AvatarUrl = avatarUrl
	}

	return user, nil
}

func NewUserFromJsonForLogin(json []byte) (User, error) {
	var user User

	user, err := NewUserFromJson(json, "profile")
	if err != nil {
		return user, err
	}

	if accountId, err := jsonparser.GetInt(json, "account", "id"); err == nil {
		user.AccountId = accountId
	}

	return user, nil
}

// NewUserFromSearchResultJson 从搜索结果json中获取用户信息
func NewUserFromSearchResultJson(json []byte) (User, error) {
	var user User
	if len(json) == 0 {
		return user, errors.New("json is empty")
	}

	user, err := NewUserFromJson(json)
	if err != nil {
		return user, err
	}

	return user, nil
}
