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

func CreateNewPieces(num_pieces int64) []Pieces {
	return make([]Pieces, num_pieces)
}
