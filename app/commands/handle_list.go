package commands

import (
	"net"
	"strconv"
	"time"

	store "github.com/codecrafters-io/redis-starter-go/app/storage"
)

func handlePush(cmd []string, conn net.Conn, toLeft bool) error {
	if len(cmd) < 3 {
		name := "RPUSH"
		if toLeft {
			name = "LPUSH"
		}
		return writeToConn(conn, Encode(ErrorString("wrong arguments for '"+name+"'")))
	}
	key := cmd[1]
	values := cmd[2:]
	length, err := store.LRPush(key, values, toLeft)
	if err != nil {
		return writeToConn(conn, Encode(ErrorString(err.Error())))
	}
	return writeToConn(conn, Encode(length))
}

func handleRPush(cmd []string, conn net.Conn) error {
	return handlePush(cmd, conn, false)
}

func handleLPush(cmd []string, conn net.Conn) error {
	return handlePush(cmd, conn, true)
}

func handleLRange(cmd []string, conn net.Conn) error {
	if len(cmd) < 4 {
		return writeToConn(conn, Encode(ErrorString("wrong arguments for 'LRange'")))
	}
	key := cmd[1]
	start, err := strconv.Atoi(cmd[2])
	if err != nil {
		return writeToConn(conn, Encode(ErrorString("value is not an integer")))
	}
	stop, err := strconv.Atoi(cmd[3])
	if err != nil {
		return writeToConn(conn, Encode(ErrorString("value is not an integer")))
	}
	values, err := store.LRange(key, start, stop)
	if err != nil {
		return writeToConn(conn, Encode(ErrorString(err.Error())))
	}
	return writeToConn(conn, Encode(values))
}

func handleLLen(cmd []string, conn net.Conn) error {
	if len(cmd) < 2 {
		return writeToConn(conn, Encode(ErrorString("wrong arguments for 'LLEN'")))
	}
	val, err := store.ListLength(cmd[1])
	if err != nil {
		return writeToConn(conn, Encode(ErrorString(err.Error())))
	}
	return writeToConn(conn, Encode(val))

}

func handleLPop(cmd []string, conn net.Conn) error {
	if len(cmd) < 2 {
		return writeToConn(conn, Encode(ErrorString("wrong arguments for 'LPOP'")))
	}
	key := cmd[1]
	if len(cmd) > 2 {
		count, err := strconv.Atoi(cmd[2])
		if err != nil || count <= 0 {
			return writeToConn(conn, Encode(ErrorString("count must be a positive integer")))
		}
		values, err := store.LPopCount(key, count)
		if err != nil {
			return writeToConn(conn, Encode(ErrorString(err.Error())))
		}
		return writeToConn(conn, Encode(values))

	} else {
		value, ok := store.LPop(key)
		if !ok {
			return writeToConn(conn, Encode(nil))
		}
		return writeToConn(conn, Encode(value))
	}
}

func handleBLPop(cmd []string, conn net.Conn) error {
	if len(cmd) < 3 {
		return writeToConn(conn, Encode(ErrorString("wrong arguments for 'BLPOP'")))
	}

	keys := cmd[1 : len(cmd)-1]
	timeoutStr := cmd[len(cmd)-1]

	timeoutFloat, err := strconv.ParseFloat(timeoutStr, 64)
	if err != nil || timeoutFloat < 0 {
		return writeToConn(conn, Encode(ErrorString("timeout must be a non-negative number")))
	}

	var timeout time.Duration
	if timeoutFloat == 0 {
		timeout = 0
	} else {
		timeout = time.Duration(timeoutFloat * float64(time.Second))
	}
	key, val, err := store.BLPop(keys, timeout)
	if err != nil {
		return writeToConn(conn, Encode(ErrorString(err.Error())))
	}
	if key == "" && val == "" {
		return writeToConn(conn, Encode(nil))
	}
	arr := []string{key, val}
	return writeToConn(conn, Encode(arr))
}
