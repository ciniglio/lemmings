package tracker

import (
	"fmt"
)

func main() {
	torrent_file := "test/test2.torrent"
	c := make(chan Message, 1000)
	torrent, err := ReadTorrentFile(torrent_file, c)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
	tracker_proxy := NewTrackerProxy(torrent)
	msg := make(chan torrentPeer)
	get_peers := InternalGetPeersMessage{ret: msg}
	tracker_proxy.msg <- get_peers

	fw := NewFileWriter(torrent, torrent_file)
	go fw.Run()

	peers := make([]chan Message, 0)

	for p := range msg {
		peer := CreatePeer(p, torrent, c)
		if peer != nil {
			go peer.runPeer()
		}
	}
	for {
		select {
		case m := <-c:
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
			case i_write_block:
				fmt.Println("About to write")
				fw.messages <- m
			case i_subscribe:
				msg := m.(*InternalSubscribeMessage)
				peers = append(peers, msg.c)
			default:
				fmt.Println("Got weird internal request")
			}
		}
	}
}
