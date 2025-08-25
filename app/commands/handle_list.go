package commands

import (
	"fmt"
	"strconv"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/store"
)

func handlePush(cmd []string, toLeft bool) (any, error) {
	if len(cmd) < 3 {
		name := "RPUSH"
		if toLeft {
			name = "LPUSH"
		}
		return nil, fmt.Errorf("wrong arguments for '%s'", name)
	}
	key := cmd[1]
	values := cmd[2:]
	length, err := store.LRPush(key, values, toLeft)
	if err != nil {
		return nil, fmt.Errorf("%s", err.Error())
	}
	return length, nil
}

func handleRPush(cmd []string) (any, error) {
	return handlePush(cmd, false)
}

func handleLPush(cmd []string) (any, error) {
	return handlePush(cmd, true)
}

func handleLRange(cmd []string) (any, error) {
	if len(cmd) < 4 {
		return nil, fmt.Errorf("wrong arguments for 'LRange'")
	}
	key := cmd[1]
	start, err := strconv.Atoi(cmd[2])
	if err != nil {
		return nil, fmt.Errorf("value is not an integer")
	}
	stop, err := strconv.Atoi(cmd[3])
	if err != nil {
		return nil, fmt.Errorf("value is not an integer")
	}
	values, err := store.LRange(key, start, stop)
	if err != nil {
		return nil, fmt.Errorf("%s", err.Error())
	}
	return values, nil
}

func handleLLen(cmd []string) (any, error) {
	if len(cmd) < 2 {
		return nil, fmt.Errorf("wrong arguments for 'LLEN'")
	}
	val, err := store.ListLength(cmd[1])
	if err != nil {
		return nil, fmt.Errorf("%s", err.Error())
	}
	return val, nil

}

func handleLPop(cmd []string) (any, error) {
	if len(cmd) < 2 {
		return nil, fmt.Errorf("wrong arguments for 'LPOP'")
	}
	key := cmd[1]
	if len(cmd) > 2 {
		count, err := strconv.Atoi(cmd[2])
		if err != nil || count <= 0 {
			return nil, fmt.Errorf("count must be a positive integer")
		}
		values, err := store.LPopCount(key, count)
		if err != nil {
			return nil, fmt.Errorf("%s", err.Error())
		}
		return values, nil

	} else {
		value, ok := store.LPop(key)
		if !ok {
			return nil, nil
		}
		return value, nil
	}
}

func handleBLPop(cmd []string) (any, error) {
	if len(cmd) < 3 {
		return nil, fmt.Errorf("wrong arguments for 'BLPOP'")
	}

	keys := cmd[1 : len(cmd)-1]
	timeoutStr := cmd[len(cmd)-1]

	timeoutFloat, err := strconv.ParseFloat(timeoutStr, 64)
	if err != nil || timeoutFloat < 0 {
		return nil, fmt.Errorf("timeout must be a non-negative number")
	}

	var timeout time.Duration
	if timeoutFloat == 0 {
		timeout = 0
	} else {
		timeout = time.Duration(timeoutFloat * float64(time.Second))
	}
	key, val, err := store.BLPop(keys, timeout)
	if err != nil {
		return nil, fmt.Errorf("%s", err.Error())
	}
	if key == "" && val == "" {
		return nil, nil
	}
	arr := []string{key, val}
	return arr, nil
}
