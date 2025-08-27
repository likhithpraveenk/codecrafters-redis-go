package commands

import (
	"bufio"
	"fmt"
	"net"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/common"
	"github.com/codecrafters-io/redis-starter-go/app/store"
)

func Init() {
	registerCommand("PING", handlePing)
	registerCommand("ECHO", handleEcho)
	registerCommand("SET", handleSet)
	registerCommand("INCR", handleIncrement)
	registerCommand("GET", handleGet)
	registerCommand("RPUSH", handleRPush)
	registerCommand("LPUSH", handleLPush)
	registerCommand("LRANGE", handleLRange)
	registerCommand("LLEN", handleLLen)
	registerCommand("LPOP", handleLPop)
	registerCommand("BLPOP", handleBLPop)
	registerCommand("TYPE", handleType)
	registerCommand("XADD", handleXAdd)
	registerCommand("XRANGE", handleXRange)
	registerCommand("XREAD", handleXRead)
	registerCommand("INFO", handleInfo)
	registerCommand("WAIT", handleWait)
	registerCommand("CONFIG", handleConfig)
	registerCommand("KEYS", handleKeys)
	registerCommand("SAVE", handleSave)
	registerCommand("PUBLISH", handlePublish)
	registerCommand("ZADD", handleZAdd)
	registerCommand("ZRANK", handleZRank)
	registerCommand("ZRANGE", handleZRange)
}

var commandHandlers = map[string]func([]string) (any, error){}

func registerCommand(name string, handler func([]string) (any, error)) {
	commandHandlers[strings.ToUpper(name)] = handler
}

func getHandler(cmd string) (func([]string) (any, error), bool) {
	h, ok := commandHandlers[cmd]
	return h, ok
}

func CentralHandler(conn net.Conn) {
	defer conn.Close()
	for {
		txn := store.GetTxnState(conn)
		r := bufio.NewReader(conn)
		cmd, _, err := common.ParseCommand(r)
		fmt.Printf("[redis-cli] received %v\n", cmd)
		if err != nil {
			fmt.Printf("Parse error: %v\n", err)
			return
		}
		if len(cmd) == 0 {
			conn.Write(common.Encode(common.SimpleError("empty command")))
			return
		}

		if handleReplica(conn, cmd) {
			continue
		}

		switch strings.ToUpper(cmd[0]) {
		case "MULTI", "EXEC", "DISCARD":
			handleTransaction(conn, cmd, txn)
			continue
		case "SUBSCRIBE":
			txn.Subscribed = true
		}

		if txn.InMulti {
			txn.QueuedCmds = append(txn.QueuedCmds, cmd)
			conn.Write(common.Encode(common.SimpleString("QUEUED")))
			continue
		}

		if txn.Subscribed {
			handleSubscription(conn, cmd, txn)
			continue
		}

		executeCommand(conn, cmd)
	}
}

func handleTransaction(conn net.Conn, cmd []string, txn *store.TxnState) {
	cmdName := strings.ToUpper(cmd[0])
	switch cmdName {
	case "MULTI":
		txn.InMulti = true
		txn.QueuedCmds = nil
		conn.Write(common.Encode(common.SimpleString("OK")))

	case "EXEC":
		if !txn.InMulti {
			conn.Write(common.Encode(common.SimpleError("ERR EXEC without MULTI")))
		} else {
			results := make([]any, 0, len(txn.QueuedCmds))
			for _, q := range txn.QueuedCmds {
				if handler, ok := getHandler(strings.ToUpper(q[0])); ok {
					result, err := handler(q)
					if err != nil {
						results = append(results, common.SimpleError(err.Error()))
					} else {
						results = append(results, result)
					}
				} else {
					results = append(results, common.SimpleError("ERR unknown command"))
				}
			}
			store.ClearTxnState(conn)
			conn.Write(common.Encode(results))
		}

	case "DISCARD":
		if !txn.InMulti {
			conn.Write(common.Encode(common.SimpleError("ERR DISCARD without MULTI")))
		} else {
			store.ClearTxnState(conn)
			conn.Write(common.Encode(common.SimpleString("OK")))
		}
	}
}

func executeCommand(conn net.Conn, cmd []string) {
	cmdName := strings.ToUpper(cmd[0])
	if handler, ok := getHandler(cmdName); ok {
		result, err := handler(cmd)
		if err != nil {
			conn.Write(common.Encode(common.SimpleError(err.Error())))
			return
		}

		conn.Write(common.Encode(result))

		if store.ReplicaRole == store.RoleMaster && isMutating(cmdName) {
			b := common.Encode(cmd)
			propagateToReplicas(b)
			store.MasterReplOffset += int64(len(b))
		}
	} else {
		conn.Write(common.Encode(common.SimpleError("ERR unknown command '" + cmd[0] + "'")))
	}
}

func isMutating(cmd string) bool {
	switch cmd {
	case "SET", "INCR", "RPUSH", "LPUSH", "LPOP", "BLPOP", "XADD":
		return true
	default:
		return false
	}
}
