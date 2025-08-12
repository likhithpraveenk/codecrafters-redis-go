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

func handleRPush(cmd []string, conn net.Conn) error {
	if len(cmd) < 3 {
		conn.Write(protocol.EncodeError("wrong arguments for 'RPUSH'"))
		return nil
	}
	key := cmd[1]
	values := cmd[2:]
	length, err := store.RPush(key, values)
	if err != nil {
		conn.Write(protocol.EncodeError(err.Error()))
		return nil
	}
	_, err = conn.Write(protocol.EncodeInteger(length))
	return err
}
