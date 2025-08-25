package store

import (
	"maps"
	"net"
	"sync"
)

var replicaConns = make(map[int]net.Conn)
var replicaConnMu sync.Mutex

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
