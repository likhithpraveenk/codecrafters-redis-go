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
		return 0, fmt.Errorf("WRONGTYPE Operation against a key holding the wrong kind of value")
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

func (ss *sortedSet) rebuildOrder() {
	ss.order = ss.order[:0]
	for member := range ss.scores {
		ss.order = append(ss.order, member)
	}

	sort.Slice(ss.order, func(i, j int) bool {
		return ss.scores[ss.order[i]] < ss.scores[ss.order[j]]
	})
}
