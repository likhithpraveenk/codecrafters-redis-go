package commands

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/common"
	"github.com/codecrafters-io/redis-starter-go/app/store"
)

func handleReplica(conn net.Conn, cmd []string) bool {
	switch strings.ToUpper(cmd[0]) {
	case "REPLCONF":
		subCmd := strings.ToUpper(cmd[1])
		if subCmd == "ACK" {
			offset, _ := strconv.ParseInt(cmd[2], 10, 64)
			store.UpdateReplicaAck(conn, offset)
			fmt.Printf("[master] replica %v acknowledged offset %d\n", conn.RemoteAddr(), offset)
			return true
		}
		conn.Write(common.Encode(common.SimpleString("OK")))
		return true
	case "PSYNC":
		result := fmt.Sprintf("FULLRESYNC %s %d", store.MasterReplID, store.MasterReplOffset)
		conn.Write(common.Encode(common.SimpleString(result)))
		conn.Write(common.Encode(common.RDB(store.EmptyRDB)))
		store.AddReplica(conn)
		return true
	}
	return false
}

func propagateToReplicas(resp []byte) {
	list := store.ListReplica()
	for _, r := range list {
		_, err := r.Conn.Write(resp)
		if err != nil {
			fmt.Printf("[master] failed to propagate to replica %v\n", err)
			r.Conn.Close()
			store.RemoveReplica(r.Conn)
		}
	}
}

func handleWait(cmd []string) (any, error) {
	if len(cmd) < 3 {
		return nil, fmt.Errorf("ERR wrong number of arguments for 'WAIT' command")
	}
	numReplicas, _ := strconv.ParseInt(cmd[1], 10, 64)
	timeoutMs, _ := strconv.ParseInt(cmd[2], 10, 64)
	timeout := time.Duration(timeoutMs) * time.Millisecond

	targetOffset := store.MasterReplOffset
	propagateToReplicas(common.Encode([]string{"REPLCONF", "GETACK", "*"}))

	ackChans := store.ListAckChans()
	merged := merge(ackChans...)

	var timer <-chan time.Time
	if timeout > 0 {
		timer = time.After(timeout)
	}

	for {
		if acks := store.CountReplicasAtLeast(targetOffset); numReplicas == 0 || acks >= numReplicas {
			return acks, nil
		}

		if timeout > 0 {
			select {
			case <-merged:
			case <-timer:
				return store.CountReplicasAtLeast(targetOffset), nil
			}
		} else {
			<-merged
		}
	}
}

func merge(cs ...chan struct{}) <-chan struct{} {
	out := make(chan struct{}, 1)
	for _, c := range cs {
		go func(ch chan struct{}) {
			<-ch
			select {
			case out <- struct{}{}:
			default:
			}
		}(c)
	}
	return out
}
