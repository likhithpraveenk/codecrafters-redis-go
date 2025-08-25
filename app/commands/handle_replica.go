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

	targetOffset := store.MasterReplOffset
	PropagateToReplicas(common.Encode([]string{"REPLCONF", "GETACK", "*"}))

	deadline := time.Now().Add(time.Duration(timeoutMs) * time.Millisecond)
	ticker := time.NewTicker(5 * time.Millisecond)
	defer ticker.Stop()

	for now := range ticker.C {
		acked := store.CountReplicasAtLeast(targetOffset)
		if acked >= numReplicas || now.After(deadline) {
			return acked, nil
		}
	}

	return store.CountReplicasAtLeast(targetOffset), nil
}
