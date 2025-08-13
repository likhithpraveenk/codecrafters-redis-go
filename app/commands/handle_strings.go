package commands

import (
	"net"
	"strconv"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/protocol"
	store "github.com/codecrafters-io/redis-starter-go/app/storage"
)

func handlePing(cmd []string, conn net.Conn) error {
 return writeToConn(conn,(protocol.EncodeSimpleString("PONG")))

}

func handleEcho(cmd []string, conn net.Conn) error {
	if len(cmd) < 2 {
		return writeToConn(conn,(protocol.EncodeError("wrong arguments for 'ECHO'")))
	}
 	return writeToConn(conn,(protocol.EncodeBulkString(cmd[1])))
}

func handleSet(cmd []string, conn net.Conn) error {
	if len(cmd) < 3 {
		return writeToConn(conn,(protocol.EncodeError("wrong arguments for 'SET'")))
	}
	key := cmd[1]
	value := cmd[2]
	expiry := 0

	if len(cmd) >= 5 && strings.ToUpper(cmd[3]) == "PX" {
		milliSec, err := strconv.Atoi(cmd[4])
		if err != nil || milliSec <= 0 {
			return writeToConn(conn,(protocol.EncodeError("invalid expire time")))
		}
		expiry = milliSec
	}
	store.SetValue(key, value, expiry)
	return writeToConn(conn,(protocol.EncodeSimpleString("OK")))

}

func handleGet(cmd []string, conn net.Conn) error {
	if len(cmd) < 2 {
		return writeToConn(conn,(protocol.EncodeError("wrong arguments for 'GET'")))
	}
	val, ok := store.GetValue(cmd[1])
	if !ok {
		return writeToConn(conn,(protocol.EncodeNullString()))
	} else {
		return writeToConn(conn,(protocol.EncodeBulkString(val)))
	}
}
