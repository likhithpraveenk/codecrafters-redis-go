package main

import (
	"sync"
	"time"
)

type item struct {
	value     string
	expiresAt time.Time
}

var (
	store = make(map[string]item)
	mu    sync.RWMutex
)

func setValue(key, value string, expireMS int) {
	mu.Lock()
	defer mu.Unlock()

	var expiry time.Time
	if expireMS > 0 {
		expiry = time.Now().Add(time.Duration(expireMS) * time.Millisecond)
	}
	store[key] = item{
		value:     value,
		expiresAt: expiry,
	}
}

func getValue(key string) (string, bool) {
	mu.RLock()
	defer mu.RUnlock()
	item, ok := store[key]
	if !ok {
		return "", false
	}
	if !item.expiresAt.IsZero() && time.Now().After(item.expiresAt) {
		mu.RUnlock()
		mu.Lock()
		delete(store, key)
		mu.Unlock()
		mu.RLock()
		return "", false
	}

	return item.value, ok
}
