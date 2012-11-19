package tracker

import (
//	"fmt"
)

type blocks struct {
	blocks []bool
}

type piece struct {
	have   bool
	blocks *[]blocks
}

type Pieces struct {
	pieces []piece
}

func (p *Pieces) length() int {
	return len(p.pieces)
}

func (p *Pieces) haveAtIndex(i int) bool {
	return p.pieces[i].have
}

func (p *Pieces) setAtIndex(i int, b bool) {
	p.pieces[i].have = b
}

func CreateNewPieces(num_pieces int) *Pieces {
	pieces := new(Pieces)
	pieces.pieces = make([]piece, num_pieces)
	return pieces
}
