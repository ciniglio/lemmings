package tracker

import (
	"testing"
)

func Test_IntConversion_1(test *testing.T) {
	testdata := map[string]int {
		"\x00\x00": 0,
		"\x00\x01": 1,
		"\x00\x10": 16,
		"\x01\x00": 256,
		"\x01\x01": 257,
	}

	for k,v := range testdata {
		i := toInt([]byte(k))
		if i != v {
			test.Error("Failed int conversion", k, v)
		}
	}
}