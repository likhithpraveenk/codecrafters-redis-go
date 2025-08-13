package commands

import (
	"net"

	"github.com/codecrafters-io/redis-starter-go/app/protocol"
	store "github.com/codecrafters-io/redis-starter-go/app/storage"
)

func handleType(cmd []string, conn net.Conn) error {
	if len(cmd) < 2 {
		return writeToConn(conn, protocol.EncodeError("wrong arguments for 'LLEN'"))
	}
	typ := store.GetType(cmd[1])
	return writeToConn(conn, protocol.EncodeSimpleString(typ))
}
