package cookiejar

type EntriesDrive struct {
	entries map[string]map[string]entry
}

func (e *EntriesDrive) Set(key string, val map[string]entry) {
	e.entries[key] = val
}

func (e *EntriesDrive) Get(key string) map[string]entry {
	return e.entries[key]
}

func (e *EntriesDrive) Delete(key string) {
	delete(e.entries, key)
}