package bencoding

import (
	"bytes"
	"strconv"
)

type bItem struct {
	s string
	i int
	l []bItem
	d map[string]bItem
}

type bError struct {
}

func (e *bError) Error() string {
	return "bError happened"
}

func bdecodeInt(p []byte) int {
	i, _ := strconv.Atoi(string(p))
	return int(i)
}

func bdecodeString(p []byte) string {
	return string(p)
}

func Bdecode(p []byte) (bItem, int) {
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
		start := 2
		q := 0
		var i bItem
		for {
			i, q = Bdecode(p[start:])
			if q == 1 {
				break
			}
			bi.l = append(bi.l, i)
			start += q
		}
		progress = start
	case 'd':
		bi.d = make(map[string]bItem)
		start := 2
		for {
			i, q := Bdecode(p[start:])
			if q == 1 {
				break
			}
			start += q
			j, r := Bdecode(p[start:])
			bi.d[i.s] = j
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
	return *bi, progress
}
