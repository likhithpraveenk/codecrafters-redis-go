package commands

import (
	"fmt"
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/app/store"
)

func handleZAdd(cmd []string) (any, error) {
	if len(cmd) < 4 {
		return nil, fmt.Errorf("wrong number of arguments for 'ZADD'")
	}
	key, member := cmd[1], cmd[3]

	score, err := strconv.ParseFloat(cmd[2], 64)
	if err != nil {
		return nil, fmt.Errorf("ERR score is not a valid float")
	}

	return store.ZAdd(key, score, member)
}
