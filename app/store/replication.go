package store

import (
	"maps"
	"net"
	"sync"
)

var replicaConns = make(map[int]net.Conn)
var replicaConnMu sync.Mutex

var replicaAcks sync.Map

func UpdateReplicaAck(conn net.Conn, offset int64) {
	replicaAcks.Store(conn, offset)
}

func CountReplicasAtLeast(offset int64) int64 {
	count := 0
	replicaAcks.Range(func(_, v any) bool {
		if v.(int64) >= offset {
			count++
		}
		return true
	})
	return int64(count)
}

func RegisterReplicaConn(id int, conn net.Conn) {
	replicaConnMu.Lock()
	defer replicaConnMu.Unlock()
	replicaConns[id] = conn
}

func ListReplicaConns() map[int]net.Conn {
	replicaConnMu.Lock()
	defer replicaConnMu.Unlock()
	list := make(map[int]net.Conn)
	maps.Copy(list, replicaConns)
	return list
}

func DeleteReplicaConn(id int) {
	delete(replicaConns, id)
	ConnectedSlaves = len(replicaConns)
}
