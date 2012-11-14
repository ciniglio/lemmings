package tracker

import (
	"bytes"
	"strconv"
)

type bItem struct {
	s   string
	i   int64
	l   []bItem
	d   map[string]bItem
	raw []byte
}

type bError struct {
}

func (e *bError) Error() string {
	return "bError happened"
}

func bdecodeInt(p []byte) int64 {
	i, _ := strconv.Atoi(string(p))
	return int64(i)
}

func bdecodeString(p []byte) string {
	return string(p)
}

func Bdecode(p []byte) (*bItem, int) {
	bi := new(bItem)
	progress := 0
	end := 0

	switch p[0] {
	case 'e':
		progress = 1
	case 'i':
		end = bytes.IndexByte(p[0:], 'e')
		bi.i = bdecodeInt(p[1:end])
		progress = end + 1
	case 'l':
		bi.l = (make([]bItem, 0))
		start := 1
		q := 0
		var i *bItem
		for {
			i, q = Bdecode(p[start:])
			if q == 1 {
				break
			}
			bi.l = append(bi.l, *i)
			start += q
		}
		progress = start
	case 'd':
		bi.d = make(map[string]bItem)
		start := 1
		for {
			i, q := Bdecode(p[start:])
			if q == 1 {
				break
			}
			start += q
			j, r := Bdecode(p[start:])
			bi.d[i.s] = *j
			start += r
		}
		progress = start + 1
	default:
		// String
		end = bytes.IndexByte(p[0:], ':')
		length := bdecodeInt(p[0:end])
		end++
		bi.s = bdecodeString(p[end : end+int(length)])
		progress = end + int(length)
	}
	bi.raw = p[0:progress]
	return bi, progress
}

func bencodeInt(i int64) []byte {
	out := []byte{}
	out = append(out, byte('i'))
	out = append(out, strconv.Itoa(int(i))...)
	out = append(out, byte('e'))
	return out
}

func bencodeString(s string) []byte {
	out := []byte{}
	out = append(out, strconv.Itoa(len(s))...)
	out = append(out, ':')
	out = append(out, s...)
	return out
}

func Bencode(b bItem) []byte {
	out := []byte{}
	switch {
	case len(b.d) > 0:
		out = append(out, byte('d'))
		for k, v := range b.d {
			out = append(out, bencodeString(k)...)
			out = append(out, Bencode(v)...)
		}
		out = append(out, byte('e'))
	case len(b.l) > 0:
		out = append(out, byte('l'))
		for _, i := range b.l {
			out = append(out, Bencode(i)...)
		}
		out = append(out, byte('e'))
	case len(b.s) > 0:
		out = append(out, bencodeString(b.s)...)
	default:
		out = append(out, bencodeInt(b.i)...)
	}
	return out
}
