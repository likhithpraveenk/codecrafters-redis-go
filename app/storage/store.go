package store

import (
	"fmt"
	"sync"
	"time"
)

type valueType int

const (
	TypeString valueType = iota
	TypeList
)

type item struct {
	typ       valueType
	value     interface{}
	expiresAt time.Time
}

var (
	store = make(map[string]item)
	mu    sync.RWMutex
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

func RPush(key string, values []string) (int, error) {
	mu.Lock()
	defer mu.Unlock()

	it, exists := store[key]
	if !exists {
		it = item{typ: TypeList, value: []string{}}
	} else if it.typ != TypeList {
		return 0, fmt.Errorf("WRONGTYPE operation against key %v", key)
	}
	list := it.value.([]string)
	list = append(list, values...)
	it.value = list
	store[key] = it
	return len(list), nil
}
