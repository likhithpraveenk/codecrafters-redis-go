package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/commands"
	"github.com/codecrafters-io/redis-starter-go/app/store"
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
		replicaOfHost, replicaOfPort := parts[0], parts[1]

		store.ReplicaRole = store.RoleSlave
		store.MasterHost = replicaOfHost
		store.MasterPort = replicaOfPort

		go func() {
			for {
				addr := net.JoinHostPort(replicaOfHost, replicaOfPort)
				conn, err := net.Dial("tcp", addr)
				if err != nil {
					fmt.Printf("Failed to connect to master %s: %v\n", addr, err)
					time.Sleep(2 * time.Second)
					continue
				}

				fmt.Printf("Connected to master %s\n", addr)
				store.MasterLinkStatus = "up"
				commands.HandleMasterConnection(conn)
				store.MasterLinkStatus = "down"
				conn.Close()
				time.Sleep(2 * time.Second)
			}
		}()
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
