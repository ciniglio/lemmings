package tracker

import (
	"fmt"
)

type piece struct {
	have      bool
	requested bool
	blocks    []bool
	data      []byte
}

type Pieces struct {
	pieces       []piece
	piece_length int
}

func (p *Pieces) Length() int {
	return len(p.pieces)
}

func (p *Pieces) HaveAtIndex(i int) bool {
	return p.pieces[i].have
}

func (p *Pieces) setAtIndex(i int, b bool) {
	p.pieces[i].have = b
}

func (p *Pieces) initBlocksAtPiece(i int) {
	size := p.piece_length / 16384
	p.pieces[i].blocks = make([]bool, size)
	p.pieces[i].data = make([]byte, p.piece_length)
}

func (p *Pieces) lengthBlocksInPiece(i int) int {
	return len(p.pieces[i].blocks)
}

func (ours *Pieces) GetPieceAndOffsetForRequest(theirs *Pieces) (int, int){
	indices := make([]int, 0)
	for i, p := range ours.pieces {
		if !p.have && p.requested && theirs.pieces[i].have {
			for j, b := range p.blocks {
				if !b {
					return i, j
				}
			}
		}
		if !p.have && theirs.pieces[i].have {
			indices = append(indices, i)
		}
	}
	if len(indices) <= 0 {
		return -1, -1
	}
	ind := indices[RandomInt(len(indices))]
	if ours.pieces[ind].blocks == nil {
		ours.initBlocksAtPiece(ind)
	}
	for i, b := range ours.pieces[ind].blocks {
		if !b {
			ours.pieces[ind].requested = true
			return ind, i
		}
	}
	return -1, -1
}

func (p *Pieces) HaveBlockAtPieceAndOffset(i, offset int) bool {
	if i >= p.Length() || offset >= p.lengthBlocksInPiece(i) {
		return false
	}
	return p.pieces[i].blocks[offset]
}

func (p *Pieces) SetBlockAtPieceAndOffest(i int, offset int, b []byte) {
	if i >= p.Length() || offset >= p.lengthBlocksInPiece(i) {
		return
	}
	if len(b) < 16384 && i < (p.Length()-1) {
		fmt.Printf("Got a bad block")
		return
	}
	p.pieces[i].blocks[offset] = true
	for j, by := range b {
		p.pieces[i].data[offset+j] = by
	}
	p.checkPiece(i)
}

func (p *Pieces) checkPiece(i int) {
	for _, b := range p.pieces[i].blocks {
		if !b { return }
	}
	p.pieces[i].have = true
}

func CreateNewPieces(num_pieces, piece_length int) *Pieces {
	pieces := new(Pieces)
	pieces.piece_length = int(piece_length)
	pieces.pieces = make([]piece, num_pieces)
	return pieces
}
