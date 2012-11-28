package tracker

import (
	"crypto/sha1"
	"fmt"
)

type piece struct {
	have             bool
	requested        bool
	blocks           []bool
	blocks_requested []bool
	data             []byte
}

type Pieces struct {
	pieces       []piece
	piece_length int
	total_length int
	hashes       []string
	client_chan  chan Message
}

func (p *Pieces) Length() int {
	return len(p.pieces)
}

func (p *Pieces) HaveAtIndex(i int) bool {
	return p.pieces[i].have
}

func (p *Pieces) pieceSize(i int) int {
	length := p.piece_length
	if i == p.Length()-1 {
		length = int(p.total_length % p.piece_length)
	}
	return length
}

func (p *Pieces) blockSize(i, o int) int {
	length := int(block_size)
	if i == p.Length()-1 {
		rem := int(p.total_length % p.piece_length)
		last_ind := int(rem / int(block_size))
		if o == last_ind {
			length = int(p.total_length % int(block_size))
		}
	}
	return length
}

func (p *Pieces) numBlocks(i int) int {
	rem := p.pieceSize(i) % int(block_size)
	num := p.pieceSize(i) / int(block_size)
	if rem > 0 {
		num++
	}
	return num
}

func (p *Pieces) setAtIndex(i int, b bool) {
	p.pieces[i].have = b
}

func (p *Pieces) initBlocksAtPiece(i int) {
	size := p.numBlocks(i)
	p.pieces[i].blocks = make([]bool, size)
	p.pieces[i].blocks_requested = make([]bool, size)
	p.pieces[i].data = make([]byte, p.pieceSize(i))
}

func (p *Pieces) lengthBlocksInPiece(i int) int {
	return p.numBlocks(i)
}

func (p *Pieces) RequestedPieceAndOffset(piece, offset int) {
	p.pieces[piece].requested = true
	if p.pieces[piece].blocks == nil {
		p.initBlocksAtPiece(piece)
	}
	p.pieces[piece].blocks_requested[offset/int(block_size)] = true
}

func (p *Pieces) requested(index, begin int) bool {
	if p.pieces[index].blocks == nil {
		p.initBlocksAtPiece(index)
	}
	return p.pieces[index].requested && p.pieces[index].blocks_requested[begin]
}

func (p *Pieces) String() string {
	s := []byte("")
	for _, v := range p.pieces {
		s = append(s, []byte(fmt.Sprintf("<%v>", v.have))...)
	}
	return string(s)
}
func (ours *Pieces) GetPieceAndOffsetForRequest(theirs *Pieces) (int, int) {
	indices := make([]int, 0)
	for i, p := range ours.pieces {
		// for an incomplete piece that is in progress, get
		// remaining blocks
		if !p.have && p.requested && theirs.pieces[i].have {
			for j, b := range p.blocks {
				if !b && !p.blocks_requested[j] {
					return i, (j * int(block_size))
				}
			}
		}
		// otherwise, let's make an array of missing pieces
		if !p.have && theirs.pieces[i].have {
			indices = append(indices, i)
		}
	}

	// if there are no missing pieces, return -1
	if len(indices) <= 0 {
		return -1, -1
	}

	// get random piece index and set up for requesting
	//ind := indices[0] 
	ind := indices[RandomInt(len(indices))]
	if ours.pieces[ind].blocks == nil {
		ours.initBlocksAtPiece(ind)
	}
	// get random block index for request
	indices = make([]int, 0)
	for i, b := range ours.pieces[ind].blocks {
		if !b {
			indices = append(indices, i)
		}
	}

	// if no blocks are un-filled return -1
	if len(indices) <= 0 {
		return -1, -1
	}

	// otherwise return random block offset too.
	//off := indices[0] 
	off := indices[RandomInt(len(indices))]
	ours.RequestedPieceAndOffset(ind, off*int(block_size))
	return ind, (off * int(block_size))
}

func (p *Pieces) HaveBlockAtPieceAndOffset(i, offset int) bool {
	if i >= p.Length() || offset >= p.lengthBlocksInPiece(i) {
		return false
	}
	return p.pieces[i].blocks[offset/int(block_size)]
}

func (p *Pieces) SetBlockAtPieceAndOffset(i int, offset int, b []byte) {
	if p.pieces[i].blocks == nil {
		p.initBlocksAtPiece(i)
	}
	if i >= p.Length() || (offset/int(block_size)) >= p.lengthBlocksInPiece(i) {
		fmt.Printf("Got a bad index: %d /offset: %d\n", i, offset)
		fmt.Printf("Compare to index: %d\n", p.Length())

		return
	}
	if len(b) < 16384 && i < (p.Length()-1) {
		fmt.Printf("Got a bad block")
		return
	}
	p.pieces[i].blocks[offset/int(block_size)] = true
	for j, by := range b {
		p.pieces[i].data[offset+j] = by
	}
	p.pieces[i].blocks_requested[offset/int(block_size)] = false
	p.checkPiece(i)
}

func (p *Pieces) checkPiece(i int) {
	for _, b := range p.pieces[i].blocks {
		if !b {
			return
		}
	}
	fmt.Println("Blocks: ", p.pieces[i].blocks)

	fmt.Println("Finished a block ", i)
	fmt.Println(p)

	h := sha1.New()
	h.Write(p.pieces[i].data[0:p.pieceSize(i)])
	hash := string(h.Sum(nil))

	for j := range hash {
		if hash[j] != p.hashes[i][j] {
			p.pieces[i].requested = false
			p.pieces[i].data = nil
			p.pieces[i].blocks = nil
			p.pieces[i].blocks_requested = nil
			fmt.Println("Bad Hash")
			return
		}
	}
	fmt.Println("Going to add to client_chan", len(p.client_chan))
	p.client_chan <- InternalWriteBlockMessage{p.pieces[i].data, i}

	p.pieces[i].have = true
	p.pieces[i].requested = false
}

func CreateNewPieces(num_pieces int, t *TorrentInfo) *Pieces {
	pieces := new(Pieces)
	pieces.piece_length = int(t.pieceLength)
	pieces.total_length = int(t.total_length)
	pieces.hashes = t.pieces
	pieces.pieces = make([]piece, num_pieces)
	pieces.client_chan = t.client_chan
	return pieces
}
