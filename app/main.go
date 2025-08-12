package main

import (
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/commands"
	"github.com/codecrafters-io/redis-starter-go/app/protocol"
)

func main() {
	commands.Init()
	l, err := net.Listen("tcp", ":6379")
	if err != nil {
		fmt.Printf("Failed to bind: %s", err)
		os.Exit(1)
	}
	defer l.Close()
	fmt.Println("Listening on :6379")
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Printf("Accept error: %v\n", err)
			continue
		}
		go func(c net.Conn) {
			defer c.Close()
			for {
				cmd, err := parseCommand(c)
				if err != nil {
					fmt.Printf("Parse error: %v\n", err)
					return
				}
				if len(cmd) == 0 {
					c.Write(protocol.EncodeError("empty command"))
					return
				}

				cmdName := strings.ToUpper(cmd[0])
				if handler, ok := commands.GetHandler(cmdName); ok {
					if err := handler(cmd, c); err != nil {
						fmt.Printf("Command handler error: %v\n", err)
						return
					}
				} else {
					c.Write(protocol.EncodeError("unknown command '" + cmd[0] + "'"))
				}
			}
		}(conn)
	}
}
