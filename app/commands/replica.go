package commands

import (
	"bufio"
	"fmt"
	"net"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/common"
	"github.com/codecrafters-io/redis-starter-go/app/store"
)

func HandleMasterConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	// 1. Handshake
	conn.Write(common.Encode([]string{"PING"}))
	if resp, err := reader.ReadString('\n'); err == nil {
		fmt.Printf("[replica] master replied: %s", resp)
	}

	conn.Write(common.Encode([]string{"REPLCONF", "listening-port", "0"}))
	if resp, err := reader.ReadString('\n'); err == nil {
		fmt.Printf("[replica] master replied: %s", resp)
	}

	conn.Write(common.Encode([]string{"REPLCONF", "capa", "psync2"}))
	if resp, err := reader.ReadString('\n'); err == nil {
		fmt.Printf("[replica] master replied: %s", resp)
	}

	conn.Write(common.Encode([]string{"PSYNC", "?", "-1"}))
	if resp, err := reader.ReadString('\n'); err == nil {
		fmt.Printf("[replica] PSYNC reply: %s", resp)
	}

	// 2. Master sends RDB bulk string
	header, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("[replica] failed reading RDB header:", err)
		return
	}
	fmt.Printf("[replica] received RDB header: %s", header)

	if strings.HasPrefix(header, "$") {
		fmt.Println("skipping rdb handling for now")
	}

	store.MasterLinkStatus = "up"
	fmt.Println("[replica] handshake completed, link up")

	// 3. Start replication loop
	for {
		cmd, err := common.ParseCommand(conn)
		if err != nil {
			fmt.Println("[replica] master disconnected or parse error:", err)
			store.MasterLinkStatus = "down"
			return
		}
		if len(cmd) == 0 {
			continue
		}

		fmt.Printf("[replica] applying command from master: %+v\n", cmd)

		if handler, ok := GetHandler(strings.ToUpper(cmd[0])); ok {
			_, err := handler(cmd)
			if err != nil {
				fmt.Printf("[replica] error applying command: %v\n", err)
			}
		}
	}
}

func PropagateToReplicas(cmd []string) {
	list := store.ListReplicaConns()
	for id, c := range list {
		_, err := c.Write(common.Encode(cmd))
		if err != nil {
			fmt.Printf("failed to propagate to replica %d: %v\n", id, err)
			c.Close()
			store.DeleteReplicaConn(id)
		}
	}
}
