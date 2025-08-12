package protocol

import (
	"fmt"
	"strconv"
)

func EncodeSimpleString(s string) []byte {
	return fmt.Appendf(nil, "+%s\r\n", s)
}

func EncodeError(msg string) []byte {
	return fmt.Appendf(nil, "-ERR %s\r\n", msg)
}

func EncodeBulkString(s string) []byte {
	return fmt.Appendf(nil, "$%d\r\n%s\r\n", len(s), s)
}

func EncodeInteger(n int) []byte {
	i := strconv.Itoa(n)
	return fmt.Appendf(nil, ":%v\r\n", i)
}
