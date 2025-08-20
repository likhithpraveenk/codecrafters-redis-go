package store

import (
	"fmt"
	"net"
	"strings"
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

func Info() string {
	var sb strings.Builder
	sb.WriteString("# Replication\r\n")
	if ReplicaRole == RoleMaster {
		sb.WriteString("role:master\r\n")
		sb.WriteString("connected_slaves:0\r\n")
		sb.WriteString("master_replid:123456789abcdef\r\n")
		sb.WriteString("master_repl_offset:0\r\n")
	} else {
		sb.WriteString("role:slave\r\n")
		sb.WriteString(fmt.Sprintf("master_host:%s\r\n", MasterHost))
		sb.WriteString(fmt.Sprintf("master_port:%d\r\n", MasterPort))
		sb.WriteString("master_link_status:up\r\n")
	}
	return sb.String()
}
