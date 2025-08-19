package commands

import (
	"net"

	"github.com/codecrafters-io/redis-starter-go/app/protocol"
	store "github.com/codecrafters-io/redis-starter-go/app/storage"
)

func handleXAdd(cmd []string, conn net.Conn) error {
	if len(cmd) < 5 || len(cmd)%2 != 1 {
		return writeToConn(conn, protocol.EncodeError("wrong number of arguments for 'XADD'"))
	}
	key := cmd[1]
	id := cmd[2]
	fields := make([]string, 0)
	for i := 3; i < len(cmd); i++ {
		fields = append(fields, cmd[i])
	}
	id, err := store.XAdd(key, id, fields)
	if err != nil {
		return writeToConn(conn, protocol.EncodeError(err.Error()))
	}
	return writeToConn(conn, protocol.EncodeBulkString(id))
}

func handleXRange(cmd []string, conn net.Conn) error {
	if len(cmd) < 4 {
		return writeToConn(conn, protocol.EncodeError("wrong number of arguments for 'XRANGE'"))
	}
	key, start, end := cmd[1], cmd[2], cmd[3]
	result, err := store.XRange(key, start, end)
	if err != nil {
		return writeToConn(conn, protocol.EncodeError(err.Error()))
	}
	return writeToConn(conn, protocol.EncodeNested(result))
}
