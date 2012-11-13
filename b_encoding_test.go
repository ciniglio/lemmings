package tracker

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
		if s.i != int64(v) {
			t.Error("Int decoding failed:", v)
		}
	}
}

func Test_DecodeList_1(t *testing.T) {
	testdata := map[string][]bItem{
		"l3:cowe":      []bItem{bItem{s: "cow"}},
		"l3:cow3:hene": []bItem{bItem{s: "cow"}, bItem{s: "hen"}},
		"l3:cowi4ee":   []bItem{bItem{s: "cow"}, bItem{i: 4}},
		"li5ei4ei3ee":  []bItem{bItem{i: 5}, bItem{i: 4}, bItem{i: 3}},
		"l:3:cowi4el:i5eee": []bItem{bItem{s: "cow"}, bItem{i: 4},
			bItem{l: []bItem{bItem{i: 5}}}},
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
	s := "d3:cowli4ei-1eee"
	b, _ := Bdecode([]byte(s))
	if b.d["cow"].l[0].i != 4 {
		t.Error("Failed dict decoding")
	}
	if b.d["cow"].l[1].i != -1 {
		t.Error("Failed dict decoding")
	}
}

func Test_DecodeDict_2(t *testing.T) {
	s := "d8:announce4:kanne"
	b, _ := Bdecode([]byte(s))
	if b.d["announce"].s != "kann" {
		t.Error("Failed dict decoding")
	}
}

func Test_EncodeInt_1(t *testing.T) {
	testdata := map[int64]string{
		0:          "i0e",
		12:         "i12e",
		1241255125: "i1241255125e",
	}
	for k, v := range testdata {
		b := bItem{i: k}
		if string(Bencode(b)) != v {
			t.Error("Failed int encoding", k)
		}
	}
}

func Test_EncodeStr_1(t *testing.T) {
	testdata := map[string]string{
		"hen":       "3:hen",
		"chicken":   "7:chicken",
		"umpteenth": "9:umpteenth",
		"5tgb":      "4:5tgb",
		"5:5:5":     "5:5:5:5",
	}
	for k, v := range testdata {
		b := bItem{s: k}
		if string(Bencode(b)) != v {
			t.Error("Failed str encoding", k)
		}
	}
}

func Test_EncodeList_1(t *testing.T) {
	testdata := map[string][]bItem{
		"l3:cowe":      []bItem{bItem{s: "cow"}},
		"l3:cow3:hene": []bItem{bItem{s: "cow"}, bItem{s: "hen"}},
		"l3:cowi4ee":   []bItem{bItem{s: "cow"}, bItem{i: 4}},
		"li5ei4ei3ee":  []bItem{bItem{i: 5}, bItem{i: 4}, bItem{i: 3}},
		"l3:cowi4eli5eee": []bItem{bItem{s: "cow"}, bItem{i: 4},
			bItem{l: []bItem{bItem{i: 5}}}},
	}
	for k, v := range testdata {
		b := bItem{l: v}
		if string(Bencode(b)) != k {
			t.Error("Failed list encoding", k, string(Bencode(b)))
		}
	}
}

func Test_EncodeDict_1(t *testing.T) {
	s := "d3:cowli4ei-1eee"
	b := bItem{d: map[string]bItem{
		"cow": bItem{l: []bItem{bItem{i: 4},
			bItem{i: -1}}}}}
	v := string(Bencode(b))
	if v != s {
		t.Error("Failed dict encoding", v, s)
	}
}

func Test_EncodeDict_2(t *testing.T) {
	s := "d8:announce4:kanne"
	b := bItem{d: map[string]bItem{
		"announce": bItem{s: "kann"}}}
	v := string(Bencode(b))
	if v != s {
		t.Error("Failed dict encoding", v, s)
	}
}
