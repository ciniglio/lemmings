package tracker

import (
	"bytes"
	"testing"
)

func Test_HandshakeMessage_1(t *testing.T) {
	main()
}

func Test_BitField_1(t *testing.T) {
	pieces := new(Pieces)
	pieces.pieces = make([]piece, 8)
	pieces.pieces[7].have = true
	if pieces.CreateBitField()[0] != byte(1) {
		t.Error("Failed bitfield")
	}

	pieces = new(Pieces)
	pieces.pieces = make([]piece, 11)
	pieces.pieces[7].have = true
	pieces.pieces[9].have = true
	pieces.pieces[10].have = true
	if !bytes.Equal(pieces.CreateBitField(), []byte{byte(1), byte(96)}) {
		t.Error("Failed bitfield", pieces.CreateBitField())
	}
}
