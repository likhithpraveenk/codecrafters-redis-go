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

type Item struct {
	typ       valueType
	value     any
	expiresAt time.Time
}

var (
	store = make(map[string]Item)
	mu    sync.RWMutex
)
