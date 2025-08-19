package commands

import (
	"net"
	"strings"
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
	registerCommand("LPOP", handleLPop)
	registerCommand("BLPOP", handleBLPop)
	registerCommand("TYPE", handleType)
	registerCommand("XADD", handleXAdd)
	registerCommand("XRANGE", handleXRange)
	registerCommand("XREAD", handleXRead)
}

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
