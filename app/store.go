package main

import "sync"

var (
	store = make(map[string]string)
	mu    sync.RWMutex
)

func setValue(key, value string) {
	mu.Lock()
	defer mu.Unlock()
	store[key] = value
}

func getValue(key string) (string, bool) {
	mu.RLock()
	defer mu.RUnlock()
	val, ok := store[key]
	return val, ok
}
