package tracker

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)


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
	torrent   Torrent
	announce  string
	event     string
	msg       chan Message
	timeout   <- chan time.Time
}

func NewTrackerProxy(ti *TorrentInfo, t Torrent) *TrackerProxy {
	tp := new(TrackerProxy)
	tp.torrent = t
	tp.announce = ti.announce
	tp.msg = make(chan Message)
	go tp.handleMessages()
	return tp
}

func (t *TrackerProxy) handleMessages() {
	for {
		select {
		case m := <- t.msg:
			switch m.kind() {
			case i_get_peers:
				msg := m.(InternalGetPeersMessage)
				for _, p := range t.getPeers() {
					msg.ret <- p
				}
				close(msg.ret)
			case i_finished_torrent:
				msg := m.(InternalFinishedTorrentMessage)
				t.sendFinished(msg.upload, msg.download)
			}
		case <- t.timeout:
			t.sendAnnounce()
		}
	}
}

func (t *TrackerProxy) newTrackerGetRequest() *trackerGetRequest {
	tgr := new(trackerGetRequest)
	tgr.announce_url = t.announce
	tgr.info_hash = t.torrent.InfoHash()
	tgr.peer_id = t.torrent.ClientId()
	tgr.port = strconv.Itoa(PORT)
	return tgr
}

func (t *TrackerProxy) getPeers() []torrentPeer {
	tgr := t.newTrackerGetRequest()
	tgr.event = "started"
	response := tgr.makeTrackerRequest()
	t.timeout = time.After(time.Duration(response.interval) * time.Second)
	return response.peers
}

func (t *TrackerProxy) sendFinished(u, d int) {
	debugl.Println("Sending completed to tracker", u, d)
	tgr := t.newTrackerGetRequest()
	tgr.event = "completed"
	tgr.uploaded = u
	tgr.downloaded = d
	response := tgr.makeTrackerRequest()
	t.timeout = time.After(time.Duration(response.interval) * time.Second)
}

func (t *TrackerProxy) sendAnnounce() {
	tgr := t.newTrackerGetRequest()
	tgr.uploaded, tgr.downloaded = t.torrent.Stats()
	response := tgr.makeTrackerRequest()
	t.timeout = time.After(time.Duration(response.interval) * time.Second)
}

func (t *TrackerProxy) Done(u, d int) {
	m := InternalFinishedTorrentMessage{u, d}
	t.msg <- m
}

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

func (t *trackerGetRequest) generateGetString() string {
	u, _ := url.Parse(t.announce_url)
	v := url.Values{}
	v.Set("info_hash", t.info_hash)
	v.Set("peer_id", t.peer_id)
	v.Set("event", t.event)
	if t.uploaded != 0 || t.downloaded != 0 {
		v.Set("uploaded", strconv.Itoa(t.uploaded))
		v.Set("downloaded", strconv.Itoa(t.downloaded))
	}
	v.Set("port", t.port)
	u.RawQuery = v.Encode()
	return u.String()
}

func (t *trackerGetRequest) makeTrackerRequest() *trackerResponse {
	q := t.generateGetString()
	t.event = ""
	debugl.Printf("Tracker Announce: %v\n\n", q)
	res, err := http.Get(q)
	if err != nil {
		log.Fatal(err)
	}
	b, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	tr := parseTrackerResponse(string(b))
	debugl.Printf("Tracker Response: %s\n\n", string(b))
	return tr
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
