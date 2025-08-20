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

var (
	ReplicationRole  = "master" // or "slave"
	MasterHost       = ""
	MasterPort       = 0
	MasterLinkStatus = "down"
	ConnectedSlaves  = 0
	MasterReplID     = "0000000000000000000000000000000000000000" // dummy
	MasterReplOffset = 0
)
