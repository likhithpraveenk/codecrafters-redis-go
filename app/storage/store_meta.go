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
	var b strings.Builder
	b.WriteString("# Replication\r\n")
	if ReplicationRole == "master" {
		b.WriteString(fmt.Sprintf("role:%s\r\n", ReplicationRole))
		b.WriteString(fmt.Sprintf("connected_slaves:%d\r\n", ConnectedSlaves))
		b.WriteString(fmt.Sprintf("master_replid:%s\r\n", MasterReplID))
		b.WriteString(fmt.Sprintf("master_repl_offset:%d\r\n", MasterReplOffset))
	} else {
		b.WriteString(fmt.Sprintf("role:%s\r\n", ReplicationRole))
		b.WriteString(fmt.Sprintf("master_host:%s\r\n", MasterHost))
		b.WriteString(fmt.Sprintf("master_port:%d\r\n", MasterPort))
		b.WriteString(fmt.Sprintf("master_link_status:%s\r\n", MasterLinkStatus))
		b.WriteString(fmt.Sprintf("master_last_io_seconds_ago:%d\r\n", 0))
	}
	return b.String()
}
