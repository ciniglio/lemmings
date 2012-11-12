package bencoding

import (
	"reflect"
	"testing"
)

func Test_DecodeString_1(t *testing.T) {
	testdata := map[string]string{
		"4:cows":  "cows",
		"5:ducks": "ducks",
		"3:hen":   "hen",
		"2:me":    "me",
	}

	for k, v := range testdata {
		s, _ := Bdecode([]byte(k))
		if s.s != v {
			t.Error("String decoding faileed")
		}
	}

	s, p := Bdecode([]byte("4:cows"))
	if s.s != "cows" {
		t.Error("String decoding failed")
	}
	if p != 6 {
		t.Error("String index failed")
	}
}

func Test_DecodeInt_1(t *testing.T) {
	testdata := map[string]int{
		"i3e":  3,
		"i23e": 23,
		"i-1e": -1,
		"i9e":  9,
	}

	for k, v := range testdata {
		s, _ := Bdecode([]byte(k))
		if s.i != v {
			t.Error("Int decoding failed:", v)
		}
	}
}

func Test_DecodeList_1(t *testing.T) {
	testdata := map[string][]bItem{
		"l:3:cowe":      []bItem{bItem{s: "cow"}},
		"l:3:cow3:hene": []bItem{bItem{s: "cow"}, bItem{s: "hen"}},
		"l:3:cowi4ee":   []bItem{bItem{s: "cow"}, bItem{i: 4}},
		"l:i5ei4ei3ee":  []bItem{bItem{i: 5}, bItem{i: 4}, bItem{i: 3}},
		// "l:3:cowi4el:i5eee": []bItem{bItem{s: "cow"}, bItem{i: 4},
		// 	bItem{l: []bItem{bItem{i: 5}}}},
	}
	for k, v := range testdata {
		s, _ := Bdecode([]byte(k))
		for j, u := range s.l {
			if !(reflect.DeepEqual(u, v[j])) {
				t.Error("Failed List")
			}
		}
	}
}

func Test_DecodeDict_1(t *testing.T) {
	s := "d:3:cowl:i4ei-1eee"
	b, _ := Bdecode([]byte(s))
	if b.d["cow"].l[0].i != 4 {
		t.Error("Failed dict decoding")
	}
	if b.d["cow"].l[1].i != -1 {
		t.Error("Failed dict decoding")
	}

}
