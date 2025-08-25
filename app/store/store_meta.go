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

type ReplicationRole string

const (
	RoleMaster ReplicationRole = "master"
	RoleSlave  ReplicationRole = "slave"
)

type ReplicaInfo struct {
	ID     int
	Addr   string
	Port   string
	Offset int64
	State  string
}

var (
	ReplicaRole      = RoleMaster
	MasterHost       string
	MasterPort       string
	MasterLinkStatus = "down"
	ConnectedSlaves  = 0
	MasterReplID     = "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb"
	MasterReplOffset = int64(0)
	ReplicaOffset    = int64(0)

	replicas  = make(map[int]*ReplicaInfo)
	replicaMu sync.Mutex
	nextRepID = 0
)

func AddReplica(addr, port string) *ReplicaInfo {
	replicaMu.Lock()
	defer replicaMu.Unlock()

	r := ReplicaInfo{
		ID:     nextRepID,
		Addr:   addr,
		Port:   port,
		Offset: 0,
		State:  "online",
	}
	replicas[nextRepID] = &r
	nextRepID++
	ConnectedSlaves = len(replicas)
	return &r
}

func ListReplicas() []*ReplicaInfo {
	replicaMu.Lock()
	defer replicaMu.Unlock()

	list := make([]*ReplicaInfo, 0, len(replicas))
	for _, r := range replicas {
		list = append(list, r)
	}
	return list
}

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
		sb.WriteString(fmt.Sprintf("connected_slaves:%d\r\n", ConnectedSlaves))
		for _, r := range ListReplicas() {
			sb.WriteString(fmt.Sprintf(
				"slave%d:ip=%s,port=%s,state=%s,offset=%d,lag=0\r\n",
				r.ID, r.Addr, r.Port, r.State, r.Offset,
			))
		}
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
