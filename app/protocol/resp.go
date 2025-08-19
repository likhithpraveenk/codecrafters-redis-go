package protocol

import (
	"fmt"
	"strconv"
	"strings"

	store "github.com/codecrafters-io/redis-starter-go/app/storage"
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

func EncodeNullString() []byte {
	return fmt.Appendf(nil, "$-1\r\n")
}

func EncodeInteger(n int) []byte {
	i := strconv.Itoa(n)
	return fmt.Appendf(nil, ":%v\r\n", i)
}

func EncodeNested(list []store.StreamEntry) []byte {
	var b strings.Builder
	fmt.Fprintf(&b, "*%d\r\n", len(list))
	for _, entry := range list {
		fmt.Fprintf(&b, "*2\r\n")
		fmt.Fprintf(&b, "$%d\r\n%s\r\n", len(entry.ID), entry.ID)
		fmt.Fprintf(&b, "*%d\r\n", len(entry.Fields))
		for _, f := range entry.Fields {
			fmt.Fprintf(&b, "$%d\r\n%s\r\n", len(f), f)
		}
	}
	return []byte(b.String())
}

func EncodeArray(values []string) []byte {
	var b strings.Builder
	fmt.Fprintf(&b, "*%d\r\n", len(values))
	for _, val := range values {
		fmt.Fprintf(&b, "$%d\r\n%s\r\n", len(val), val)
	}
	return []byte(b.String())
}
