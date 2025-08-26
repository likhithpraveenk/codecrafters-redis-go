package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/commands"
	"github.com/codecrafters-io/redis-starter-go/app/store"
)

func main() {
	cfg := loadConfig()
	store.ServerConfig = *cfg

	rdbPath := filepath.Join(cfg.Dir, cfg.DBFilename)
	if err := store.LoadRDB(rdbPath); err != nil {
		fmt.Printf("Failed to load RDB file %s: %v\n", rdbPath, err)
	}

	if cfg.ReplicaOf != "" {
		initReplication(cfg)
	}

	addr := fmt.Sprintf(":%d", cfg.Port)
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

func initReplication(cfg *store.Config) {
	parts := strings.Split(cfg.ReplicaOf, " ")
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
			commands.HandleMasterConnection(conn, cfg.Port)
			store.MasterLinkStatus = "down"
			conn.Close()
			time.Sleep(2 * time.Second)
		}
	}()
}

func loadConfig() *store.Config {
	port := flag.Int("port", 6379, "Port to listen on")
	replica := flag.String("replicaof", "", "Replication of master (host port)")
	dir := flag.String("dir", ".", "Directory for RDB files")
	dbfilename := flag.String("dbfilename", "dump.rdb", "RDB filename")
	flag.Parse()

	return &store.Config{
		Port:       *port,
		ReplicaOf:  *replica,
		Dir:        *dir,
		DBFilename: *dbfilename,
	}
}
