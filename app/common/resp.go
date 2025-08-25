package common

import (
	"fmt"
	"strconv"
	"strings"
)

type SimpleString string
type SimpleError string
type RDB []byte

func Encode(value any) []byte {
	var b strings.Builder
	encodeValue(&b, value)
	return []byte(b.String())
}

func encodeValue(b *strings.Builder, value any) {
	switch v := value.(type) {
	case nil:
		b.WriteString("$-1\r\n")

	case RDB:
		b.WriteString("$")
		b.WriteString(strconv.Itoa(len(v)))
		b.WriteString("\r\n")
		b.Write(v)

	case SimpleString:
		b.WriteString("+")
		b.WriteString(string(v))
		b.WriteString("\r\n")

	case SimpleError:
		b.WriteString("-")
		b.WriteString(string(v))
		b.WriteString("\r\n")

	case int64:
		b.WriteString(":")
		b.WriteString(strconv.FormatInt(v, 10))
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
