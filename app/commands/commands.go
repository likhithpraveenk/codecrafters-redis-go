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
}

var commandHandlers = map[string]func([]string) (any, error){}

func registerCommand(name string, handler func([]string) (any, error)) {
	commandHandlers[strings.ToUpper(name)] = handler
}

func GetHandler(cmd string) (func([]string) (any, error), bool) {
	h, ok := commandHandlers[cmd]
	return h, ok
}

func CentralHandler(conn net.Conn) {
	defer conn.Close()
	for {
		txn := store.GetTxnState(conn)
		r := bufio.NewReader(conn)
		cmd, _, err := common.ParseCommand(r)
		if err != nil {
			fmt.Printf("Parse error: %v\n", err)
			return
		}
		if len(cmd) == 0 {
			conn.Write(common.Encode(common.SimpleError("empty command")))
			return
		}

		if HandleHandshake(conn, cmd) {
			continue
		}

		switch strings.ToUpper(cmd[0]) {
		case "MULTI", "EXEC", "DISCARD":
			HandleTransaction(conn, cmd, txn)
			continue
		}

		if txn.InMulti {
			txn.QueuedCmds = append(txn.QueuedCmds, cmd)
			conn.Write(common.Encode(common.SimpleString("QUEUED")))
		} else {
			ExecuteCommand(conn, cmd)
		}

	}
}

func HandleTransaction(conn net.Conn, cmd []string, txn *store.TxnState) {
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
				if handler, ok := GetHandler(strings.ToUpper(q[0])); ok {
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

func ExecuteCommand(conn net.Conn, cmd []string) {
	cmdName := strings.ToUpper(cmd[0])
	if handler, ok := GetHandler(cmdName); ok {
		result, err := handler(cmd)
		if err != nil {
			conn.Write(common.Encode(common.SimpleError(err.Error())))
			return
		}

		conn.Write(common.Encode(result))

		if store.ReplicaRole == store.RoleMaster && isMutating(cmdName) {
			fmt.Printf("[master] propagating command: %v\n", cmd)
			PropagateToReplicas(cmd)
		}

	} else {
		conn.Write(common.Encode(common.SimpleError("ERR unknown command '" + cmd[0] + "'")))
	}
}

func HandleHandshake(conn net.Conn, cmd []string) bool {
	switch strings.ToUpper(cmd[0]) {
	case "PING":
		conn.Write(common.Encode(common.SimpleString("PONG")))
		return true
	case "REPLCONF":
		if len(cmd) >= 3 && strings.ToLower(cmd[1]) == "listening-port" {
			replicaPort := cmd[2]
			host, _, _ := net.SplitHostPort(conn.RemoteAddr().String())
			r := store.AddReplica(host, replicaPort)
			store.RegisterReplicaConn(r.ID, conn)

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

func isMutating(cmd string) bool {
	switch cmd {
	case "SET", "INCR", "RPUSH", "LPUSH", "LPOP", "BLPOP", "XADD":
		return true
	default:
		return false
	}
}
