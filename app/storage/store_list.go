package store

import (
	"fmt"
	"time"
)

func notifyWaiters(key string) {
	waitersMu.Lock()
	defer waitersMu.Unlock()
	if chans, ok := waiters[key]; ok && len(chans) > 0 {
		close(chans[0])
		waiters[key] = chans[1:]
	}
}

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
	notifyWaiters(key)
	return len(list), nil
}

func LPop(key string) (string, bool) {
	mu.Lock()
	defer mu.Unlock()

	it, exists := store[key]
	if !exists || it.typ != TypeList {
		return "", false
	}
	list := it.value.([]string)
	if len(list) == 0 {
		return "", false
	}
	value := list[0]
	rest := list[1:]
	if len(rest) == 0 {
		delete(store, key)
	} else {
		it.value = rest
		store[key] = it
	}
	return value, true
}

func BLPop(keys []string, timeout time.Duration) (string, string, error) {
	var deadline time.Time
	if timeout != 0 {
		deadline = time.Now().Add(timeout)
	}
	waitCh := make(chan struct{}, 1)
	defer cleanupWaiters(keys, waitCh)

	for {
		if key, val, ok := tryPopFromKeys(keys); ok {
			return key, val, nil
		}

		waitersMu.Lock()
		for _, key := range keys {
			waiters[key] = append(waiters[key], waitCh)
		}
		waitersMu.Unlock()

		if timeout == 0 {
			<-waitCh
		} else {
			waitTime := time.Until(deadline)
			if waitTime <= 0 {
				return "", "", nil
			}

			select {
			case <-waitCh:
			case <-time.After(waitTime):
				return "", "", nil
			}
		}
	}
}

func tryPopFromKeys(keys []string) (string, string, bool) {
	mu.Lock()
	defer mu.Unlock()
	for _, key := range keys {
		it, exists := store[key]
		if exists && it.typ == TypeList {
			list := it.value.([]string)
			if len(list) > 0 {
				val := list[0]
				it.value = list[1:]
				store[key] = it
				return key, val, true
			}
		}
	}
	return "", "", false
}

func cleanupWaiters(keys []string, waitCh chan struct{}) {
	waitersMu.Lock()
	defer waitersMu.Unlock()
	for _, key := range keys {
		chans := waiters[key]
		for i, ch := range chans {
			if ch == waitCh {
				waiters[key] = append(chans[:i], chans[i+1:]...)
				break
			}
		}
	}

}

func LPopCount(key string, count int) ([]string, error) {
	it, exists := store[key]
	if !exists {
		return nil, fmt.Errorf("key does not exist")
	}
	if it.typ != TypeList {
		return nil, fmt.Errorf("WRONGTYPE Operation against a key")
	}
	list := it.value.([]string)
	if len(list) == 0 {
		return []string{}, nil
	}
	if count > len(list) {
		count = len(list)
	}
	values := list[:count]
	rest := list[count:]
	if len(rest) == 0 {
		delete(store, key)
	} else {
		it.value = rest
		store[key] = it
	}
	return values, nil
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
