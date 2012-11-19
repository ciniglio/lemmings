package tracker

import (
	"testing"
)

func Test_CreatingStructure_1(test *testing.T) {
	c := CreateNewPieces(16, 16384)
	if c.Length() != 16 {
		test.Error("Didn't init the right number of pieces")
	}
	for i := 0; i < 16; i++ {
		if c.HaveAtIndex(i) != false {
			test.Error("Failed init with false", c)
		}
	}
	c.initBlocksAtPiece(0)
	if c.HaveBlockAtPieceAndOffset(0, 0) != false {
		test.Error("Failed block init", c)
	}
	c.SetBlockAtPieceAndOffest(0, 0, make([]byte, 16384))
	if c.HaveBlockAtPieceAndOffset(0, 0) != true {
		test.Error("Failed block set", c)
	}
}
