package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

func main() {
	l, err := net.Listen("tcp", ":6379")
	if err != nil {
		fmt.Printf("Failed to bind: %s", err)
		os.Exit(1)
	}
	fmt.Println("Listening on :6379")

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Printf("Accept error: %v\n", err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	r := bufio.NewReader(conn)
	for {
		args, err := parseRESP(r)
		if err != nil {
			fmt.Printf("Parse error: %v\n", err)
			return
		}

		if len(args) == 0 {
			continue
		}

		cmd := strings.ToUpper(args[0])
		switch cmd {
		case "PING":
			conn.Write(encodeSimpleString("PONG"))
		case "ECHO":
			if len(args) > 1 {
				conn.Write(encodeBulkString(args[1]))
			} else {
				conn.Write(encodeBulkString(""))
			}
		default:
			conn.Write(encodeError("unknown command '" + args[0] + "'"))
		}
	}
}

func parseRESP(r *bufio.Reader) ([]string, error) {
	line, err := r.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("reading array header: %w", err)
	}
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "*") {
		return nil, fmt.Errorf("invalid array header: %q", line)
	}
	numElems, err := strconv.Atoi(line[1:])
	if err != nil {
		return nil, fmt.Errorf("invalid array length: %w", err)
	}

	args := make([]string, numElems)
	for i := range numElems {
		lenLine, err := r.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("reading bulk length: %w", err)
		}

		length, err := strconv.Atoi(strings.TrimSpace(lenLine)[1:])
		if err != nil {
			return nil, fmt.Errorf("invalid bulk length: %w", err)
		}

		data := make([]byte, length+2)
		if _, err := r.Read(data); err != nil {
			return nil, fmt.Errorf("reading bulk data: %w", err)
		}
		args[i] = string(data[:length])
	}

	fmt.Printf("DEBUG: Parsed args: %v\n", args)
	return args, nil
}

func encodeSimpleString(s string) []byte {
	return fmt.Appendf(nil, "+%s\r\n", s)
}

func encodeError(msg string) []byte {
	return fmt.Appendf(nil, "-ERR %s\r\n", msg)
}

func encodeBulkString(s string) []byte {
	return fmt.Appendf(nil, "$%d\r\n%s\r\n", len(s), s)
}
