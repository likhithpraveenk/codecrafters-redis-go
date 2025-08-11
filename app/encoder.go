package main

import "fmt"

func encodeSimpleString(s string) []byte {
	return fmt.Appendf(nil, "+%s\r\n", s)
}

func encodeError(msg string) []byte {
	return fmt.Appendf(nil, "-ERR %s\r\n", msg)
}

func encodeBulkString(s string) []byte {
	return fmt.Appendf(nil, "$%d\r\n%s\r\n", len(s), s)
}
