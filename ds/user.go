package ds

import (
	"errors"
	"github.com/buger/jsonparser"
)

type User struct {
	UserId    int64  `json:"user_id"`
	Nickname  string `json:"nickname"`
	AvatarUrl string `json:"avatar_url"`
	AccountId int64  `json:"account_id"`
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

func NewUserFromJson(json []byte) (User, error) {
	var user User
	if len(json) == 0 {
		return user, errors.New("json is empty")
	}

	userId, err := jsonparser.GetInt(json, "profile", "userId")
	if err != nil {
		return user, err
	}
	user.UserId = userId

	if nickname, err := jsonparser.GetString(json, "profile", "nickname"); err == nil {
		user.Nickname = nickname
	}

	if avatarUrl, err := jsonparser.GetString(json, "profile", "avatarUrl"); err == nil {
		user.AvatarUrl = avatarUrl
	}

	if accountId, err := jsonparser.GetInt(json, "account", "id"); err == nil {
		user.AccountId = accountId
	}

	return user, nil
}