package commands

import (
	"fmt"
	"net"
	"strings"

	store "github.com/codecrafters-io/redis-starter-go/app/storage"
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
		cmd, err := parseCommand(conn)
		if err != nil {
			fmt.Printf("Parse error: %v\n", err)
			return
		}
		if len(cmd) == 0 {
			conn.Write(Encode(SimpleError("empty command")))
			return
		}

		cmdName := strings.ToUpper(cmd[0])

		txn := store.GetTxnState(conn)

		switch cmdName {
		case "MULTI":
			txn.InMulti = true
			txn.QueuedCmds = nil
			conn.Write(Encode(SimpleString("OK")))

		case "EXEC":
			if !txn.InMulti {
				conn.Write(Encode(SimpleError("ERR EXEC without MULTI")))
			} else {
				results := make([]any, 0, len(txn.QueuedCmds))
				for _, q := range txn.QueuedCmds {
					if handler, ok := GetHandler(strings.ToUpper(q[0])); ok {
						result, err := handler(q)
						if err != nil {
							results = append(results, SimpleError(err.Error()))
						} else {
							results = append(results, result)
						}
					} else {
						results = append(results, SimpleError("ERR unknown command"))
					}
				}
				store.ClearTxnState(conn)
				conn.Write(Encode(results))
			}

		case "DISCARD":
			if !txn.InMulti {
				conn.Write(Encode(SimpleError("ERR DISCARD without MULTI")))
			} else {
				store.ClearTxnState(conn)
				conn.Write(Encode(SimpleString("OK")))
			}

		default:
			if handler, ok := GetHandler(cmdName); ok {
				if txn.InMulti {
					txn.QueuedCmds = append(txn.QueuedCmds, cmd)
					conn.Write(Encode(SimpleString("QUEUED")))
				} else {
					store.ClearTxnState(conn)
					result, err := handler(cmd)
					if err != nil {
						conn.Write(Encode(SimpleError(err.Error())))
					} else {
						conn.Write(Encode(result))
					}
				}
			} else {
				conn.Write(Encode(SimpleError("ERR unknown command '" + cmd[0] + "'")))
			}
		}
	}
}
