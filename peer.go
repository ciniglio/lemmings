package tracker

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"runtime"
	"strings"
	"time"
)

var block_size int64 = 16384

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
	torrent_info              *TorrentInfo
	torrent_peer              torrentPeer
	connection                *net.TCPConn
	connection_info           *PeerConnectionInfo
	their_id                  [20]byte
	shook_hands               bool
	their_pieces              *Pieces
	outstanding_request_count int
	messageChannel            chan Message
	clientChannel             chan Message
}

func InitialConnectionInfo() *PeerConnectionInfo {
	ci := new(PeerConnectionInfo)
	ci.peer_choking, ci.am_choking = true, true
	ci.peer_interested, ci.am_interested = false, false
	return ci
}

func (peer *Peer) connect() bool {
	fmt.Println("Calling connect(): ", peer)
	p := peer.torrent_peer
	dest_addr := new(net.TCPAddr)
	dest_addr.IP = net.ParseIP(p.ip)
	if dest_addr.IP == nil {
		fmt.Println("Couldn't get a valid IP")
		return false
	}
	dest_addr.Port = p.port

	var err error

	peer.connection, err = net.DialTCP("tcp", nil, dest_addr)
	if err != nil {
		fmt.Println("Couldn't connect: ", err)
		return false
	} else {
		fmt.Println("Connected to a peer: ", dest_addr.IP, dest_addr.Port)
	}
	peer.connection.SetKeepAlive(true)

	return true
}

func CreatePeer(p torrentPeer, t *TorrentInfo, m chan Message) *Peer {
	peer := new(Peer)
	peer.torrent_info = t
	peer.torrent_peer = p
	fmt.Println("Peer in CreatePeer", p)

	if !peer.connect() {
		fmt.Println("Connection problem")
		return nil
	}
	peer.connection_info = InitialConnectionInfo()
	peer.their_pieces = CreateNewPieces(t.numpieces, int(t.pieceLength))
	peer.messageChannel = make(chan Message)
	peer.clientChannel = m
	peer.runPeer()
	return peer
}

func (peer *Peer) runPeer() {
	peer.initiateHandshake()
	go peer.readerRoutine()
	fmt.Println("Sent Handshake")
	for {
		runtime.Gosched()
		select {
		case msg := <-peer.messageChannel:
			switch msg.kind() {
			case choke:
				fmt.Println("Choked")
				peer.connection_info.peer_choking = true
			case unchoke:
				fmt.Println("Unchoked")
				peer.connection_info.peer_choking = false
			case interested:
				fmt.Println("Interested")
				peer.connection_info.peer_interested = true
			case not_interested:
				fmt.Println("Not Interested")
				peer.connection_info.peer_interested = false
			case bitfield:
				fmt.Println("BitField")
				peer.handleBitField(msg.(BitFieldMessage))
			case have:
				fmt.Println("Have")
				peer.handleHave(msg.(HaveMessage))
			case request:
				fmt.Println("Request")
				// talk to client, see if we have it, send piece
				peer.handleRequest(msg.(RequestMessage))
			case piece_t:
				fmt.Println("Piece")
				// talk to client
				peer.handlePiece(msg.(PieceMessage))
				//case client_have: 
				// client tells me we just recvd piece
				// I send haves and cancels
			}
		default:
			peer.act()
		}
	}
}

func (p *Peer) Send(b []byte) {
	n := 0
	var err error
	for n < len(b) {
		b = b[n:]
		n, err = p.connection.Write(b)
		if err != nil {
			fmt.Println("Send error", err)
		}
	}
}

func (p *Peer) GetIndexAndBeginForRequest() (int, int) {
	m := new(InternalGetRequestMessage)
	m.pieces = p.their_pieces
	m.ret = make(chan int, 2)
	p.clientChannel <- m
	index := <-m.ret
	begin := <-m.ret
	return index, begin
}

func (p *Peer) SendRequest(index, begin int) {
	if p.outstanding_request_count < 2 {
		m := RequestMessage{}
		m.index = index
		m.begin = begin
		length := block_size

		if index == p.their_pieces.Length()-1 {
			rem := p.torrent_info.total_length % p.torrent_info.pieceLength
			last_ind := int(rem / block_size)
			if begin == last_ind {
				length = p.torrent_info.total_length % block_size
				fmt.Println("Last block, length: ", length)
			}
		}
		m.length = int(length)
		p.outstanding_request_count += 1
		p.Send(m.bytes())
		fmt.Println("Sending request")
		n := new(InternalSendingRequestMessage)
		n.index = index
		n.begin = begin
		p.clientChannel <- n
	}
}

func (p *Peer) act() {
	switch {
	case p.connection_info.am_interested && p.connection_info.peer_choking:
		//p.Send(UnchokeMessage{}.bytes())
	default:
		n, b := p.GetIndexAndBeginForRequest()
		if n >= 0 && b >= 0 {
			switch {
			case !p.connection_info.am_interested:
				p.connection_info.am_interested = true
				fmt.Println("Sending Interested")
				p.Send(InterestedMessage{}.bytes())
			case !p.connection_info.peer_choking:
				p.SendRequest(n, b)
			}
		} else {
			switch {
			case p.connection_info.am_interested:
				p.Send(NotInterestedMessage{}.bytes())
			}
		}
	}
}

func (p *Peer) handlePiece(m PieceMessage) {
	p.outstanding_request_count -= 1
	p.clientChannel <- m
}

func (peer *Peer) handleRequest(m RequestMessage) {
}

func (peer *Peer) handleBitField(m BitFieldMessage) {
	b := m.bitfield
	ind := 0
	for _, by := range b {
		for j := 7; j >= 0; j-- {
			have := ((by>>uint(j))&1 == 1)
			peer.their_pieces.setAtIndex(ind, have)
			ind++
			if ind >= peer.torrent_info.numpieces {
				return
			}
		}
	}
}

func (peer *Peer) handleHave(m HaveMessage) {
	peer.their_pieces.setAtIndex(m.index, true)
}

func (p *Peer) readerRoutine() {
	var buffer []byte
	for {
		//fmt.Println("Reader Routine", time.Now())
		bufsize := 1024
		data := make([]byte, bufsize)
		n, err := p.connection.Read(data)

		if err != io.EOF && err != nil {
			fmt.Printf("Read %d bytes\n", n)
			fmt.Println("Reading from connection: ", err, time.Now())
			return
		}

		if n > 0 {
			fmt.Printf("Read %d bytes @ %v\n", n, time.Now())
			buffer = append(buffer, data[0:n]...)
		}
		if len(buffer) < 4 {
			continue
		}

		p.parseHandshakeMessage(&buffer)

		for msg, curpos := parseBytesToMessage(buffer); msg != nil; {
			buffer = buffer[curpos:]
			if msg != nil {
				p.messageChannel <- msg
			}
			msg, curpos = parseBytesToMessage(buffer)
		}
	}
}

func parseBytesToMessage(buffer []byte) (Message, int) {
	if len(buffer) < 4 {
		return nil, 0
	}

	var msg Message
	size := 4 + toInt(buffer[:4])
	id := 0
	if size > 0 {
		id = toInt(buffer[4:5])
	}
	if size > len(buffer) {
		return nil, 0
	}
	curpos := size
	switch {
	case size == 0:
		//do nothing
	default:
		fmt.Println("Parsing Message, ID: ", id)
		switch id {
		case 0:
			msg = ChokeMessage{}
		case 1:
			msg = UnchokeMessage{}
		case 2:
			msg = InterestedMessage{}
		case 3:
			msg = NotInterestedMessage{}
		case 4:
			msg = HaveMessage{index: toInt(buffer[5:curpos])}
		case 5:
			msg = BitFieldMessage{bitfield: buffer[5:curpos]}
		case 6:
			index := toInt(buffer[5:9])
			begin := toInt(buffer[9:13])
			length := toInt(buffer[13:17])
			msg = RequestMessage{index: index, begin: begin, length: length}
		case 7:
			index := toInt(buffer[5:9])
			begin := toInt(buffer[9:13])
			block := buffer[13:curpos]
			msg = PieceMessage{index: index, begin: begin, block: block}
		default:
			msg = CancelMessage{}
		}
	}

	return msg, curpos
}

func (p *Peer) parseHandshakeMessage(b *[]byte) (*HandshakeMessage, int) {
	if p.shook_hands {
		return nil, 0
	}
	fmt.Println("Parsing handshake")
	m := *b
	curpos := 0
	message := new(HandshakeMessage)
	message.pstrlen = m[0]
	if message.pstrlen != byte(len("BitTorrent protocol")) {
		return nil, 0
	}
	curpos += 1
	message.pstr = m[curpos : int(m[0])+curpos]
	fmt.Printf("Message length: %d\n", int(m[0]))
	for i, b := range message.pstr {
		if b != byte("BitTorrent protocol"[i]) {
			return nil, 0
		}
	}
	curpos += int(m[0])
	curpos += 8 // Reserved bytes
	for i, _ := range m[curpos : 20+curpos] {
		message.info_hash[i] = m[curpos : 20+curpos][i]
		if message.info_hash[i] != byte(p.torrent_info.info_hash[i]) {
			return nil, 0
		}
	}
	curpos += 20 // info_hash
	for i, _ := range m[curpos : 20+curpos] {
		message.peer_id[i] = m[curpos : 20+curpos][i]
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
	fmt.Printf("msg %x", message.bytes())

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
