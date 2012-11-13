package tracker

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

type TrackerGetRequest struct {
	info_hash    string
	peer_id      string
	port         string
	ip_addr      string //optional
	uploaded     int
	downloaded   int
	left         int
	compact      int    // 0 or 1
	no_peer_id   int    // 0 or 1, optional
	event        string // started, stopped or completed (can be blank)
	numwant      int    // optional
	key          string // optional
	tracker_id   string // optional
	torrent_info *TorrentInfo
}

type torrentPeer struct {
	peer_id string
	ip      string
	port    int
}

type TrackerResponse struct {
	failure_reason string
	interval       int
	min_interval   int
	tracker_id     string
	complete       int
	incomplete     int
	peers          []torrentPeer
}

func CreateTrackerGetRequest(t *TorrentInfo) *TrackerGetRequest {
	tgr := new(TrackerGetRequest)
	tgr.info_hash = t.info_hash
	tgr.peer_id = t.client_id
	tgr.torrent_info = t
	return tgr
}

func (t *TrackerGetRequest) GenerateGetString() string {
	u, _ := url.Parse(t.torrent_info.announce)
	v := url.Values{}
	v.Set("info_hash", t.info_hash)
	v.Set("peer_id", t.peer_id)
	u.RawQuery = v.Encode()
	return u.String()
}

func (t *TrackerGetRequest) MakeTrackerRequest() {
	q := t.GenerateGetString()
	res, err := http.Get(q)
	if err != nil {
		log.Fatal(err)
	}
	robots, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	tr := parseTrackerResponse(string(robots))
	tr.complete = tr.complete
}

func parsePeersDictionary(m []bItem) []torrentPeer {
	out := []torrentPeer{}
	for _, v := range m {
		p := new(torrentPeer)
		p.peer_id = v.d["peer id"].s
		p.ip = v.d["ip"].s
		p.port = int(v.d["port"].i)
		out = append(out, *p)
	}
	return out
}

func parsePeersString(s string) []torrentPeer {
	out := []torrentPeer{}
	for i := 0; i < len(s); i += 6 {
		p := new(torrentPeer)

		// get IP as x.x.x.x
		tmp := make([]uint8, 4)
		for j := 0; j < 4; j++ {
			tmp[j] = uint8(s[i+j])
		}
		p.ip = fmt.Sprintf("%d.%d.%d.%d", tmp[0], tmp[1], tmp[2], tmp[3])

		// get Port
		var n uint16
		tmp = make([]byte, 2)
		for j := 0; j < 2; j++ {
			tmp[j] = byte(s[i+j])
		}
		buf := bytes.NewBuffer(tmp)
		binary.Read(buf, binary.BigEndian, &n)
		p.port = int(n)

		out = append(out, *p)
	}
	return out
}

func parseTrackerResponse(s string) *TrackerResponse {
	tr := new(TrackerResponse)
	bi, _ := Bdecode([]byte(s))
	b := bi.d
	tr.complete = int(b["complete"].i)
	tr.incomplete = int(b["incomplete"].i)
	tr.tracker_id = b["tracker id"].s
	tr.interval = int(b["interval"].i)
	tr.min_interval = int(b["min interval"].i)
	tr.failure_reason = b["failure reason"].s

	if len(b["peers"].d) != 0 {
		tr.peers = parsePeersDictionary(b["peers"].l)
	} else {
		tr.peers = parsePeersString(b["peers"].s)
	}
	return tr
}
