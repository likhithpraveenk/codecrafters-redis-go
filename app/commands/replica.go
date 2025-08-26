package commands

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/common"
	"github.com/codecrafters-io/redis-starter-go/app/store"
)

func HandleMasterConnection(conn net.Conn, port int) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	// --- Handshake ---
	handshakeSteps := [][]string{
		{"PING"},
		{"REPLCONF", "listening-port", fmt.Sprintf("%d", port)},
		{"REPLCONF", "capa", "psync2"},
		{"PSYNC", "?", "-1"},
	}

	for _, step := range handshakeSteps {
		conn.Write(common.Encode(step))
		if _, err := reader.ReadString('\n'); err != nil {
			fmt.Println("[replica] handshake failed:", err)
			return
		}
	}

	// --- Read RDB bulk string ---
	header, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("[replica] failed reading RDB header:", err)
		return
	}

	if strings.HasPrefix(header, "$") {
		lengthStr := strings.TrimSpace(header[1:])
		length, err := strconv.Atoi(lengthStr)
		if err != nil {
			fmt.Println("[replica] invalid RDB length:", err)
			return
		}

		buf := make([]byte, length)
		_, err = io.ReadFull(reader, buf)
		if err != nil {
			fmt.Println("[replica] failed reading RDB payload:", err)
			return
		}
	}

	store.MasterLinkStatus = "up"
	fmt.Println("[replica] handshake completed, link up")

	// --- Replication loop ---
	for {
		cmd, consumed, err := common.ParseCommand(reader)
		if err != nil {
			fmt.Println("[replica] master disconnected or parse error:", err)
			store.MasterLinkStatus = "down"
			return
		}
		if len(cmd) == 0 {
			continue
		}

		fmt.Printf("[replica] received: %+v\n", cmd)

		if err := handleReplicaCommand(cmd, conn); err != nil {
			fmt.Printf("[replica] error applying command: %v\n", err)
		}

		store.ReplicaOffset += consumed
	}
}

func handleReplicaCommand(cmd []string, conn net.Conn) error {
	cmdName := strings.ToUpper(cmd[0])

	switch cmdName {
	case "REPLCONF":
		if len(cmd) >= 2 && strings.ToUpper(cmd[1]) == "GETACK" {
			ack := []string{"REPLCONF", "ACK", fmt.Sprintf("%d", store.ReplicaOffset)}
			_, err := conn.Write(common.Encode(ack))
			if err != nil {
				return err
			}
			fmt.Printf("[replica] sent REPLCONF ACK %d\n", store.ReplicaOffset)
		}
		return nil

	case "PING":
		return nil

	default:
		if handler, ok := getHandler(cmdName); ok {
			_, err := handler(cmd)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
