package tracker

import (
	"net"
)

type kind int

const (
	choke kind = iota
	unchoke
	interested
	not_interested
	have
	bitfield
	request
	piece_t
	cancel
	port

	i_get_request
	i_sent_request
	i_get_peers
	i_recv_block
	i_write_block
	i_subscribe
	i_request
	i_have
	i_cancel
	i_add_peer
	i_can_unchoke
	i_will_choke
)

type Message interface {
	kind() kind
}

type ChokeMessage struct{}
type UnchokeMessage struct{}
type InterestedMessage struct{}
type NotInterestedMessage struct{}

func (c ChokeMessage) kind() kind { return choke }

func (c ChokeMessage) bytes() []byte {
	b := make([]byte, 0)
	b = append(b, to4Bytes(uint32(1))...)
	b = append(b, byte(choke))
	return b
}

func (c UnchokeMessage) kind() kind { return unchoke }

func (c UnchokeMessage) bytes() []byte {
	b := make([]byte, 0)
	b = append(b, to4Bytes(uint32(1))...)
	b = append(b, byte(unchoke))
	return b
}

func (c InterestedMessage) kind() kind { return interested }

func (c InterestedMessage) bytes() []byte {
	b := make([]byte, 0)
	b = append(b, to4Bytes(uint32(1))...)
	b = append(b, byte(interested))
	return b
}

func (c NotInterestedMessage) kind() kind { return not_interested }

func (c NotInterestedMessage) bytes() []byte {
	b := make([]byte, 0)
	b = append(b, to4Bytes(uint32(1))...)
	b = append(b, byte(not_interested))
	return b
}

type HaveMessage struct {
	index int
}

func (c HaveMessage) kind() kind { return have }

func (m HaveMessage) bytes() (b []byte) {
	b = make([]byte, 0)
	b = append(b, to4Bytes(uint32(5))...)
	b = append(b, byte(have))
	b = append(b, to4Bytes(uint32(m.index))...)
	return b
}

type BitFieldMessage struct {
	bitfield []byte
}

func (c BitFieldMessage) kind() kind { return bitfield }

func (c BitFieldMessage) bytes() []byte {
	return c.bitfield
}

type RequestMessage struct {
	index  int
	begin  int
	length int
}

func (c RequestMessage) kind() kind { return request }

func (m RequestMessage) bytes() []byte {
	b := make([]byte, 0)
	b = append(b, to4Bytes(uint32(13))...)
	b = append(b, byte(request))
	b = append(b, to4Bytes(uint32(m.index))...)
	b = append(b, to4Bytes(uint32(m.begin))...)
	b = append(b, to4Bytes(uint32(m.length))...)
	return b
}

type PieceMessage struct {
	index int
	begin int
	block []byte
}

func (c PieceMessage) kind() kind { return piece_t }

func (m PieceMessage) bytes() []byte {
	b := make([]byte, 0)
	b = append(b, to4Bytes(uint32(9+len(m.block)))...)
	b = append(b, byte(piece_t))
	b = append(b, to4Bytes(uint32(m.index))...)
	b = append(b, to4Bytes(uint32(m.begin))...)
	b = append(b, m.block...)
	return b
}

type CancelMessage struct {
	index  int
	begin  int
	length int
}

func (c CancelMessage) kind() kind { return cancel }

func (m CancelMessage) bytes() []byte {
	b := make([]byte, 0)
	b = append(b, to4Bytes(uint32(13))...)
	b = append(b, byte(cancel))
	b = append(b, to4Bytes(uint32(m.index))...)
	b = append(b, to4Bytes(uint32(m.begin))...)
	b = append(b, to4Bytes(uint32(m.length))...)
	return b
}

type InternalGetRequestMessage struct {
	pieces *Pieces
	ret    chan int
}

func (c InternalGetRequestMessage) kind() kind { return i_get_request }

type InternalSendingRequestMessage struct {
	index int
	begin int
}

func (c InternalSendingRequestMessage) kind() kind { return i_sent_request }

type InternalGetPeersMessage struct {
	ret chan torrentPeer
}

func (c InternalGetPeersMessage) kind() kind { return i_get_peers }

type InternalReceivedBlockMessage struct {
	index int
	begin int
}

func (c InternalReceivedBlockMessage) kind() kind { return i_recv_block }

type InternalWriteBlockMessage struct {
	bytes []byte
	index int
}

func (c InternalWriteBlockMessage) kind() kind { return i_write_block }

type InternalSubscribeMessage struct {
	c chan Message
}

func (c InternalSubscribeMessage) kind() kind { return i_subscribe }

type InternalRequestMessage struct {
	m   RequestMessage
	ret chan *PieceMessage
}

func (c InternalRequestMessage) kind() kind { return i_request }

type InternalHaveMessage struct {
	index int
}

func (c InternalHaveMessage) kind() kind { return i_have }

type InternalCancelMessage struct {
	index  int
	begin  int
	length int
}

func (c InternalCancelMessage) kind() kind { return i_cancel }

type InternalAddPeerMessage struct {
	c       *net.TCPConn
	peer_id string
}

func (c InternalAddPeerMessage) kind() kind { return i_add_peer }

type InternalCanUnchokeMessage struct {
	ret chan bool
}

func (c InternalCanUnchokeMessage) kind() kind { return i_can_unchoke }

type InternalChokingMessage struct{}

func (c InternalChokingMessage) kind() kind { return i_will_choke }
