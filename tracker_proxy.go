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

type trackerGetRequest struct {
	announce_url string
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

}

type torrentPeer struct {
	peer_id string //optional initially
	ip      string
	port    int
}

type trackerResponse struct {
	failure_reason string
	interval       int
	min_interval   int
	tracker_id     string
	complete       int
	incomplete     int
	peers          []torrentPeer
}

type TrackerProxy struct {
	tgr      *trackerGetRequest
	response *trackerResponse
	msg      chan Message
}

func NewTrackerProxy(t *TorrentInfo) *TrackerProxy {
	tp := new(TrackerProxy)
	tgr := new(trackerGetRequest)
	tgr.info_hash = t.info_hash
	tgr.peer_id = t.client_id
	tgr.announce_url = t.announce
	tp.tgr = tgr
	tp.msg = make(chan Message)
	go tp.handleMessages()
	return tp
}

func (t *TrackerProxy) handleMessages() {
	for m := range t.msg {
		switch m.kind() {
		case i_get_peers:
			msg := m.(InternalGetPeersMessage)
			for _, p := range t.GetPeers() {
				msg.ret <- p
			}
			close(msg.ret)
		}
	}
}

func (t *trackerGetRequest) generateGetString() string {
	u, _ := url.Parse(t.announce_url)
	v := url.Values{}
	v.Set("info_hash", t.info_hash)
	v.Set("peer_id", t.peer_id)
	u.RawQuery = v.Encode()
	return u.String()
}

func (t *trackerGetRequest) makeTrackerRequest() *trackerResponse {
	q := t.generateGetString()
	fmt.Printf("Tracker Announce: %v\n\n", q)
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
	fmt.Printf("Tracker Response: %s\n\n", string(robots))
	return tr
}

func (t *TrackerProxy) GetPeers() []torrentPeer {
	t.response = t.tgr.makeTrackerRequest()
	return t.response.peers
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
		for j := 4; j < 6; j++ {
			tmp[j-4] = byte(s[i+j])

		}
		buf := bytes.NewBuffer(tmp)
		binary.Read(buf, binary.BigEndian, &n)
		p.port = int(n)

		out = append(out, *p)
	}
	return out
}

func parseTrackerResponse(s string) *trackerResponse {
	tr := new(trackerResponse)
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
