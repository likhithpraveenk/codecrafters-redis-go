package store

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"time"
)

func SaveRDB(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)

	if _, err := w.Write([]byte("REDIS0011")); err != nil {
		return err
	}

	w.WriteByte(0xFE)
	writeLength(w, 0)
	w.WriteByte(0xFB)

	GlobalStore.mu.RLock()
	defer GlobalStore.mu.RUnlock()

	writeLength(w, uint64(len(GlobalStore.items)))
	var expCount uint64
	for _, it := range GlobalStore.items {
		if !it.expiresAt.IsZero() {
			expCount++
		}
	}
	writeLength(w, expCount)

	for key, it := range GlobalStore.items {
		if !it.expiresAt.IsZero() {
			w.WriteByte(0xFC)
			ts := uint64(it.expiresAt.UnixMilli())
			buf := make([]byte, 8)
			binary.LittleEndian.PutUint64(buf, ts)
			w.Write(buf)
		}
		switch it.typ {
		case TypeString:
			w.WriteByte(0x00)
			writeString(w, key)
			writeString(w, it.value.(string))
		default:
			return fmt.Errorf("unsupported type %v", it.typ)
		}
	}

	w.WriteByte(0xFF)
	w.Write(make([]byte, 8))

	return w.Flush()
}

func LoadRDB(path string) error {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer f.Close()

	r := bufio.NewReader(f)

	header := make([]byte, 9)
	io.ReadFull(r, header)
	if !bytes.Contains(header, []byte("REDIS")) {
		return fmt.Errorf("invalid rdb header")
	}

	for {
		b, err := r.ReadByte()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		switch b {
		case 0xFA: // metadata
			readString(r)
			readString(r)

		case 0xFE: // db selector
			readLength(r)

		case 0xFB: // hash table sizes
			readLength(r)
			readLength(r)

		case 0xFC: // expire ms
			buf := make([]byte, 8)
			io.ReadFull(r, buf)
			expireAt := time.UnixMilli(int64(binary.LittleEndian.Uint64(buf)))

			t, _ := r.ReadByte()
			if t != 0x00 {
				return fmt.Errorf("unsupported type %d", t)
			}
			key, _ := readString(r)
			val, _ := readString(r)

			GlobalStore.mu.Lock()
			GlobalStore.items[key] = Item{
				typ:       TypeString,
				value:     val,
				expiresAt: expireAt,
			}
			GlobalStore.mu.Unlock()

		case 0xFD: // expire s
			buf := make([]byte, 4)
			io.ReadFull(r, buf)
			expireAt := time.Unix(int64(binary.LittleEndian.Uint32(buf)), 0)

			t, _ := r.ReadByte()
			if t != 0x00 {
				return fmt.Errorf("unsupported type %d", t)
			}
			key, _ := readString(r)
			val, _ := readString(r)

			GlobalStore.mu.Lock()
			GlobalStore.items[key] = Item{
				typ:       TypeString,
				value:     val,
				expiresAt: expireAt,
			}
			GlobalStore.mu.Unlock()

		case 0x00: // string type
			key, err := readString(r)
			if err != nil {
				return err
			}
			val, err := readString(r)
			if err != nil {
				return err
			}

			GlobalStore.mu.Lock()
			GlobalStore.items[key] = Item{
				typ:   TypeString,
				value: val,
			}
			GlobalStore.mu.Unlock()

		case 0xFF: // EOF
			checksum := make([]byte, 8)
			io.ReadFull(r, checksum)
			return nil
		}
	}
	return nil
}

func writeLength(w io.Writer, n uint64) error {
	switch {
	case n < (1 << 6): // 6 bits
		return binary.Write(w, binary.BigEndian, uint8(n))
	case n < (1 << 14): // 14 bits
		b1 := uint8((n>>8)&0x3F) | 0x40
		b2 := uint8(n & 0xFF)
		_, err := w.Write([]byte{b1, b2})
		return err
	case n < (1 << 32): // 32 bits
		b := []byte{0x80}
		if _, err := w.Write(b); err != nil {
			return err
		}
		return binary.Write(w, binary.BigEndian, uint32(n))
	default:
		return fmt.Errorf("length too large")
	}
}

func writeString(w io.Writer, s string) error {
	if err := writeLength(w, uint64(len(s))); err != nil {
		return err
	}
	_, err := w.Write([]byte(s))
	return err
}

func readLength(r *bufio.Reader) (uint64, error) {
	first, err := r.ReadByte()
	if err != nil {
		return 0, err
	}
	enc := first >> 6
	switch enc {
	case 0: // 6-bit
		return uint64(first & 0x3F), nil
	case 1: // 14-bit
		second, err := r.ReadByte()
		if err != nil {
			return 0, err
		}
		return uint64(((uint16(first) & 0x3F) << 8) | uint16(second)), nil
	case 2: // 32-bit
		buf := make([]byte, 4)
		if _, err := io.ReadFull(r, buf); err != nil {
			return 0, err
		}
		return uint64(binary.BigEndian.Uint32(buf)), nil
	case 3:
		return 0, fmt.Errorf("special encoding not supported")
	}
	return 0, fmt.Errorf("bad length encoding")
}

func readString(r *bufio.Reader) (string, error) {
	n, err := readLength(r)
	if err != nil {
		return "", err
	}
	buf := make([]byte, n)
	if _, err := io.ReadFull(r, buf); err != nil {
		return "", err
	}
	return string(buf), nil
}
