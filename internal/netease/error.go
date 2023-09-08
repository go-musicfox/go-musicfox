package netease

import (
	"fmt"
)

type Error struct {
	CodeType int
	Msg      string
}

func (e Error) Error() string {
	return fmt.Sprintf("code: %d, msg: %s", e.CodeType, e.Msg)
}

var NetworkErr = Error{CodeType: -1, Msg: "网络错误"}
