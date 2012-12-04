package tracker

import (
	"bytes"
	"encoding/binary"
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

type Peer struct {
	torrent_peer torrentPeer
	connection   *net.TCPConn

	connected                 bool
	connection_info           *PeerConnectionInfo
	shook_hands               bool
	outstanding_request_count int

	their_pieces *Pieces

	messageChannel chan Message
	clientChannel  chan Message
	torrent        Torrent
}

func initialConnectionInfo() *PeerConnectionInfo {
	ci := new(PeerConnectionInfo)
	ci.peer_choking, ci.am_choking = true, true
	ci.peer_interested, ci.am_interested = false, false
	return ci
}

func (peer *Peer) connect() bool {
	debugl.Println("Calling connect(): ", peer)
	p := peer.torrent_peer
	dest_addr := new(net.TCPAddr)
	dest_addr.IP = net.ParseIP(p.ip)
	if dest_addr.IP == nil {
		errorl.Println("Couldn't get a valid IP")
		return false
	}
	dest_addr.Port = p.port

	var err error

	peer.connection, err = net.DialTCP("tcp", nil, dest_addr)
	if err != nil {
		errorl.Println("Couldn't connect: ", err)
		return false
	} else {
		debugl.Println("Connected to a peer: ", dest_addr.IP, dest_addr.Port)
	}
	peer.connection.SetKeepAlive(true)
	peer.connected = true
	return true
}

func CreatePeer(p torrentPeer, t *TorrentInfo, m chan Message, torrent Torrent) *Peer {
	peer := new(Peer)
	peer.torrent_peer = p
	debugl.Println("Peer in CreatePeer", p)

	peer.connection_info = initialConnectionInfo()
	peer.their_pieces = CreateNewPieces(t.numpieces, t)
	peer.messageChannel = make(chan Message, 10) // magic number
	peer.clientChannel = m
	peer.torrent = torrent
	return peer
}

func (peer *Peer) RunPeer() {
	if !peer.connected && !peer.connect() {
		errorl.Println("Connection problem")
		return
	}
	if !peer.shook_hands {
		peer.initiateHandshake()
		debugl.Println("Sent Handshake")
	}
	peer.clientChannel <- InternalSubscribeMessage{peer.messageChannel}
	go peer.readerRoutine()

	for {
		runtime.Gosched()
		select {
		case msg := <-peer.messageChannel:
			debugl.Println("Recvd Message: ", peer)
			switch msg.kind() {
			case choke:
				debugl.Println("Choked")
				peer.connection_info.peer_choking = true
			case unchoke:
				debugl.Println("Unchoked")
				peer.connection_info.peer_choking = false
			case interested:
				debugl.Println("Interested")
				peer.connection_info.peer_interested = true
			case not_interested:
				debugl.Println("Not Interested")
				peer.connection_info.peer_interested = false
			case bitfield:
				debugl.Println("BitField")
				peer.handleBitField(msg.(BitFieldMessage))
			case have:
				debugl.Println("Have")
				peer.handleHave(msg.(HaveMessage))
			case request:
				debugl.Println("Request")
				// talk to client, see if we have it, send piece
				peer.handleRequest(msg.(RequestMessage))
			case piece_t:
				debugl.Println("Piece")
				// talk to client
				peer.handlePiece(msg.(PieceMessage))
			case i_cancel:
				debugl.Println("Other peer recieved block")
				peer.sendCancel(msg.(InternalCancelMessage))
			case i_have:
				debugl.Println("Recieved Broadcast")
				peer.send(HaveMessage(msg.(InternalHaveMessage)).bytes())
			default:
				debugl.Println("Something weird")
			}
		default:
			peer.act()
		}
	}
}

func (p *Peer) send(b []byte) {
	n := 0
	var err error
	debugl.Println("trying to send:", b)
	for n < len(b) {
		b = b[n:]
		n, err = p.connection.Write(b)
		if err != nil {
			errorl.Println("Send error", err)
			return
		}
	}
}

func (p *Peer) getIndexAndBeginForRequest() (int, int) {
	m := new(InternalGetRequestMessage)
	m.pieces = p.their_pieces
	m.ret = make(chan int, 2)
	p.clientChannel <- m
	index := <-m.ret
	begin := <-m.ret
	return index, begin
}

func (p *Peer) sendRequest(index, begin int) {
	if p.outstanding_request_count < 2 {
		m := RequestMessage{
			index,
			begin,
			p.their_pieces.blockSize(index, begin/int(block_size)),
		}
		p.outstanding_request_count += 1
		p.send(m.bytes())
		n := InternalSendingRequestMessage{index, begin}
		debugl.Println("Adding sent request to clientchan", len(p.clientChannel))
		p.clientChannel <- &n
	}
}

func (p *Peer) sendCancel(m InternalCancelMessage) {
	if p.outstanding_request_count > 0 {
		if p.their_pieces.requested(m.index, m.begin) {
			p.outstanding_request_count -= 1
			p.send(CancelMessage(m).bytes())
		}
	}
}

func (p *Peer) act() {
	switch {
	case p.connection_info.am_interested && p.connection_info.peer_choking:
		p.send(InterestedMessage{}.bytes())
	case p.connection_info.peer_interested && p.connection_info.am_choking:
		if p.torrent.CanUnchoke() {
			p.connection_info.am_choking = false
			p.send(UnchokeMessage{}.bytes())
		}
	case !p.connection_info.peer_interested && !p.connection_info.am_choking:
		p.torrent.WillChoke()
		p.connection_info.am_choking = true
		p.send(ChokeMessage{}.bytes())
	default:
		n, b := p.getIndexAndBeginForRequest()
		if n >= 0 && b >= 0 {
			switch {
			case !p.connection_info.am_interested:
				p.connection_info.am_interested = true
				debugl.Println("Sending Interested")
				p.send(InterestedMessage{}.bytes())
				p.sendRequest(n, b)
			case !p.connection_info.peer_choking:
				p.sendRequest(n, b)
			}
		} else {
			switch {
			case p.connection_info.am_interested:
				p.send(NotInterestedMessage{}.bytes())
				p.connection_info.am_interested = false
			}
		}
	}
}

func (p *Peer) handlePiece(m PieceMessage) {
	p.outstanding_request_count -= 1
	debugl.Println("Adding piece to clientchan", len(p.clientChannel))
	p.clientChannel <- m
}

func (p *Peer) handleRequest(m RequestMessage) {
	msg := InternalRequestMessage{m, make(chan *PieceMessage)}
	p.clientChannel <- msg
	ret := <-msg.ret
	if ret != nil {
		p.send(ret.bytes())
	}
}

func (peer *Peer) handleBitField(m BitFieldMessage) {
	peer.their_pieces.AddBitField(m.bitfield)
}

func (peer *Peer) handleHave(m HaveMessage) {
	peer.their_pieces.AddHave(m.index)
}

func (p *Peer) readerRoutine() {
	var buffer []byte
	for {
		bufsize := 1024
		data := make([]byte, bufsize)
		n, err := p.connection.Read(data)

		if err != io.EOF && err != nil {
			errorl.Printf("Read %d bytes\n", n)
			errorl.Println("Reading from connection: ", err, time.Now())
			return
		}

		if n > 0 {
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
			debugl.Println("Adding to message queue; unread: ", len(p.messageChannel))
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

	if size > len(buffer) {
		return nil, 0
	}
	if size > 4 {
		id = toInt(buffer[4:5])
	}

	curpos := size
	switch {
	case size == 4:
		//do nothing
	default:
		debugl.Println("Parsing Message, ID: ", id)
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

type handshakeMessage struct {
	pstrlen   byte
	pstr      []byte
	reserved  [8]byte
	info_hash [20]byte
	peer_id   [20]byte
}

func (p *Peer) parseHandshakeMessage(b *[]byte) (*handshakeMessage, int) {
	if p.shook_hands {
		return nil, 0
	}
	debugl.Println("Parsing handshake")
	m := *b
	curpos := 0
	message := new(handshakeMessage)
	message.pstrlen = m[0]
	if message.pstrlen != byte(len("BitTorrent protocol")) {
		return nil, 0
	}
	curpos += 1
	message.pstr = m[curpos : int(m[0])+curpos]
	debugl.Printf("Message length: %d\n", int(m[0]))
	for i, b := range message.pstr {
		if b != byte("BitTorrent protocol"[i]) {
			return nil, 0
		}
	}
	curpos += int(m[0])
	curpos += 8 // Reserved bytes
	for i, _ := range m[curpos : 20+curpos] {
		message.info_hash[i] = m[curpos : 20+curpos][i]
		if message.info_hash[i] != byte(p.torrent.InfoHash()[i]) {
			return nil, 0
		}
	}
	curpos += 20 // info_hash
	for i, _ := range m[curpos : 20+curpos] {
		message.peer_id[i] = m[curpos : 20+curpos][i]
	}
	curpos += 20 // peer_id
	debugl.Printf("Message: %q\n", message.info_hash)
	debugl.Printf("Peer: %q\n", message.peer_id)

	p.shook_hands = true
	*b = m[curpos:]
	return message, curpos
}

func (p *Peer) initiateHandshake() {
	message := new(handshakeMessage)
	message.pstrlen = byte(len("BitTorrent protocol"))
	message.pstr = make([]byte, len("BitTorrent protocol"))
	binary.Read(strings.NewReader("BitTorrent protocol"),
		binary.BigEndian, &message.pstr)

	binary.Read(strings.NewReader(p.torrent.InfoHash()),
		binary.BigEndian, &message.info_hash)
	binary.Read(strings.NewReader(p.torrent.ClientId()),
		binary.BigEndian, &message.peer_id)

	p.connection.Write(message.bytes())
	debugl.Printf("msg %x", message.bytes())

	return
}

func (h *handshakeMessage) bytes() []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(h.pstrlen)
	buf.Write(h.pstr)
	binary.Write(buf, binary.BigEndian, h.reserved)
	binary.Write(buf, binary.BigEndian, h.info_hash)
	binary.Write(buf, binary.BigEndian, h.peer_id)
	return buf.Bytes()
}
