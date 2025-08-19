package store

import (
	"sync"
	"time"
)

type valueType int

const (
	TypeString valueType = iota
	TypeList
	TypeStream
	TypeSet
	TypeZSet
	TypeHash
)

type item struct {
	typ       valueType
	value     any
	expiresAt time.Time
}

type StreamEntry struct {
	ID     string
	Fields []string
}

var (
	store = make(map[string]item)
	mu    sync.RWMutex
)

var (
	waitersMu sync.Mutex
	waiters   = make(map[string][]chan struct{})
)
