package tracker

import (
	"net"
)

const PORT int = 56565
const handshake_length int = 68

type Client struct {
	torrents map[string](*Torrent)
	messages chan Message
}

func NewClient() Client {
	c := Client{make(map[string](*Torrent)), make(chan Message)}
	return c
}

func (self Client) Run() {
	addr := &net.TCPAddr{nil, PORT}
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		errorl.Println("Listening error:", err)
	}
	new_peers := make(chan *net.TCPConn)

	go listenOnPort(listener, new_peers)

	done := make(chan int)
	for {
		select {
		case m := <-self.messages:
			msg := m.(InternalAddTorrentMessage)

			s, t := LaunchTorrent(msg.filename, done)
			self.torrents[s] = &t
		case c := <- new_peers:
			ih, peer_id := getInfoHashFromPeer(c)
			if ih == "" {
				continue
			}
			t := self.torrents[ih]
			if t == nil {
				continue
			}
			t.messages <- InternalAddPeerMessage{c, peer_id}
		}
	}
}

func listenOnPort(listener *net.TCPListener, c chan *net.TCPConn) {
	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			errorl.Println("Accept Error:", err)
			continue
		}
		c <- conn
	}
}

func (self Client) AddTorrent(name string) {
	self.messages <- InternalAddTorrentMessage{name}
}

func getInfoHashFromPeer(c net.Conn) (string, string) {
	b := make([]byte, handshake_length)
	debugl.Println("Parsing handshake from client")
	i := 0
	for i < handshake_length {
		t := make([]byte, handshake_length)
		n, err := c.Read(t)
		if err != nil {
			errorl.Println("Error recieving", err)
			return "", ""
		}
		b = append(b, t...)
		i += n
	}

	info_hash := string(b[28:48])
	peer_id := string(b[48:68])

	return info_hash, peer_id
}
