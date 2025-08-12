package store

import (
	"fmt"
)

func LRPush(key string, values []string, toLeft bool) (int, error) {
	mu.Lock()
	defer mu.Unlock()

	it, exists := store[key]
	if !exists {
		it = item{typ: TypeList, value: []string{}}
	} else if it.typ != TypeList {
		return 0, fmt.Errorf("WRONGTYPE Operation against a key")
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

func ListLength(key string) (int, error) {
	mu.RLock()
	defer mu.RUnlock()
	it, exists := store[key]
	if !exists {
		return 0, nil
	} else if it.typ != TypeList {
		return 0, fmt.Errorf("WRONGTYPE Operation against a key")
	}
	list := it.value.([]string)
	return len(list), nil
}

func LRange(key string, start int, stop int) ([]string, error) {
	mu.RLock()
	defer mu.RUnlock()
	it, exists := store[key]
	if !exists {
		return []string{}, nil
	} else if it.typ != TypeList {
		return nil, fmt.Errorf("WRONGTYPE Operation against a key")
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
