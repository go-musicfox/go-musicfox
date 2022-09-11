package utils

func Recover(ignore bool) {
	err := recover()
	if err != nil {
		DefaultLogger().Printf("catch panic, err: %+v", err)
		if ignore {
			return
		}
		panic(err)
	}
}
