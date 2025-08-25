package common

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
)

func ParseCommand(conn net.Conn) ([]string, error) {
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
		return nil, fmt.Errorf("expected array start '*'")
	}

	elementCount, err := strconv.Atoi(arrayHeaderLine[1:])
	if err != nil {
		return nil, fmt.Errorf("invalid array length")
	}

	commandParts := make([]string, 0, elementCount)
	for range elementCount {
		stringHeaderLine, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		stringHeaderLine = strings.TrimSpace(stringHeaderLine)
		if len(stringHeaderLine) == 0 || stringHeaderLine[0] != '$' {
			return nil, fmt.Errorf("expected bulk string start '$'")
		}
		stringLength, err := strconv.Atoi(stringHeaderLine[1:])
		if err != nil {
			return nil, fmt.Errorf("invalid bulk string length")
		}
		stringData := make([]byte, stringLength+2)
		_, err = reader.Read(stringData)
		if err != nil {
			return nil, err
		}
		stringValue := string(stringData[:stringLength])

		commandParts = append(commandParts, stringValue)
	}
	fmt.Printf("[redis-cli] received %v\n", commandParts)
	return commandParts, nil
}
