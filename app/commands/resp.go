package commands

import (
	"fmt"
	"strconv"
	"strings"
)

type SimpleString string
type ErrorString string

func Encode(value any) []byte {
	var b strings.Builder
	encodeValue(&b, value)
	return []byte(b.String())
}

func encodeValue(b *strings.Builder, value any) {
	switch v := value.(type) {
	case nil:
		b.WriteString("$-1\r\n")

	case SimpleString:
		b.WriteString("+")
		b.WriteString(string(v))
		b.WriteString("\r\n")

	case ErrorString:
		b.WriteString("-")
		b.WriteString(string(v))
		b.WriteString("\r\n")

	case int:
		b.WriteString(":")
		b.WriteString(strconv.Itoa(v))
		b.WriteString("\r\n")

	case string:
		b.WriteString("$")
		b.WriteString(strconv.Itoa(len(v)))
		b.WriteString("\r\n")
		b.WriteString(v)
		b.WriteString("\r\n")

	case []string:
		if v == nil {
			b.WriteString("$-1\r\n")
			return
		}
		b.WriteString("*")
		b.WriteString(strconv.Itoa(len(v)))
		b.WriteString("\r\n")
		for _, str := range v {
			encodeValue(b, str)
		}

	case []any:
		if v == nil {
			b.WriteString("$-1\r\n")
			return
		}
		b.WriteString("*")
		b.WriteString(strconv.Itoa(len(v)))
		b.WriteString("\r\n")
		for _, element := range v {
			encodeValue(b, element)
		}

	case [][]any:
		if v == nil {
			b.WriteString("$-1\r\n")
			return
		}
		b.WriteString("*")
		b.WriteString(strconv.Itoa(len(v)))
		b.WriteString("\r\n")
		for _, inner := range v {
			encodeValue(b, inner)
		}

	default:
		panic(fmt.Sprintf("unsupported type: %T", v))
	}
}
