package commands

import (
	"fmt"
	"strconv"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/common"
	"github.com/codecrafters-io/redis-starter-go/app/store"
)

func handleWait(cmd []string) (any, error) {
	if len(cmd) < 3 {
		return nil, fmt.Errorf("ERR wrong number of arguments for 'WAIT' command")
	}
	numReplicas, _ := strconv.ParseInt(cmd[1], 10, 64)
	timeoutMs, _ := strconv.ParseInt(cmd[2], 10, 64)
	timeout := time.Duration(timeoutMs) * time.Millisecond

	targetOffset := store.MasterReplOffset
	PropagateToReplicas(common.Encode([]string{"REPLCONF", "GETACK", "*"}))

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
