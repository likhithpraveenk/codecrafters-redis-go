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

func HandleReplica(conn net.Conn, cmd []string) bool {
	switch strings.ToUpper(cmd[0]) {
	case "REPLCONF":
		subCmd := strings.ToUpper(cmd[1])
		switch subCmd {
		case "LISTENING-PORT":
			replicaPort := cmd[2]
			host, _, _ := net.SplitHostPort(conn.RemoteAddr().String())
			r := store.AddReplica(host, replicaPort)
			store.RegisterReplicaConn(r.ID, conn)

		case "ACK":
			offset, _ := strconv.ParseInt(cmd[2], 10, 64)
			store.UpdateReplicaAck(conn, offset)
			fmt.Printf("[master] replica %v acknowledged offset %d\n", conn.RemoteAddr(), offset)
		case "CAPA":
		case "GETACK":
		}
		conn.Write(common.Encode(common.SimpleString("OK")))
		return true
	case "PSYNC":
		result := fmt.Sprintf("FULLRESYNC %s %d", store.MasterReplID, store.MasterReplOffset)
		conn.Write(common.Encode(common.SimpleString(result)))
		conn.Write(common.Encode(common.RDB(store.EmptyRDB)))
		return true
	}
	return false
}

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
		if resp, err := reader.ReadString('\n'); err == nil {
			fmt.Printf("[replica] master replied: %s", resp)
		} else {
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
	fmt.Printf("[replica] received RDB header: %s", header)

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
		fmt.Printf("[replica] consumed RDB payload of %d bytes\n", length)
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

		if err := HandleReplicaCommand(cmd, conn); err != nil {
			fmt.Printf("[replica] error applying command: %v\n", err)
		}

		store.ReplicaOffset += consumed
	}
}

func PropagateToReplicas(resp []byte) {
	list := store.ListReplicaConns()
	for id, c := range list {
		_, err := c.Write(resp)
		if err != nil {
			fmt.Printf("failed to propagate to replica %d: %v\n", id, err)
			c.Close()
			store.DeleteReplicaConn(id)
		}
	}
}

func HandleReplicaCommand(cmd []string, conn net.Conn) error {
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
		if handler, ok := GetHandler(cmdName); ok {
			_, err := handler(cmd)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
