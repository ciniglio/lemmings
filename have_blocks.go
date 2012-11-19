package tracker

import (
//	"fmt"
)

type Blocks struct {
	blocks []bool
}

type Pieces struct {
	have   bool
	blocks *[]Blocks
}

func CreateNewPieces(num_pieces int) *[]Pieces {
	p := make([]Pieces, num_pieces)
	return &p
}
