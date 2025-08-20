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

type ReplicationRole string

const (
	RoleMaster ReplicationRole = "master"
	RoleSlave  ReplicationRole = "slave"
)

var (
	ReplicaRole      = RoleMaster
	MasterHost       string
	MasterPort       int
	MasterLinkStatus = "down"
	ConnectedSlaves  = 0
	MasterReplID     = "0000000000000000000000000000000000000000"
	MasterReplOffset = int64(0)
)
