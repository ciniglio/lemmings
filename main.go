package tracker

import (
	"fmt"
)

func main() {
	torrent := ReadTorrentFile("test/test.torrent")
	tracker_proxy := CreateTrackerProxy(torrent)
	peers_info := tracker_proxy.GetPeers()
	c := make(chan Message)
	for _, p := range peers_info {
		fmt.Println("CREATING NEW PEER:", peers_info)
		go CreatePeer(p, torrent, c)
	}

	for m := range c {
		switch m.kind() {
		case i_get_request:
			msg := m.(*InternalGetRequestMessage)
			i, b := torrent.our_pieces.GetPieceAndOffsetForRequest(msg.pieces)
			msg.ret <- i
			msg.ret <- b
		default:
			fmt.Println("Got weird internal request")
		}
	}
}
