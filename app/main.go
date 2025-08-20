package main

import (
	"flag"
	"fmt"
	"net"
	"os"

	"github.com/codecrafters-io/redis-starter-go/app/commands"
)

func main() {
	port := flag.Int("port", 6379, "Port to listen on")
	flag.Parse()

	addr := fmt.Sprintf(":%d", *port)
	commands.Init()
	l, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Printf("Failed to listen on %s: %v\n", addr, err)
		os.Exit(1)
	}
	defer l.Close()

	fmt.Printf("[redis-cli] Listening on %s\n", addr)

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Printf("Failed to accept connection: %v\n", err)
			continue
		}
		go commands.CentralHandler(conn)
	}
}
