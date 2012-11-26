/******************************************************************************/
/**                                                                          **/
/**       b_encoding                                                         **/
/**   Spec: http://www.bittorrent.org/beps/bep_0003.html                     **/
/**                                                                          **/
/******************************************************************************/

package tracker

import (
	"bytes"
	"strconv"
)

// this is what a bencoding is made up of, it can nest itself as well
// only one of the instance objects should exist for each bItem
type bItem struct {
	s   string           // string type
	i   int64            // integer type
	l   []bItem          // list type
	d   map[string]bItem // dictionary strings => bItems
	raw []byte           // this contains the raw string that was decoded into this bItem
}

type bError struct {
}

func (e *bError) Error() string {
	return "bError happened"
}

// if an integer, let's just convert it from a string
func bdecodeInt(p []byte) (int64, error) {
	i, err := strconv.Atoi(string(p))
	if err != nil {
		return 0, new(bError)
	}
	return int64(i), nil
}

// if a string, let's return the string
func bdecodeString(p []byte) string {
	return string(p)
}

// assumes that p corresponds to exactly 1 bItem. 
// e.g. i8ei9e would only return 8, li8ei9ee would pass
// recursive function
func Bdecode(p []byte) (*bItem, int) {
	bi := new(bItem)
	progress := 0
	end := 0
	switch p[0] {
	case 'e':
		// reached the end of an item, do nothing except advance
		progress = 1
	case 'i':
		// number goes until the first e
		end = bytes.IndexByte(p[0:], 'e')
		var err error
		bi.i, err = bdecodeInt(p[1:end])
		if err != nil {
			return nil, -1
		}
		progress = end + 1
	case 'l':
		// parse list until we reach empty item
		bi.l = (make([]bItem, 0))
		start := 1
		q := 0
		var i *bItem
		for {
			i, q = Bdecode(p[start:])
			if i == nil {
				return nil, -1
			}
			if q == 1 {
				// reached empty item; stop
				break
			}
			bi.l = append(bi.l, *i)
			start += q
		}
		progress = start
	case 'd':
		// parse dictionary until we reach empty item
		bi.d = make(map[string]bItem)
		start := 1
		for {
			i, q := Bdecode(p[start:])
			if i == nil {
				return nil, -1
			}
			if q == 1 {
				break
			}
			start += q
			j, r := Bdecode(p[start:])
			if j == nil {
				return nil, -1
			}
			// dict[i.string] = j
			bi.d[i.s] = *j
			start += r
		}
		progress = start + 1
	default:
		// String
		end = bytes.IndexByte(p[0:], ':')
		length, err := bdecodeInt(p[0:end])
		if err != nil {
			return nil, -1
		}
		end++
		bi.s = bdecodeString(p[end : end+int(length)])
		progress = end + int(length)
	}
	bi.raw = p[0:progress]
	return bi, progress
}

////////////////////////////////////////////////////////////////////////////////
///////
/////// Encoding functions
///////
////////////////////////////////////////////////////////////////////////////////

func bencodeInt(i int64) []byte {
	// straightforward iXe (x is int)
	out := []byte{}
	out = append(out, byte('i'))
	out = append(out, strconv.Itoa(int(i))...)
	out = append(out, byte('e'))
	return out
}

func bencodeString(s string) []byte {
	// straightforward len:string
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
		// if we have a dict, lets prefix with a d,
		out = append(out, byte('d'))

		// then encode all of it's members
		for k, v := range b.d {
			out = append(out, bencodeString(k)...)
			out = append(out, Bencode(v)...)
		}

		// then e as suffix
		out = append(out, byte('e'))
	case len(b.l) > 0:
		//if we have a list, we prefix with l
		out = append(out, byte('l'))
		// then encode all members
		for _, i := range b.l {
			out = append(out, Bencode(i)...)
		}
		// then suffix with e
		out = append(out, byte('e'))
	case len(b.s) > 0:
		// trivial case -> encode string
		out = append(out, bencodeString(b.s)...)
	default:
		// trivial case -> encode int
		out = append(out, bencodeInt(b.i)...)
	}
	return out
}
