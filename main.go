package tracker

import (
	"fmt"
)

func main() {
	torrent, err := ReadTorrentFile("test/test.torrent")
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
	tracker_proxy := NewTrackerProxy(torrent)
	msg := make(chan torrentPeer)
	get_peers := InternalGetPeersMessage{ret: msg}
	tracker_proxy.msg <- get_peers

	c := make(chan Message)

	for p := range msg {
		go CreatePeer(p, torrent, c)
	}

	for m := range c {
		switch m.kind() {
		case i_get_request:
			msg := m.(*InternalGetRequestMessage)
			i, b := torrent.our_pieces.GetPieceAndOffsetForRequest(msg.pieces)
			msg.ret <- i
			msg.ret <- b
		case i_sent_request:
			msg := m.(*InternalSendingRequestMessage)
			i := msg.index
			b := msg.begin
			torrent.our_pieces.RequestedPieceAndOffset(i, b)
		case piece_t:
			msg := m.(PieceMessage)
			torrent.our_pieces.SetBlockAtPieceAndOffset(msg.index, msg.begin, msg.block)
		default:
			fmt.Println("Got weird internal request")
		}
	}
}
