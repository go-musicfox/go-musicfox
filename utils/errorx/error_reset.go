package errorx

type errorReset interface {
	ResetError()
}

func ResetError(i any) {
	if r, ok := i.(errorReset); ok {
		r.ResetError()
	}
}
