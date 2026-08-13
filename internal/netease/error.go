package netease

import (
	"fmt"

	_struct "github.com/go-musicfox/go-musicfox/utils/struct"
)

type Error struct {
	CodeType int
	Msg      string
}

func (e Error) Error() string {
	return fmt.Sprintf("code: %d, msg: %s", e.CodeType, e.Msg)
}

var (
	NetworkErr   = Error{CodeType: -1, Msg: "网络错误"}
	NeedLoginErr = Error{CodeType: 301, Msg: "登录已失效，请重新登录"}
)

// mapResCodeToErr 将 API 响应码映射为错误。登录失效（301）单独区分：
// 运行时会话过期时用户应看到"请重新登录"而不是误导性的"网络错误"。
func mapResCodeToErr(codeType _struct.ResCode) error {
	if codeType == _struct.NeedLogin {
		return NeedLoginErr
	}
	return NetworkErr
}
