package common

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

func ParseCommand(reader *bufio.Reader) ([]string, int64, error) {
	arrayHeaderLine, err := reader.ReadString('\n')
	if err != nil {
		if err == io.EOF {
			return nil, 0, nil
		}
		return nil, 0, err
	}

	if len(arrayHeaderLine) == 0 || arrayHeaderLine[0] != '*' {
		return nil, 0, fmt.Errorf("expected array start '*'")
	}

	elementCount, err := strconv.Atoi(strings.TrimSpace(arrayHeaderLine[1:]))
	if err != nil {
		return nil, 0, fmt.Errorf("invalid array length")
	}
	totalSize := int64(len(arrayHeaderLine))

	commandParts := make([]string, 0, elementCount)
	for range elementCount {
		stringHeaderLine, err := reader.ReadString('\n')
		if err != nil {
			return nil, totalSize, err
		}
		totalSize += int64(len(stringHeaderLine))

		if len(stringHeaderLine) == 0 || stringHeaderLine[0] != '$' {
			return nil, totalSize, fmt.Errorf("expected bulk string start '$'")
		}

		stringLength, err := strconv.Atoi(strings.TrimSpace(stringHeaderLine[1:]))
		if err != nil {
			return nil, totalSize, fmt.Errorf("invalid bulk string length")
		}

		stringData := make([]byte, stringLength+2)
		if _, err := io.ReadFull(reader, stringData); err != nil {
			return nil, totalSize, err
		}

		totalSize += int64(len(stringData))
		stringValue := string(stringData[:stringLength])
		commandParts = append(commandParts, stringValue)
	}

	return commandParts, totalSize, nil
}
