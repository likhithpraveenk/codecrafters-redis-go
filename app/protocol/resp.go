package protocol

import (
	"fmt"
	"strconv"
	"strings"
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

func EncodeArray(values []string) []byte {
	var b strings.Builder
	fmt.Fprintf(&b, "*%d\r\n", len(values))
	for _, val := range values {
		fmt.Fprintf(&b, "$%d\r\n%s\r\n", len(val), val)
	}
	return []byte(b.String())
}
