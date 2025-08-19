package store

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

func XAdd(key, id string, fields []string) (string, error) {
	mu.Lock()
	defer mu.Unlock()

	it, exists := store[key]
	if !exists {
		it = item{typ: TypeStream, value: []StreamEntry{}}
	} else if it.typ != TypeStream {
		return "", fmt.Errorf("WRONGTYPE Operation against a key holding the wrong kind of value")
	}
	stream := it.value.([]StreamEntry)
	var lastID string
	if len(stream) > 0 {
		lastID = stream[len(stream)-1].ID
	}
	newId, err := validateAndGenerateStreamID(id, lastID)
	if err != nil {
		return "", err
	}
	stream = append(stream, StreamEntry{ID: newId, Fields: fields})
	it.value = stream
	store[key] = it
	sWaitersMu.Lock()
	for _, ch := range sWaiters[key] {
		close(ch)
	}
	sWaiters[key] = nil
	sWaitersMu.Unlock()
	return newId, nil
}

func validateAndGenerateStreamID(id string, lastID string) (string, error) {
	lastMS, lastSeq := parseIDParts(lastID)
	switch {
	case id == "*":
		msInt := time.Now().UnixNano() / 1e6
		ms := fmt.Sprintf("%d", msInt)
		var seq int64 = 0
		if lastID != "" {
			parts := strings.Split(lastID, "-")
			if len(parts) == 2 && parts[0] == ms {
				lastSeq, _ := strconv.ParseInt(parts[1], 10, 64)
				seq = lastSeq + 1
			}
		}
		return fmt.Sprintf("%s-%d", ms, seq), nil
	case strings.HasSuffix(id, "-*"):
		ms := strings.TrimSuffix(id, "-*")
		msInt, err := strconv.ParseInt(ms, 10, 64)
		if err != nil || msInt < 0 {
			return "", fmt.Errorf("ERR invalid stream ID '%s'", id)
		}
		var seq int64 = 0
		switch msInt {
		case lastMS:
			seq = lastSeq + 1
		case 0:
			seq = 1
		}
		if !isIDGreater(msInt, seq, lastMS, lastSeq) {
			return "", fmt.Errorf("ERR The ID specified is equal or smaller than the target stream top item")
		}
		return fmt.Sprintf("%d-%d", msInt, seq), nil
	default:
		parts := strings.Split(id, "-")
		if len(parts) != 2 {
			return "", fmt.Errorf("ERR invalid stream ID '%s'", id)
		}
		ms, err1 := strconv.ParseInt(parts[0], 10, 64)
		seq, err2 := strconv.ParseInt(parts[1], 10, 64)
		if err1 != nil || err2 != nil || ms < 0 || seq < 0 || (ms == 0 && seq == 0) {
			return "", fmt.Errorf("ERR The ID specified in XADD must be greater than 0-0")
		}
		if !isIDGreater(ms, seq, lastMS, lastSeq) {
			return "", fmt.Errorf("ERR The ID specified in XADD is equal or smaller than the target stream top item")
		}
		return id, nil
	}
}

func parseIDParts(id string) (int64, int64) {
	if id == "" {
		return 0, 0
	}
	parts := strings.Split(id, "-")
	if len(parts) != 2 {
		return 0, 0
	}

	ms, _ := strconv.ParseInt(parts[0], 10, 64)
	seq, _ := strconv.ParseInt(parts[1], 10, 64)
	return ms, seq
}

func isIDGreater(ms, seq, lastMs, lastSeq int64) bool {
	if ms > lastMs {
		return true
	}
	if ms == lastMs && seq > lastSeq {
		return true
	}
	return false
}

func isIDLesser(ms, seq, lastMs, lastSeq int64) bool {
	if ms < lastMs {
		return true
	}
	if ms == lastMs && seq < lastSeq {
		return true
	}
	return false
}

func isIDEqual(ms, seq, lastMs, lastSeq int64) bool {
	if ms == lastMs && seq == lastSeq {
		return true
	}
	return false
}

func XRange(key, startStr, endStr string) ([][]any, error) {
	mu.RLock()
	defer mu.RUnlock()
	it, exists := store[key]
	if !exists {
		return [][]any{}, nil
	} else if it.typ != TypeStream {
		return nil, fmt.Errorf("WRONGTYPE Operation against a key holding the wrong kind of value")
	}
	stream := it.value.([]StreamEntry)
	var startMs, startSeq, endMs, endSeq int64
	if startStr == "-" {
		startMs, startSeq = 0, 0
	} else {
		startMs, startSeq = parseIDParts(startStr)
	}
	if endStr == "+" {
		endMs, endSeq = math.MaxInt64, math.MaxInt64
	} else {
		endMs, endSeq = parseIDParts(endStr)
	}

	result := make([][]any, 0)

	for _, entry := range stream {
		idMs, idSeq := parseIDParts(entry.ID)
		if isIDGreater(idMs, idSeq, startMs, startSeq) || isIDEqual(idMs, idSeq, startMs, startSeq) {
			if isIDLesser(idMs, idSeq, endMs, endSeq) || isIDEqual(idMs, idSeq, endMs, endSeq) {
				result = append(result, []any{entry.ID, entry.Fields})
			}
		}
	}
	return result, nil
}

func XRead(keys, ids []string) ([][]any, error) {
	mu.RLock()
	defer mu.RUnlock()

	noOfStreams := len(keys)
	result := make([][]any, 0)

	for i := range noOfStreams {
		item, exists := store[keys[i]]
		if !exists {
			continue
		}
		if item.typ != TypeStream {
			return nil, fmt.Errorf("WRONGTYPE Operation against a key holding the wrong kind of value")
		}
		streamEntries := item.value.([]StreamEntry)
		idMs, idSeq := parseIDParts(ids[i])
		matched := []any{}
		for _, entry := range streamEntries {
			entryMs, entrySeq := parseIDParts(entry.ID)
			if isIDGreater(entryMs, entrySeq, idMs, idSeq) {
				matched = append(matched, []any{entry.ID, entry.Fields})
			}
		}
		if len(matched) > 0 {
			result = append(result, []any{keys[i], matched})
		}
	}
	if len(result) == 0 {
		return nil, nil
	}
	return result, nil
}

func XReadBlock(keys, ids []string, timeout time.Duration) ([][]any, error) {
	chans := make([]chan struct{}, len(keys))
	addToStreamWaiters(keys, chans)
	defer cleanupStreamWaiters(keys, chans)

	var timer <-chan time.Time
	if timeout > 0 {
		timer = time.After(timeout)
	}

	for {
		result, err := XRead(keys, ids)
		if err != nil {
			return nil, err
		}
		if len(result) > 0 {
			return result, nil
		}

		if timeout > 0 {
			select {
			case <-merge(chans...):
			case <-timer:
				return nil, nil
			}
		} else {
			<-merge(chans...)
		}

	}
}

func addToStreamWaiters(keys []string, chans []chan struct{}) {
	sWaitersMu.Lock()
	defer sWaitersMu.Unlock()
	for i, key := range keys {
		ch := make(chan struct{}, 1)
		chans[i] = ch
		if sWaiters == nil {
			sWaiters = make(map[string][]chan struct{})
		}
		sWaiters[key] = append(sWaiters[key], ch)
	}
}

func cleanupStreamWaiters(keys []string, chans []chan struct{}) {
	sWaitersMu.Lock()
	defer sWaitersMu.Unlock()
	for i, key := range keys {
		waits := sWaiters[key]
		newWaiters := waits[:0]
		for _, w := range waits {
			if chans[i] != w {
				newWaiters = append(newWaiters, w)
			}
		}
		sWaiters[key] = newWaiters
	}
}

func merge(cs ...chan struct{}) <-chan struct{} {
	out := make(chan struct{}, 1)
	for _, c := range cs {
		go func(ch chan struct{}) {
			<-ch
			select {
			case out <- struct{}{}:
			default:
			}
		}(c)
	}
	return out
}
