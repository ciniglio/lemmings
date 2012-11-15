package tracker

import (
	"math/rand"
	"time"
	"encoding/binary"
	"bytes"
	"fmt"
)

func RandomBytes(n int) []byte {
	out := make([]byte, n)
	rand.Seed(time.Now().UTC().UnixNano())
	for i := 0; i < n; i++ {
		out[i] = byte(rand.Int())
	}
	return out
}

func toInt64(b []byte) int64 {
	var i int64
	if len(b) < 8 {
		tmp := make([]byte, 8 - len(b))
		tmp = append(tmp, b...)
		b = tmp
	}
	err := binary.Read(bytes.NewReader(b), binary.BigEndian, &i)
	if err != nil {
		fmt.Println("Error converting to int: ", err)
	}
	return i
}

func toInt(b []byte) int {
	return int(toInt64(b))
}