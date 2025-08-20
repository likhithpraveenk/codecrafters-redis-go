package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/commands"
	store "github.com/codecrafters-io/redis-starter-go/app/storage"
)

func main() {
	port := flag.Int("port", 6379, "Port to listen on")
	replica := flag.String("replicaof", "", "Replication of master (host port)")
	flag.Parse()
	if *replica != "" {
		parts := strings.Split(*replica, " ")
		if len(parts) != 2 {
			fmt.Println("Invalid --replicaof argument, expected '<host> <port>'")
		}
		replicaOfHost := parts[0]
		replicaOfPort, err := strconv.Atoi(parts[1])
		if err != nil {
			fmt.Printf("Invalid replica port: %v\n", err)
		}
		store.ReplicaRole = store.RoleSlave
		store.MasterHost = replicaOfHost
		store.MasterPort = replicaOfPort
	}

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
