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

	peers := make([]Peer, 0)

	for p := range msg {
		peer := CreatePeer(p, torrent, c)
		if peer != nil {
			peers = append(peers, *peer)
			go peer.runPeer()
		}
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
			for i := range peers {
				peers[i].messageChannel <- InternalReceivedBlockMessage{index: msg.index, begin: msg.begin}
			}

		default:
			fmt.Println("Got weird internal request")
		}
	}
}
