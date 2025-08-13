package store

import (
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
	value     any
	expiresAt time.Time
}

var (
	store = make(map[string]item)
	mu    sync.RWMutex
)

var (
	waitersMu sync.Mutex
	waiters   = make(map[string][]chan struct{})
)
