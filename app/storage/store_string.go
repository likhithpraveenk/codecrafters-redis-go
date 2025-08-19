package store

import (
	"fmt"
	"strconv"
	"time"
)

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

func Increment(key string) (int64, error) {
	mu.Lock()
	defer mu.Unlock()
	it, ok := store[key]
	if !ok {
		store[key] = item{typ: TypeString, value: "1"}
		return 1, nil
	}
	if it.typ != TypeString {
		return 0, fmt.Errorf("ERR value is not an integer or out of range")
	}

	strVal := it.value.(string)
	n, err := strconv.ParseInt(strVal, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("ERR value is not an integer or out of range")
	}
	n++
	it.value = strconv.FormatInt(n, 10)
	store[key] = it
	return n, nil
}
