package commands

import (
	"net"
	"strconv"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/protocol"
	store "github.com/codecrafters-io/redis-starter-go/app/storage"
)

func Init() {
	registerCommand("PING", handlePing)
	registerCommand("ECHO", handleEcho)
	registerCommand("SET", handleSet)
	registerCommand("GET", handleGet)
	registerCommand("RPUSH", handleRPush)
	registerCommand("LPUSH", handleLPush)
	registerCommand("LRANGE", handleLRange)
	registerCommand("LLEN", handleLLen)
}

func handlePing(cmd []string, conn net.Conn) error {
	_, err := conn.Write(protocol.EncodeSimpleString("PONG"))
	return err
}

func handleEcho(cmd []string, conn net.Conn) error {
	if len(cmd) < 2 {
		conn.Write(protocol.EncodeError("wrong arguments for 'ECHO'"))
		return nil
	}
	_, err := conn.Write(protocol.EncodeBulkString(cmd[1]))
	return err
}

func handleSet(cmd []string, conn net.Conn) error {
	if len(cmd) < 3 {
		conn.Write(protocol.EncodeError("wrong arguments for 'SET'"))
		return nil
	}
	key := cmd[1]
	value := cmd[2]
	expiry := 0

	if len(cmd) >= 5 && strings.ToUpper(cmd[3]) == "PX" {
		milliSec, err := strconv.Atoi(cmd[4])
		if err != nil || milliSec <= 0 {
			conn.Write(protocol.EncodeError("invalid expire time"))
		}
		expiry = milliSec
	}
	store.SetValue(key, value, expiry)
	_, err := conn.Write(protocol.EncodeSimpleString("OK"))
	return err
}

func handleGet(cmd []string, conn net.Conn) error {
	if len(cmd) < 2 {
		conn.Write(protocol.EncodeError("wrong arguments for 'GET'"))
		return nil
	}
	val, ok := store.GetValue(cmd[1])
	if !ok {
		_, err := conn.Write([]byte("$-1\r\n"))
		return err
	} else {
		_, err := conn.Write(protocol.EncodeBulkString(val))
		return err
	}
}

func handlePush(cmd []string, conn net.Conn, toLeft bool) error {
	if len(cmd) < 3 {
		name := "RPUSH"
		if toLeft {
			name = "LPUSH"
		}
		conn.Write(protocol.EncodeError("wrong arguments for '" + name + "'"))
		return nil
	}
	key := cmd[1]
	values := cmd[2:]
	length, err := store.LRPush(key, values, toLeft)
	if err != nil {
		conn.Write(protocol.EncodeError(err.Error()))
		return nil
	}
	_, err = conn.Write(protocol.EncodeInteger(length))
	return err
}

func handleRPush(cmd []string, conn net.Conn) error {
	return handlePush(cmd, conn, false)
}

func handleLPush(cmd []string, conn net.Conn) error {
	return handlePush(cmd, conn, true)
}

func handleLRange(cmd []string, conn net.Conn) error {
	if len(cmd) < 4 {
		conn.Write(protocol.EncodeError("wrong arguments for 'LRange'"))
		return nil
	}
	key := cmd[1]
	start, err := strconv.Atoi(cmd[2])
	if err != nil {
		conn.Write(protocol.EncodeError("value is not an integer or out of range"))
		return nil
	}
	stop, err := strconv.Atoi(cmd[3])
	if err != nil {
		conn.Write(protocol.EncodeError("value is not an integer or out of range"))
		return nil
	}
	values, err := store.LRange(key, start, stop)
	if err != nil {
		conn.Write(protocol.EncodeError(err.Error()))
		return nil
	}
	conn.Write(protocol.EncodeArray(values))
	return nil
}

func handleLLen(cmd []string, conn net.Conn) error {
	if len(cmd) < 2 {
		conn.Write(protocol.EncodeError("wrong arguments for 'LLEN'"))
		return nil
	}
	val, err := store.ListLength(cmd[1])
	if err != nil {
		_, writeErr := conn.Write(protocol.EncodeError(err.Error()))
		return writeErr
	}
	_, writeErr := conn.Write(protocol.EncodeInteger(val))
	return writeErr

}
