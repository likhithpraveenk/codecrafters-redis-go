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
	fields := make(map[string]string)
	for i := 3; i < len(cmd); i += 2 {
		fields[cmd[i]] = cmd[i+1]
	}
	id, err := store.XAdd(key, id, fields)
	if err != nil {
		return writeToConn(conn, protocol.EncodeError(err.Error()))
	}
	return writeToConn(conn, protocol.EncodeBulkString(id))
}
