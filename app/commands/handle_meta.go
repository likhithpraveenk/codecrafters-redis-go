package commands

import (
	"net"

	store "github.com/codecrafters-io/redis-starter-go/app/storage"
)

func handleType(cmd []string, conn net.Conn) error {
	if len(cmd) < 2 {
		return writeToConn(conn, Encode(ErrorString("wrong arguments for 'LLEN'")))
	}
	typ := store.GetType(cmd[1])
	return writeToConn(conn, Encode(SimpleString(typ)))
}
