package commands

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/common"
	"github.com/codecrafters-io/redis-starter-go/app/store"
)

func handleType(cmd []string) (any, error) {
	if len(cmd) < 2 {
		return nil, fmt.Errorf("wrong arguments for 'TYPE'")
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

func handleConfig(cmd []string) (any, error) {
	if cmd[1] == "GET" {
		return handleConfigGet(cmd)
	}
	return nil, fmt.Errorf("ERR wrong arguments for 'CONFIG' %v", cmd)
}

func handleSave(cmd []string) (any, error) {
	rdbPath := filepath.Join(store.ServerConfig.Dir, store.ServerConfig.DBFilename)

	if err := store.SaveRDB(rdbPath); err != nil {
		return nil, fmt.Errorf("ERR %v", err)
	}

	return common.SimpleString("OK"), nil
}

func handleConfigGet(cmd []string) (any, error) {
	dir := store.ServerConfig.Dir
	dbfilename := store.ServerConfig.DBFilename

	switch strings.ToLower(cmd[2]) {
	case "dir":
		return []string{"dir", dir}, nil
	case "dbfilename":
		return []string{"dbfilename", dbfilename}, nil
	}
	return []string{"dir", dir, "dbfilename", dbfilename}, nil
}

func handleKeys(cmd []string) (any, error) {
	pattern := cmd[1]
	keys := store.Keys(pattern)
	return keys, nil
}
