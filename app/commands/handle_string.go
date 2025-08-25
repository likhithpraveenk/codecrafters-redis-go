package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/common"
	"github.com/codecrafters-io/redis-starter-go/app/store"
)

func handlePing(cmd []string) (any, error) {
	return common.SimpleString("PONG"), nil
}

func handleEcho(cmd []string) (any, error) {
	if len(cmd) < 2 {
		return nil, fmt.Errorf("wrong arguments for 'ECHO'")
	}
	return cmd[1], nil
}

func handleSet(cmd []string) (any, error) {
	if len(cmd) < 3 {
		return nil, fmt.Errorf("wrong arguments for 'SET'")
	}
	key := cmd[1]
	value := cmd[2]
	expiry := 0

	if len(cmd) >= 5 && strings.ToUpper(cmd[3]) == "PX" {
		milliSec, err := strconv.Atoi(cmd[4])
		if err != nil || milliSec <= 0 {
			return nil, fmt.Errorf("invalid expire time")
		}
		expiry = milliSec
	}
	store.SetValue(key, value, expiry)
	return common.SimpleString("OK"), nil
}

func handleGet(cmd []string) (any, error) {
	if len(cmd) < 2 {
		return nil, fmt.Errorf("wrong arguments for 'GET'")
	}
	val, ok := store.GetValue(cmd[1])
	if !ok {
		return nil, nil
	} else {
		return val, nil
	}
}

func handleIncrement(cmd []string) (any, error) {
	if len(cmd) < 2 {
		return nil, fmt.Errorf("wrong arguments for 'INCR'")
	}
	key := cmd[1]
	val, err := store.Increment(key)
	if err != nil {
		return nil, fmt.Errorf("%s", err.Error())
	}
	return val, nil
}
