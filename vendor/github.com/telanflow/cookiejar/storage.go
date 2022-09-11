package cookiejar

import (
	"net/http"

	"crypto/sha1"
	"encoding/hex"

	"github.com/gomodule/redigo/redis"
)

type Storage interface {
	Set(string, map[string]entry)
	Get(string) map[string]entry
	Delete(string)
}

func NewFileJar(filename string, o *Options) (http.CookieJar, error) {
	store := &FileDrive{
		filename: filename,
		entries:  make(map[string]map[string]entry),
	}
	store.readEntries()

	return New(store, o)
}

func NewEntriesJar(o *Options) (http.CookieJar, error) {
	store := &EntriesDrive{
		entries: make(map[string]map[string]entry),
	}
	return New(store, o)
}

func NewRedisJar(namespaces string, pool *redis.Pool, o *Options) (http.CookieJar, error) {
	if namespaces == "" {
		namespaces = "cookiejar"
	}

	r := sha1.Sum([]byte(namespaces))
	namespaces = hex.EncodeToString(r[:])

	store := &RedisDrive{
		pool:       pool,
		namespaces: namespaces,
		entries:    make(map[string]map[string]entry),
	}
	store.readEntries()
	return New(store, o)
}