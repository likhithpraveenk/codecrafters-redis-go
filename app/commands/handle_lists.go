package commands

import (
	"net"
	"strconv"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/protocol"
	store "github.com/codecrafters-io/redis-starter-go/app/storage"
)

func handlePush(cmd []string, conn net.Conn, toLeft bool) error {
	if len(cmd) < 3 {
		name := "RPUSH"
		if toLeft {
			name = "LPUSH"
		}
		return writeToConn(conn, protocol.EncodeError("wrong arguments for '"+name+"'"))
	}
	key := cmd[1]
	values := cmd[2:]
	length, err := store.LRPush(key, values, toLeft)
	if err != nil {
		return writeToConn(conn, protocol.EncodeError(err.Error()))
	}
	return writeToConn(conn, protocol.EncodeInteger(length))
}

func handleRPush(cmd []string, conn net.Conn) error {
	return handlePush(cmd, conn, false)
}

func handleLPush(cmd []string, conn net.Conn) error {
	return handlePush(cmd, conn, true)
}

func handleLRange(cmd []string, conn net.Conn) error {
	if len(cmd) < 4 {
		return writeToConn(conn, protocol.EncodeError("wrong arguments for 'LRange'"))
	}
	key := cmd[1]
	start, err := strconv.Atoi(cmd[2])
	if err != nil {
		return writeToConn(conn, protocol.EncodeError("value is not an integer"))
	}
	stop, err := strconv.Atoi(cmd[3])
	if err != nil {
		return writeToConn(conn, protocol.EncodeError("value is not an integer"))
	}
	values, err := store.LRange(key, start, stop)
	if err != nil {
		return writeToConn(conn, protocol.EncodeError(err.Error()))
	}
	return writeToConn(conn, protocol.EncodeArray(values))
}

func handleLLen(cmd []string, conn net.Conn) error {
	if len(cmd) < 2 {
		return writeToConn(conn, protocol.EncodeError("wrong arguments for 'LLEN'"))
	}
	val, err := store.ListLength(cmd[1])
	if err != nil {
		return writeToConn(conn, protocol.EncodeError(err.Error()))
	}
	return writeToConn(conn, protocol.EncodeInteger(val))

}

func handleLPop(cmd []string, conn net.Conn) error {
	if len(cmd) < 2 {
		return writeToConn(conn, protocol.EncodeError("wrong arguments for 'LPOP'"))
	}
	key := cmd[1]
	if len(cmd) > 2 {
		count, err := strconv.Atoi(cmd[2])
		if err != nil || count <= 0 {
			return writeToConn(conn, protocol.EncodeError("count must be a positive integer"))
		}
		values, err := store.LPopCount(key, count)
		if err != nil {
			return writeToConn(conn, protocol.EncodeError(err.Error()))
		}
		return writeToConn(conn, protocol.EncodeArray(values))

	} else {
		value, ok := store.LPop(key)
		if !ok {
			return writeToConn(conn, protocol.EncodeNullString())
		}
		return writeToConn(conn, protocol.EncodeBulkString(value))
	}
}

func handleBLPop(cmd []string, conn net.Conn) error {
	if len(cmd) < 3 {
		return writeToConn(conn, protocol.EncodeError("wrong arguments for 'BLPOP'"))
	}

	keys := cmd[1 : len(cmd)-1]
	timeoutStr := cmd[len(cmd)-1]

	timeoutFloat, err := strconv.ParseFloat(timeoutStr, 64)
	if err != nil || timeoutFloat < 0 {
		return writeToConn(conn, protocol.EncodeError("timeout must be a non-negative number"))
	}

	var timeout time.Duration
	if timeoutFloat == 0 {
		timeout = 0
	} else {
		timeout = time.Duration(timeoutFloat * float64(time.Second))
	}
	key, val, err := store.BLPop(keys, timeout)
	if err != nil {
		return writeToConn(conn, protocol.EncodeError(err.Error()))
	}
	if key == "" && val == "" {
		return writeToConn(conn, protocol.EncodeNullString())
	}
	arr := []string{key, val}
	return writeToConn(conn, protocol.EncodeArray(arr))
}
