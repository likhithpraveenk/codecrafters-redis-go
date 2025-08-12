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

func LRPush(key string, values []string, toLeft bool) (int, error) {
	mu.Lock()
	defer mu.Unlock()

	it, exists := store[key]
	if !exists {
		it = item{typ: TypeList, value: []string{}}
	} else if it.typ != TypeList {
		return 0, fmt.Errorf("WRONGTYPE operation against key %v", key)
	}
	list := it.value.([]string)
	if toLeft {
		for _, v := range values {
			list = append([]string{v}, list...)
		}
	} else {
		list = append(list, values...)
	}
	it.value = list
	store[key] = it
	return len(list), nil
}

func LRange(key string, start int, stop int) ([]string, error) {
	mu.RLock()
	defer mu.RUnlock()
	it, exists := store[key]
	if !exists {
		return []string{}, nil
	} else if it.typ != TypeList {
		return nil, fmt.Errorf("WRONGTYPE operation against key %v", key)
	}
	list := it.value.([]string)
	n := len(list)
	if start < 0 {
		start = n + start
	}
	if stop < 0 {
		stop = n + stop
	}
	if start < 0 {
		start = 0
	}
	if stop >= n {
		stop = n - 1
	}
	if start > stop || start >= n {
		return []string{}, nil
	}
	return list[start : stop+1], nil
}
