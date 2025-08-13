package commands

import (
	"net"
	"strings"
)

var commandHandlers = map[string]func([]string, net.Conn) error{}

func registerCommand(name string, handler func([]string, net.Conn) error) {
	commandHandlers[strings.ToUpper(name)] = handler
}

func GetHandler(cmd string) (func([]string, net.Conn) error, bool) {
	h, ok := commandHandlers[cmd]
	return h, ok
}

func writeToConn(conn net.Conn, resp []byte) error {
	_, err := conn.Write(resp)
	return err
}
