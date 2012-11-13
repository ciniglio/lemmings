package tracker

import (
	"testing"
)

func Test_DecodeTracker_1(t *testing.T) {
	teststr := "d8:announce78:http://tracker.broadcasthe.net:34000/os6ch62cq8f1jftymy25e1gtpmigp0so/announce10:created by13:mktorrent 1.013:creation datei1352516454e4:infod6:lengthi1321388219e4:name42:Fringe.S05E06.720p.HDTV.X264-DIMENSION.mkv12:piece lengthi2097152e6:pieces20:ffffffffffffffffffffee"
	ti := ParseTorrentInfo([]byte(teststr))
	if ti.announce != "http://tracker.broadcasthe.net:34000/os6ch62cq8f1jftymy25e1gtpmigp0so/announce" {
		t.Error("Failed parsing announce: %s", ti.announce)
	}
	if ti.pieceLength != 2097152 {
		t.Error("Failed parsing Piece Length", ti.pieceLength)
	}
	if len(ti.pieces) != 1 {
		t.Error("Failed parsing number of pieces", ti.pieces)
	}
	if ti.pieces[0] != "ffffffffffffffffffff" {
		t.Error("Failed parsing pieces", ti.pieces)
	}
	if len(ti.files) != 1 {
		t.Error("Failed parsing number of files", ti.files)
	}
	if ti.files[0].length != 1321388219 {
		t.Error("Failed parsing file size", ti.files[0])
	}
}
