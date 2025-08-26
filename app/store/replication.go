package store

import (
	"net"
	"sync"
)

type Replica struct {
	ID    int
	Conn  net.Conn
	Ack   int64
	AckCh chan struct{}
}

var (
	replicaMu     sync.Mutex
	replicas      = make(map[int]*Replica)
	connToReplica = make(map[net.Conn]int)
	nextID        = 1
)

func UpdateReplicaAck(conn net.Conn, offset int64) {
	replicaMu.Lock()
	defer replicaMu.Unlock()

	id, ok := connToReplica[conn]
	if !ok {
		return
	}

	r, ok := replicas[id]
	if !ok {
		return
	}

	if offset > r.Ack {
		r.Ack = offset
		select {
		case r.AckCh <- struct{}{}:
		default:
		}
	}
}

func ListAckChans() []chan struct{} {
	replicaMu.Lock()
	defer replicaMu.Unlock()
	ackChans := make([]chan struct{}, 0, len(replicas))
	for _, r := range replicas {
		ackChans = append(ackChans, r.AckCh)
	}
	return ackChans
}

func CountReplicasAtLeast(offset int64) int64 {
	replicaMu.Lock()
	defer replicaMu.Unlock()
	count := 0
	for _, r := range replicas {
		if r.Ack >= offset {
			count++
		}
	}
	return int64(count)
}

func AddReplica(conn net.Conn) {
	replicaMu.Lock()
	defer replicaMu.Unlock()
	id := nextID
	nextID++

	r := &Replica{
		ID:    id,
		Conn:  conn,
		Ack:   0,
		AckCh: make(chan struct{}, 1),
	}
	replicas[id] = r
	ConnectedSlaves = len(replicas)
	connToReplica[conn] = id
}

func ListReplica() []*Replica {
	replicaMu.Lock()
	defer replicaMu.Unlock()
	list := make([]*Replica, 0, len(replicas))
	for _, r := range replicas {
		list = append(list, r)
	}
	return list
}

func RemoveReplica(conn net.Conn) {
	replicaMu.Lock()
	defer replicaMu.Unlock()
	id, ok := connToReplica[conn]
	if !ok {
		return
	}
	delete(connToReplica, conn)
	delete(replicas, id)
	ConnectedSlaves = len(replicas)
}
