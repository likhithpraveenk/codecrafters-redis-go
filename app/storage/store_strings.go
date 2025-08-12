package store

import "time"

func SetValue(key, value string, expireMS int) {
	mu.Lock()
	defer mu.Unlock()

	it := item{typ: TypeString, value: value}
	if expireMS > 0 {
		it.expiresAt = time.Now().Add(time.Duration(expireMS) * time.Millisecond)
	}
	store[key] = it
}

func GetValue(key string) (string, bool) {
	mu.RLock()
	defer mu.RUnlock()
	it, ok := store[key]
	if !ok || it.typ != TypeString {
		return "", false
	}
	if !it.expiresAt.IsZero() && time.Now().After(it.expiresAt) {
		mu.RUnlock()
		mu.Lock()
		delete(store, key)
		mu.Unlock()
		mu.RLock()
		return "", false
	}

	return it.value.(string), ok
}
