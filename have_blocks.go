package tracker

import (
	"crypto/sha1"
	"fmt"
	"time"
)

type piece struct {
	have             bool
	requested        bool
	blocks           []bool
	blocks_requested []bool
	b_requested_at   []time.Time
	data             []byte
}

type Pieces struct {
	pieces       []piece
	num_have     int
	piece_length int
	total_length int
	hashes       []string
	client_chan  chan Message
}

func (p *Pieces) String() string {
	s := []byte("")
	for _, v := range p.pieces {
		s = append(s, []byte(fmt.Sprintf("<%v>", v.have))...)
	}
	return string(s)
}

func (p *Pieces) Length() int {
	return len(p.pieces)
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
	p.pieces[i].b_requested_at = make([]time.Time, size)
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
	p.pieces[piece].b_requested_at[offset/int(block_size)] = time.Now()
}

func (p *Pieces) requested(index, begin int) bool {
	if p.pieces[index].blocks == nil {
		p.initBlocksAtPiece(index)
	}
	return p.pieces[index].requested && p.pieces[index].blocks_requested[begin]
}


func (ours *Pieces) GetBlockAtPieceAndOffset(i, o, l int) []byte {
	size := ours.blockSize(i, o)

	if i > ours.Length() || !ours.pieces[i].have || l > (size - o) {
		return nil
	}

	return ours.pieces[i].data[o : o+l]
}

func (ours *Pieces) CreateBitField() (b []byte) {
	b = make([]byte, 0)
	var i uint
	for i = 0; i < uint(len(ours.pieces)); i += 8 {
		by := byte(0)
		var j uint
		for j = 0; j < 8; j++ {
			if i+j >= uint(len(ours.pieces)) {
				break
			}
			if ours.pieces[i+j].have {
				by = by | (1 << (7 - j))
			}
		}
		b = append(b, by)
	}
	return b
}

func (p *Pieces) AddHave(i int) {
	p.num_have++
	p.setAtIndex(i, true)
}

func (p *Pieces) AddBitField(b []byte) {
	ind := 0
	for i := range b {
		for j := 7; j >= 0; j-- {
			have := ((b[i]>>uint(j))&1 == 1)
			p.setAtIndex(ind, have)
			ind++
			if ind >= p.Length() {
				return
			}
		}
	}
}

func (p piece) needBlock(o, rem int) bool {
	max_age := 2 * time.Minute
	have := p.blocks[o]
	req  := p.blocks_requested[o]
	stale := time.Since(p.b_requested_at[o]) > max_age
	if rem >5 {
		return !have && (!req || stale)
	}
	return !have
}

func (ours *Pieces) GetPieceAndOffsetForRequest(theirs *Pieces) (int, int) {
	remaining := ours.Length() - ours.num_have 
	indices := make([]int, 0)
	for i, p := range ours.pieces {
		// for an incomplete piece that is in progress, get
		// remaining blocks
		if !p.have && p.requested && theirs.pieces[i].have {
			for j := range p.blocks {
				if p.needBlock(j, remaining) {
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

func (p *Pieces) SetBlockAtPieceAndOffset(i int, offset int, b []byte) bool {
	if p.pieces[i].blocks == nil {
		p.initBlocksAtPiece(i)
	}
	if i >= p.Length() || (offset/int(block_size)) >= p.lengthBlocksInPiece(i) {
		fmt.Printf("Got a bad index: %d /offset: %d\n", i, offset)
		fmt.Printf("Compare to index: %d\n", p.Length())

		return false
	}
	if len(b) != 16384 && i < (p.Length()-1) {
		fmt.Printf("Got a bad block")
		return false
	}
	p.pieces[i].blocks[offset/int(block_size)] = true
	for j, by := range b {
		p.pieces[i].data[offset+j] = by
	}
	p.pieces[i].blocks_requested[offset/int(block_size)] = false
	return p.checkPiece(i)
}

func (p *Pieces) checkPiece(i int) bool {
	for _, b := range p.pieces[i].blocks {
		if !b {
			return false
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
			return false
		}
	}
	fmt.Println("Going to add write message to client_chan", len(p.client_chan))

	p.setAtIndex(i, true)
	p.pieces[i].requested = false
	return true
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
