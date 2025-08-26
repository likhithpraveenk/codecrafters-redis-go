package store

import (
	"fmt"
	"strconv"
	"time"
)

func SetValue(key, value string, expireMS int) {
	GlobalStore.mu.Lock()
	defer GlobalStore.mu.Unlock()

	it := Item{typ: TypeString, value: value}
	if expireMS > 0 {
		it.expiresAt = time.Now().Add(time.Duration(expireMS) * time.Millisecond)
	}
	GlobalStore.items[key] = it
}

func GetValue(key string) (string, bool) {
	GlobalStore.mu.RLock()
	it, ok := GlobalStore.items[key]
	GlobalStore.mu.RUnlock()

	if !ok || it.typ != TypeString {
		return "", false
	}

	if !it.expiresAt.IsZero() && time.Now().After(it.expiresAt) {
		GlobalStore.mu.Lock()
		delete(GlobalStore.items, key)
		GlobalStore.mu.Unlock()
		return "", false
	}

	return it.value.(string), ok
}

func Increment(key string) (int64, error) {
	GlobalStore.mu.Lock()
	defer GlobalStore.mu.Unlock()
	it, ok := GlobalStore.items[key]
	if !ok {
		GlobalStore.items[key] = Item{typ: TypeString, value: "1"}
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
	GlobalStore.items[key] = it
	return n, nil
}
