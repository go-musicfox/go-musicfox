package cookiejar

import (
	"github.com/gomodule/redigo/redis"
	"github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type RedisDrive struct {
	pool       *redis.Pool
	entries    map[string]map[string]entry
	namespaces string
}

func (r *RedisDrive) Set(key string, val map[string]entry) {
	r.entries[key] = val
	r.saveEntries(key)
}

func (r *RedisDrive) Get(key string) map[string]entry {
	return r.entries[key]
}

func (r *RedisDrive) Delete(key string) {
	delete(r.entries, key)
}

func (r *RedisDrive) saveEntries(k string) error {
	conn := r.pool.Get()
	defer conn.Close()

	v, err := json.MarshalToString(r.entries[k])
	if err != nil {
		return err
	}

	if _, err := conn.Do("HSET", r.namespaces, k, v); err != nil {
		return err
	}

	return nil
}

func (r *RedisDrive) readEntries() {
	c := r.pool.Get()
	defer c.Close()

	keys, err := redis.Strings(c.Do("HKEYS", r.namespaces))
	if err != nil {
		return
	}

	for _, k := range keys {
		b, err := redis.Bytes(c.Do("HGET", r.namespaces, k))
		if err != nil {
			continue
		}

		e := make(map[string]entry)
		if err := json.Unmarshal(b, &e); err != nil {
			// resolve fail
		}

		r.entries[k] = e
	}
}
