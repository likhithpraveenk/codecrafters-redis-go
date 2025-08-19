package commands

import (
	"net"
	"strconv"
	"strings"

	store "github.com/codecrafters-io/redis-starter-go/app/storage"
)

func handlePing(cmd []string, conn net.Conn) error {
	return writeToConn(conn, Encode(SimpleString("PONG")))
}

func handleEcho(cmd []string, conn net.Conn) error {
	if len(cmd) < 2 {
		return writeToConn(conn, (Encode(ErrorString("wrong arguments for 'ECHO'"))))
	}
	return writeToConn(conn, Encode(cmd[1]))
}

func handleSet(cmd []string, conn net.Conn) error {
	if len(cmd) < 3 {
		return writeToConn(conn, Encode(ErrorString("wrong arguments for 'SET'")))
	}
	key := cmd[1]
	value := cmd[2]
	expiry := 0

	if len(cmd) >= 5 && strings.ToUpper(cmd[3]) == "PX" {
		milliSec, err := strconv.Atoi(cmd[4])
		if err != nil || milliSec <= 0 {
			return writeToConn(conn, Encode(ErrorString("invalid expire time")))
		}
		expiry = milliSec
	}
	store.SetValue(key, value, expiry)
	return writeToConn(conn, Encode(SimpleString("OK")))
}

func handleGet(cmd []string, conn net.Conn) error {
	if len(cmd) < 2 {
		return writeToConn(conn, Encode(ErrorString("wrong arguments for 'GET'")))
	}
	val, ok := store.GetValue(cmd[1])
	if !ok {
		return writeToConn(conn, Encode(nil))
	} else {
		return writeToConn(conn, Encode(val))
	}
}

func handleIncrement(cmd []string, conn net.Conn) error {
	if len(cmd) < 2 {
		return writeToConn(conn, Encode(ErrorString("wrong arguments for 'INCR'")))
	}
	key := cmd[1]
	val, err := store.Increment(key)
	if err != nil {
		return writeToConn(conn, Encode(ErrorString(err.Error())))
	}
	return writeToConn(conn, Encode(val))
}
