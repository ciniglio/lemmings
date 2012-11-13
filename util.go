package tracker

import (
	"math/rand"
	"time"
)

func RandomBytes(n int) []byte {
	out := make([]byte, n)
	rand.Seed(time.Now().UTC().UnixNano())
	for i := 0; i < n; i++ {
		out[i] = byte(rand.Int())
	}
	return out
}
