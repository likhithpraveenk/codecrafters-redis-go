package store

import (
	"fmt"
	"time"
)

func XAdd(key, id string, fields map[string]string) (string, error) {
	mu.Lock()
	defer mu.Unlock()

	it, exists := store[key]
	if !exists {
		it = item{typ: TypeStream, value: []StreamEntry{}}
	} else if it.typ != TypeStream {
		return "", fmt.Errorf("WRONGTYPE Operation against a key")
	}
	stream := it.value.([]StreamEntry)

	if id == "*" {
		ms := time.Now().UnixNano()
		id = fmt.Sprintf("%d-0", ms)
	}
	stream = append(stream, StreamEntry{ID: id, Fields: fields})
	it.value = stream
	store[key] = it
	return id, nil
}
