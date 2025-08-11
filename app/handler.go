package main

import (
	"fmt"
	"net"
	"strings"
)

func handleConnection(conn net.Conn) {
	for {
		cmd, err := parseCommand(conn)
		if err != nil {
			fmt.Printf("Parse error: %v\n", err)
			return
		}
		handleCommand(cmd, conn)
	}
}

func handleCommand(cmd []string, conn net.Conn) error {
	if len(cmd) == 0 {
		conn.Write(encodeError("empty command"))
		return nil
	}

	switch strings.ToUpper(cmd[0]) {
	case "PING":
		conn.Write(encodeSimpleString("PONG"))
	case "ECHO":
		if len(cmd) < 2 {
			conn.Write(encodeError("wrong arguments for 'ECHO'"))
			return nil
		}
		conn.Write(encodeBulkString(cmd[1]))
	case "SET":
		if len(cmd) < 3 {
			conn.Write(encodeError("wrong arguments for 'SET'"))
			return nil
		}
		setValue(cmd[1], cmd[2])
		conn.Write(encodeSimpleString("OK"))
	case "GET":
		if len(cmd) < 2 {
			conn.Write(encodeError("wrong arguments for 'GET'"))
			return nil
		}
		val, ok := getValue(cmd[1])
		if !ok {
			conn.Write([]byte("$-1\r\n"))

		} else {
			conn.Write(encodeBulkString(val))
		}

	default:
		conn.Write(encodeError("unknown command '" + cmd[0] + "'"))
	}

	return nil
}
