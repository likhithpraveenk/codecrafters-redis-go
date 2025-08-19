package store

import (
	"net"
	"sync"
)

type TxnState struct {
	InMulti    bool
	QueuedCmds [][]string
}

var (
	txnMu    sync.Mutex
	connTxns = make(map[net.Conn]*TxnState)
)

func GetTxnState(conn net.Conn) *TxnState {
	txnMu.Lock()
	defer txnMu.Unlock()
	if _, ok := connTxns[conn]; !ok {
		connTxns[conn] = &TxnState{}
	}
	return connTxns[conn]
}

func ClearTxnState(conn net.Conn) {
	txnMu.Lock()
	defer txnMu.Unlock()
	delete(connTxns, conn)
}

func GetType(key string) string {
	mu.RLock()
	defer mu.RUnlock()
	it, exists := store[key]
	if !exists {
		return "none"
	}
	switch it.typ {
	case TypeString:
		return "string"
	case TypeList:
		return "list"
	case TypeStream:
		return "stream"
	case TypeHash:
		return "hash"
	case TypeSet:
		return "set"
	case TypeZSet:
		return "zset"
	default:
		return "none"
	}
}
