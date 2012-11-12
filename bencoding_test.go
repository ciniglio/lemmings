package bencoding
import "testing"

func Test_DecodeString_1(t *testing.T) {
	testdata := map[string]string {
		"4:cows" : "cows",
		"5:ducks": "ducks",
		"3:hen"  : "hen",
		"2:me"   : "me",
	}

	for k, v := range testdata {
		s, _ := Bdecode([]byte(k))
		if (s.s != v) {
			t.Error("String decoding faileed")
		}
	}

	s, _ := Bdecode([]byte("4:cows"))
	if (s.s != "cows") {
		t.Error("String decoding failed")
	}
}