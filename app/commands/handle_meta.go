package commands

import (
	"fmt"

	store "github.com/codecrafters-io/redis-starter-go/app/storage"
)

func handleType(cmd []string) (any, error) {
	if len(cmd) < 2 {
		return nil, fmt.Errorf("wrong arguments for 'LLEN'")
	}
	typ := store.GetType(cmd[1])
	return SimpleString(typ), nil
}
