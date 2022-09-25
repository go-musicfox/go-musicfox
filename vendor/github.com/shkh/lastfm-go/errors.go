package lastfm_go

import (
	"fmt"
	"reflect"
	"strings"
)

const (
	ApiResponseStatusOk     = "ok"
	ApiResponseStatusFailed = "failed"
)

const (
	ErrorAuthRequired          = 50
	ErrorParameterMissing      = 51
	ErrorInvalidTypeOfArgument = 52
)

var Messages = map[int]string{
	50: "This method requires authentication.",
	51: "Required parameter missing. Required: %v, Missing: %v.",
	52: "Invalid type of argument passed. Supported types are int, string and []string.",
}

type LastfmError struct {
	Code    int
	Message string
	Where   string
	Caller  string
}

func (e *LastfmError) Error() string {
	return fmt.Sprintf("LastfmError[%d]: %s (%s)", e.Code, e.Message, e.Caller)
}

func newApiError(errorXml *ApiError) (e *LastfmError) {
	e = new(LastfmError)
	e.Code = errorXml.Code
	e.Message = strings.TrimSpace(errorXml.Message)
	return e
}

func newLibError(code int, message string) (e *LastfmError) {
	e = new(LastfmError)
	e.Code = code
	e.Message = message
	return e
}

func appendCaller(err error, caller string) {
	if err != nil && reflect.TypeOf(err).String() == "*lastfm.LastfmError" {
		err.(*LastfmError).Caller = caller
	}
}
