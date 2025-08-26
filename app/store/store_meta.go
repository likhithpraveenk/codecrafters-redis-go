package store

import (
	"fmt"
	"net"
	"path/filepath"
	"strings"
	"sync"
)

type TxnState struct {
	InMulti    bool
	QueuedCmds [][]string
	Subscribed bool
}

var (
	txnMu    sync.Mutex
	connTxns = make(map[net.Conn]*TxnState)
)

type ReplicationRole string

const (
	RoleMaster ReplicationRole = "master"
	RoleSlave  ReplicationRole = "slave"
)

var (
	ReplicaRole      = RoleMaster
	MasterHost       string
	MasterPort       string
	MasterLinkStatus = "down"
	ConnectedSlaves  = 0
	MasterReplID     = "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb"
	MasterReplOffset = int64(0)
	ReplicaOffset    = int64(0)
)

type Config struct {
	Port       int
	ReplicaOf  string
	Dir        string
	DBFilename string
}

var ServerConfig Config

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
	GlobalStore.mu.RLock()
	defer GlobalStore.mu.RUnlock()
	it, exists := GlobalStore.items[key]
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
		sb.WriteString(fmt.Sprintf("connected_slaves:%d\r\n", ConnectedSlaves))
		sb.WriteString("master_replid:123456789abcdef\r\n")
		sb.WriteString("master_repl_offset:0\r\n")
	} else {
		sb.WriteString("role:slave\r\n")
		sb.WriteString(fmt.Sprintf("master_host:%s\r\n", MasterHost))
		sb.WriteString(fmt.Sprintf("master_port:%s\r\n", MasterPort))
		sb.WriteString(fmt.Sprintf("master_link_status:%s\r\n", MasterLinkStatus))
	}
	return sb.String()
}

func Keys(pattern string) []string {
	GlobalStore.mu.RLock()
	defer GlobalStore.mu.RUnlock()

	var results []string
	for key := range GlobalStore.items {
		matched, err := filepath.Match(pattern, key)
		if err != nil {
			continue
		}
		if matched {
			results = append(results, key)
		}
	}
	return results
}
