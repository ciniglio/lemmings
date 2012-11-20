package tracker

import (
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
)

type Message interface {
	kind() kind
}

type ChokeMessage struct {}
type UnchokeMessage struct {}
type InterestedMessage struct {}
type NotInterestedMessage struct {}

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

func (m HaveMessage) bytes() []byte {
	b := make([]byte, 0)
	b = append(b, to4Bytes(uint32(5))...)
	b = append(b, byte(have))
	b = append(b, to4Bytes(uint32(m.index))...)
	return b
}

type BitFieldMessage struct {
	bitfield []byte
}

func (c BitFieldMessage) kind() kind { return bitfield }

type RequestMessage struct {
	index int
	begin int
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
	b = append(b, to4Bytes(uint32(9 + len(m.block)))...)
	b = append(b, byte(piece_t))
	b = append(b, to4Bytes(uint32(m.index))...)
	b = append(b, to4Bytes(uint32(m.begin))...)
	b = append(b, m.block...)
	return b
}

type CancelMessage struct {}

func (c CancelMessage) kind() kind { return cancel }

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