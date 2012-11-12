package bencoding
import (
	"strconv"
	"bytes"
)

type bItem struct {
	s string
	i int64
	l *bList
	d *bDict
}

type bList struct {
	l []bItem
}

type bDict struct {
	d map[string]bItem
}


type bError struct {
}

func (e * bError) Error() string {
	return "bError happened"
}

func bdecodeInt(p []byte) int64 {
	i, _ := strconv.Atoi(string(p))
	return int64(i)
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
		end = bytes.IndexByte(p[2:], 'e')
		bi.i = bdecodeInt(p[2:end])
		progress = end + 1
	case 'l':
		bi.l = new(bList)
		bi.l.l = make([]bItem, 0)
		start := 2
		for i, q := Bdecode(p[start:]); q > 1; {
			bi.l.l = append(bi.l.l, i)
			start += q
		}
		progress = start + 1
	case 'd':
		bi.d = new(bDict)
		bi.d.d = make(map[string]bItem)
		start := 2
		for i, q := Bdecode(p[start:]); q > 1; {
			start += q
			j, r := Bdecode(p[start:])
			bi.d.d[i.s] = j
			start += r
		}
		progress = start + 1
	default:
		// String
		end = bytes.IndexByte(p[0:], ':')
		length := bdecodeInt(p[0:end])
		end ++
		bi.s = bdecodeString(p[end:end+int(length)])
		progress = end + int(length)
	}
	return *bi, progress
}
