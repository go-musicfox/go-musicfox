package utils

type errorReset interface {
	ResetError()
}

func ResetError(i interface{}) {
	if r, ok := i.(errorReset); ok {
		r.ResetError()
	}
}
