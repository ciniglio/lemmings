package tracker

import (
	"testing"
)

func Test_CreatingStructure_1(test *testing.T) {
	c := CreateNewPieces(16)
	if len(*c) != 16 {
		test.Error("Didn't init the right number of pieces")
	}
	for i := 0; i < 16; i++ {
		if (*c)[i].have != false {
			test.Error("Failed init with false", (*c)[i])
		}
		if (*c)[i].blocks != nil {
			test.Error("Failed init with nil", (*c)[i])
		}
	}
}
