package tracker

import (
	"io/ioutil"
)

type torrentFile struct {
	length int    //size of file in bytes
	path   string //filename
}

type TrackerInfo struct {
	announce    string        // tracker url
	name        string        // filename or dirname depending on length of files
	pieceLength int           // size of each piece
	pieces      []string      // checksums for each piece
	files       []torrentFile // info for each file
	numfiles    int           // not part of file, but helpful
}

func ReadTorrentFile(path string) *TrackerInfo {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return ParseTorrentInfo(b)
}

func ParseTorrentInfo(b []byte) *TrackerInfo {
	bi, _ := Bdecode(b)
	t := new(TrackerInfo)
	t.announce = bi.d["announce"].s

	info := bi.d["info"].d
	t.name = info["name"].s
	t.pieceLength = info["piece length"].i
	for i := 0; i < len(info["pieces"].s); i += 20 {
		t.pieces = append(t.pieces, info["pieces"].s[i:i+20])
	}

	if info["length"].i > 0 {
		f := new(torrentFile)
		f.length = info["length"].i
		f.path = t.name
		t.files = append(t.files, *f)
		t.numfiles = 1
	} else {
		t.numfiles = 0
		for _, v := range info["files"].l {
			f := new(torrentFile)
			f.length = v.d["length"].i
			f.path = v.d["path"].s
			t.files = append(t.files, *f)
			t.numfiles++
		}
	}

	return t
}
