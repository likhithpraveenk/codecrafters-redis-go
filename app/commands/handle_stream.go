package commands

import (
	"net"

	store "github.com/codecrafters-io/redis-starter-go/app/storage"
)

func handleXAdd(cmd []string, conn net.Conn) error {
	if len(cmd) < 5 || len(cmd)%2 != 1 {
		return writeToConn(conn, Encode(ErrorString("wrong number of arguments for 'XADD'")))
	}
	key := cmd[1]
	id := cmd[2]
	fields := make([]string, 0)
	for i := 3; i < len(cmd); i++ {
		fields = append(fields, cmd[i])
	}
	id, err := store.XAdd(key, id, fields)
	if err != nil {
		return writeToConn(conn, Encode(ErrorString(err.Error())))
	}
	return writeToConn(conn, Encode(id))
}

func handleXRange(cmd []string, conn net.Conn) error {
	if len(cmd) < 4 {
		return writeToConn(conn, Encode(ErrorString("wrong number of arguments for 'XRANGE'")))
	}
	key, start, end := cmd[1], cmd[2], cmd[3]
	result, err := store.XRange(key, start, end)
	if err != nil {
		return writeToConn(conn, Encode(ErrorString(err.Error())))
	}
	return writeToConn(conn, Encode(result))
}

func handleXRead(cmd []string, conn net.Conn) error {
	if len(cmd) < 4 || len(cmd)%2 != 0 {
		return writeToConn(conn, Encode(ErrorString("wrong number of arguments for 'XREAD'")))
	}
	keysIds := cmd[2:]
	keysLen := len(keysIds) / 2
	keys, ids := []string{}, []string{}
	for i, s := range keysIds {
		if i < keysLen {
			keys = append(keys, s)
		} else {
			ids = append(ids, s)
		}
	}
	result, err := store.XRead(keys, ids)
	if err != nil {
		return writeToConn(conn, Encode(ErrorString(err.Error())))
	}
	return writeToConn(conn, Encode(result))
}
