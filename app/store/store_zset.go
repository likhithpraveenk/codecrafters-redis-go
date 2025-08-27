package store

import (
	"fmt"
	"sort"
)

type sortedSet struct {
	scores map[string]float64
	order  []string
}

func ZAdd(key string, score float64, member string) (int64, error) {
	GlobalStore.mu.Lock()
	defer GlobalStore.mu.Unlock()
	it, ok := GlobalStore.items[key]
	if !ok {
		it = Item{
			typ: TypeZSet,
			value: sortedSet{
				scores: make(map[string]float64),
				order:  []string{},
			},
		}
	} else if it.typ != TypeZSet {
		return -1, fmt.Errorf("WRONGTYPE Operation against a key holding the wrong kind of value")
	}
	ss := it.value.(sortedSet)
	_, exists := ss.scores[member]
	ss.scores[member] = score
	ss.rebuildOrder()
	it.value = ss
	GlobalStore.items[key] = it
	if exists {
		return 0, nil
	}
	return 1, nil
}

func ZRank(key, member string) (int64, error) {
	GlobalStore.mu.RLock()
	defer GlobalStore.mu.RUnlock()
	it, ok := GlobalStore.items[key]
	if !ok {
		return -1, nil
	}
	if it.typ != TypeZSet {
		return -1, fmt.Errorf("WRONGTYPE Operation against a key holding the wrong kind of value")
	}
	ss := it.value.(sortedSet)
	for i, m := range ss.order {
		if m == member {
			return int64(i), nil
		}
	}
	return -1, nil
}

func (ss *sortedSet) rebuildOrder() {
	ss.order = ss.order[:0]
	for member := range ss.scores {
		ss.order = append(ss.order, member)
	}

	sort.Slice(ss.order, func(i, j int) bool {
		si, sj := ss.scores[ss.order[i]], ss.scores[ss.order[j]]
		if si == sj {
			return ss.order[i] < ss.order[j]
		}
		return si < sj
	})
}
