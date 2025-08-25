package commands

import (
	"fmt"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/common"
	"github.com/codecrafters-io/redis-starter-go/app/store"
)

func handleType(cmd []string) (any, error) {
	if len(cmd) < 2 {
		return nil, fmt.Errorf("wrong arguments for 'LLEN'")
	}
	typ := store.GetType(cmd[1])
	return common.SimpleString(typ), nil
}

func handleInfo(cmd []string) (any, error) {
	section := ""
	if len(cmd) > 1 {
		section = strings.ToLower(cmd[1])
	}
	var out string
	switch section {
	case "", "replication":
		out = store.Info()
	default:
		out = "# " + section + "\r\n"
	}

	return out, nil
}
