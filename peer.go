package tracker

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"strings"
)

type PeerConnection struct {
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
	peer.initiateHandshake()
	return peer
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

	//p.connection
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
