package tracker

import (
	"testing"
)

func Test_CreatingStructure_1(test *testing.T) {
	c := CreateNewPieces(16)
	if c.length() != 16 {
		test.Error("Didn't init the right number of pieces")
	}
	for i := 0; i < 16; i++ {
		if c.haveAtIndex(i) != false {
			test.Error("Failed init with false", c)
		}
	}
}
