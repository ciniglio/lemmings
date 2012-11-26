package tracker

import (
	"fmt"
	"strings"
	"testing"
)

func Test_TGR_1(test *testing.T) {
	t := new(TorrentInfo)
	t.info_hash = "hash"
	t.client_id = "id"
	t.announce = "http://example.com"
	tgr := NewTrackerProxy(t)

	s := tgr.tgr.GenerateGetString()
	if (s != "http://example.com?info_hash=hash&peer_id=id") &&
		(s != "http://example.com?peer_id=id&info_hash=hash") {
		test.Error("failed at generating query string", s)
	}
}

func Test_Escaping(test *testing.T) {
	t := new(TorrentInfo)
	b := make([]byte, 20)
	in := strings.NewReader("123456789abcdef123456789abcdef123456789a")
	fmt.Fscanf(in, "%x", &b)
	t.info_hash = string(b)
	t.client_id = "1"
	t.announce = "http://example.com"
	tgr := NewTrackerProxy(t)

	s := tgr.tgr.GenerateGetString()
	if (s != "http://example.com?info_hash=%124Vx%9A%BC%DE%F1%23Eg%89%AB%CD%EF%124Vx%9A&peer_id=1") &&
		(s != "http://example.com?peer_id=1&info_hash=%124Vx%9A%BC%DE%F1%23Eg%89%AB%CD%EF%124Vx%9A") {
		test.Error("failed at generating query string", s)
	}
}

func Test_Connection_1(test *testing.T) {
	ti, _ := ReadTorrentFile("test/test.torrent")
	if ti.announce != "http://thomasballinger.com:6969/announce" ||
		len(ti.files) != 1 || ti.files[0].length != 1751391 {
		test.Error("Failed at parsing file", ti.announce)
	}
	tgr := NewTrackerProxy(ti)
	tgr.tgr.MakeTrackerRequest()
}
