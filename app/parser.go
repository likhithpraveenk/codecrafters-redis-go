package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
)

func parseCommand(conn net.Conn) ([]string, error) {
	reader := bufio.NewReader(conn)

	arrayHeaderLine, err := reader.ReadString('\n')
	if err != nil {
		if err == io.EOF {
			return nil, nil
		}
		return nil, err
	}

	arrayHeaderLine = strings.TrimSpace(arrayHeaderLine)
	if len(arrayHeaderLine) == 0 || arrayHeaderLine[0] != '*' {
		return nil, errors.New("expected array start '*'")
	}

	elementCount, err := strconv.Atoi(arrayHeaderLine[1:])
	if err != nil {
		return nil, errors.New("invalid array length")
	}

	commandParts := make([]string, 0, elementCount)
	for range elementCount {
		stringHeaderLine, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		stringHeaderLine = strings.TrimSpace(stringHeaderLine)
		if len(stringHeaderLine) == 0 || stringHeaderLine[0] != '$' {
			return nil, errors.New("expected bulk string start '$'")
		}
		stringLength, err := strconv.Atoi(stringHeaderLine[1:])
		if err != nil {
			return nil, errors.New("invalid bulk string length")
		}
		stringData := make([]byte, stringLength+2)
		_, err = reader.Read(stringData)
		if err != nil {
			return nil, err
		}
		stringValue := string(stringData[:stringLength])

		commandParts = append(commandParts, stringValue)
	}
	fmt.Printf("Received: %v\n", commandParts)
	return commandParts, nil
}
