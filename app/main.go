package main

import (
	"fmt"
	"net"
	"os"

	"github.com/codecrafters-io/redis-starter-go/app/commands"
)

func main() {
	commands.Init()
	l, err := net.Listen("tcp", ":6379")
	if err != nil {
		fmt.Errorf("Failed to bind: %s", err)
		os.Exit(1)
	}
	defer l.Close()
	fmt.Println("Listening on :6379")
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Errorf("Listener error: %v\n", err)
			continue
		}
		go commands.CentralHandler(conn)
	}
}
