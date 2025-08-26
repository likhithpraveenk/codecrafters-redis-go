package commands

import (
	"fmt"
	"net"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/common"
	"github.com/codecrafters-io/redis-starter-go/app/store"
)

func handleSubscription(conn net.Conn, cmd []string, txn *store.TxnState) {
	client := store.GetClient(conn)

	cmdName := strings.ToUpper(cmd[0])
	switch cmdName {
	case "SUBSCRIBE":
		for _, ch := range cmd[1:] {
			result := store.Subscribe(ch, client)
			client.Messages <- result
		}

	case "UNSUBSCRIBE":
		for _, ch := range cmd[1:] {
			result := store.UnSubscribe(ch, client)
			client.Messages <- result
			txn.Subscribed = false
		}

	case "PING":
		conn.Write(common.Encode([]string{"pong", ""}))

	case "QUIT":
		store.RemoveClient(conn)
		conn.Close()
		return

	case "PSUBSCRIBE", "PUNSUBSCRIBE", "RESET":
	default:
		err := fmt.Sprintf("ERR Can't execute '%s': only (P|S)SUBSCRIBE / (P|S)UNSUBSCRIBE / PING / QUIT / RESET are allowed in this context", cmd[0])
		conn.Write(common.Encode(common.SimpleError(err)))
	}
}

func handlePublish(cmd []string) (any, error) {
	channel, message := cmd[1], cmd[2]
	result := store.Publish(channel, message)
	return result, nil
}
