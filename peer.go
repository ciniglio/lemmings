package tracker

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"strings"
	"io"
)

type PeerConnectionInfo struct {
	peer_choking    bool
	peer_interested bool
	am_choking      bool
	am_interested   bool
}

type HandshakeMessage struct {
	pstrlen   byte
	pstr      []byte
	reserved  [8]byte
	info_hash [20]byte
	peer_id   [20]byte
}

type Peer struct {
	torrent_info *TorrentInfo
	torrent_peer *torrentPeer
	connection   *net.TCPConn
	connection_info *PeerConnectionInfo
	their_id     [20]byte
	shook_hands  bool
}

func CreatePeer(p *torrentPeer, t *TorrentInfo) *Peer {
	peer := new(Peer)
	peer.torrent_info = t
	peer.torrent_peer = p

	dest_addr := new(net.TCPAddr)
	dest_addr.IP = net.ParseIP(p.ip)
	if dest_addr.IP == nil {
		fmt.Println("Couldn't get a valid IP")
		return nil
	}
	dest_addr.Port = p.port

	var err error

	peer.connection, err = net.DialTCP("tcp", nil, dest_addr)
	if err != nil {
		fmt.Println("Couldn't connect: ", err)
		return nil
	} else {
		fmt.Println("Connected to a peer")
	}
	peer.runPeer()
	return peer
}

func (peer *Peer) runPeer() {
	peer.initiateHandshake()
	fmt.Println("Sent Handshake")
	var data []byte
	for {
		n := peer.readRawBytesFromConnection(&data)
		fmt.Printf("OUTER Data has %d bytes: % X\n", len(data), data)
		if len(data) > 0 {
			fmt.Printf("Receieved %d bytes: % X\n", n, data)
			if msg, _ := peer.parseHandshakeMessage(&data); msg == nil {
				fmt.Printf("Data has %d bytes: % X\n", len(data), data)
				peer.parseProtocolMessage(&data)
				fmt.Printf("Data has %d bytes: % X\n", len(data), data)
			}
		}
	}
}

func (peer *Peer) parseProtocolMessage(b *[]byte) {
	m := *b
	curpos := 0
	switch {
	case bytes.Compare(m[:2], []byte("\x00\x00")) == 0 :
		// Keep alive, do nothing
		curpos += 2
		fmt.Printf("Keep Alive\n")
	case bytes.Compare(m[:2], []byte("\x00\x01")) == 0 :
		curpos += 2 + toInt(m[:2])
		fmt.Printf("Choke/Unchoke/Interested/Uninterested")
		peer.recieveChokeAndInterest(m[2:curpos])
	}
	*b = m[curpos:]
	return
}

func (peer *Peer) recieveChokeAndInterest(b []byte) {
	switch {
	case bytes.Compare(b, []byte("\x00")) == 0:
		peer.connection_info.peer_choking = true
	case bytes.Compare(b, []byte("\x01")) == 0:
		peer.connection_info.peer_choking = false
	case bytes.Compare(b, []byte("\x02")) == 0:
		peer.connection_info.peer_interested = true
	case bytes.Compare(b, []byte("\x03")) == 0:
		peer.connection_info.peer_interested = false
		
	}
}

func (p *Peer) readRawBytesFromConnection(out *[]byte) (int) {
	readcount := 0
	bufsize := 1024
	//out := make([]byte, 0)
	for {
		data := make([]byte, bufsize)
		n, err := p.connection.Read(data) 
		if err != io.EOF && err != nil {
			fmt.Printf("Read %d bytes\n", n)
			fmt.Println("Reading from connection: ", err)
		}
		readcount += n
		*out = append(*out, data[0:n]...)
		if n < bufsize {
			break
		}
	}
	return readcount
}

func (p *Peer) parseHandshakeMessage(b *[]byte) (*HandshakeMessage, int) {
	if p.shook_hands {
		return nil, 0
	}
	m := *b
	curpos := 0
	message := new(HandshakeMessage)
	message.pstrlen = m[0]
	if message.pstrlen != byte(len("BitTorrent protocol")) {
		return nil, 0
	}
	curpos += 1
	message.pstr = m[curpos:int(m[0])+curpos]
	fmt.Printf("Message length: %d\n", int(m[0]))
	for i, b := range message.pstr{
		if b != byte("BitTorrent protocol"[i]) {
			return nil, 0
		}
	}
	curpos += int(m[0])
	curpos += 8  // Reserved bytes
	for i, _ := range m[curpos:20+curpos]{
		message.info_hash[i] = m[curpos:20+curpos][i]
		if message.info_hash[i] != byte(p.torrent_info.info_hash[i]) {
			return nil, 0
		}
	}
	curpos += 20 // info_hash
	for i, _ := range m[curpos:20+curpos]{
		message.peer_id[i] = m[curpos:20+curpos][i]
	}
	curpos += 20 // peer_id
	fmt.Printf("Message: %q\n", message.info_hash)
	fmt.Printf("Peer: %q\n", message.peer_id)
	
	p.shook_hands = true
	*b = m[curpos:]
	return message, curpos
}

func (p *Peer) initiateHandshake() {
	message := new(HandshakeMessage)
	message.pstrlen = byte(len("BitTorrent protocol"))
	message.pstr = make([]byte, len("BitTorrent protocol"))
	binary.Read(strings.NewReader("BitTorrent protocol"),
		binary.BigEndian, &message.pstr)

	binary.Read(strings.NewReader(p.torrent_info.info_hash),
		binary.BigEndian, &message.info_hash)
	binary.Read(strings.NewReader(p.torrent_info.client_id),
		binary.BigEndian, &message.peer_id)

	p.connection.Write(message.bytes())

	return
}

func (h *HandshakeMessage) bytes() []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(h.pstrlen)
	buf.Write(h.pstr)
	binary.Write(buf, binary.BigEndian, h.reserved)
	binary.Write(buf, binary.BigEndian, h.info_hash)
	binary.Write(buf, binary.BigEndian, h.peer_id)
	return buf.Bytes()
}
