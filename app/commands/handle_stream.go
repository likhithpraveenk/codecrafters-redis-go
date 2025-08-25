package commands

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/store"
)

func handleXAdd(cmd []string) (any, error) {
	if len(cmd) < 5 || len(cmd)%2 != 1 {
		return nil, fmt.Errorf("wrong number of arguments for 'XADD'")
	}
	key := cmd[1]
	id := cmd[2]
	fields := make([]string, 0)
	for i := 3; i < len(cmd); i++ {
		fields = append(fields, cmd[i])
	}
	id, err := store.XAdd(key, id, fields)
	if err != nil {
		return nil, fmt.Errorf("%s", err.Error())
	}
	return id, nil
}

func handleXRange(cmd []string) (any, error) {
	if len(cmd) < 4 {
		return nil, fmt.Errorf("wrong number of arguments for 'XRANGE'")
	}
	key, start, end := cmd[1], cmd[2], cmd[3]
	result, err := store.XRange(key, start, end)
	if err != nil {
		return nil, fmt.Errorf("%s", err.Error())
	}
	return result, nil
}

func handleXRead(cmd []string) (any, error) {
	if len(cmd) < 4 {
		return nil, fmt.Errorf("wrong number of arguments for 'XREAD'")
	}
	var blockMs int
	var hasBlock bool
	var i int
	if strings.ToUpper(cmd[1]) == "BLOCK" {
		hasBlock = true
		if len(cmd) < 6 {
			return nil, fmt.Errorf("wrong number of arguments for 'XREAD'")
		}
		ms, err := strconv.Atoi(cmd[2])
		if err != nil {
			return nil, fmt.Errorf("invalid BLOCK timeout")
		}
		blockMs = ms
		i = 3
	} else {
		hasBlock = false
		i = 1
	}

	if strings.ToUpper(cmd[i]) != "STREAMS" {
		return nil, fmt.Errorf("syntax error")
	}

	keysIds := cmd[i+1:]
	if len(keysIds)%2 != 0 {
		return nil, fmt.Errorf("wrong number of arguments for 'XREAD'")
	}

	keysLen := len(keysIds) / 2
	keys := keysIds[:keysLen]
	ids := keysIds[keysLen:]

	if !hasBlock {
		result, err := store.XRead(keys, ids)
		if err != nil {
			return nil, fmt.Errorf("%s", err.Error())
		}
		return result, nil
	} else {
		timeout := time.Duration(blockMs) * time.Millisecond
		if blockMs == 0 {
			timeout = 0
		}

		blockResult, blockErr := store.XReadBlock(keys, ids, timeout)
		if blockErr != nil {
			return nil, fmt.Errorf("%s", blockErr.Error())
		}
		return blockResult, nil
	}
}
