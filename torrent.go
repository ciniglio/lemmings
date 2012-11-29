package tracker

import (
	"fmt"
)

func LaunchTorrent(torrent_file string, done chan int) {
	c := make(chan Message, 1)
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

	for p := range msg {
		peer := CreatePeer(p, torrent, c)
		if peer != nil {
			go peer.runPeer()
		}
	}
	peers := make([]chan Message, 0)
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
				if torrent.our_pieces.SetBlockAtPieceAndOffset(msg.index,
					msg.begin,
					msg.block) {
					fw.messages <- InternalWriteBlockMessage{
						torrent.our_pieces.pieces[msg.index].data,
						msg.index,
					}
					broadcast(peers, InternalHaveMessage{msg.index})
				}
			case i_write_block:
				fmt.Println("About to write")
				fw.messages <- m
			case i_subscribe:
				fmt.Println("Subscribed a peer")
				msg := m.(InternalSubscribeMessage)
				peers = append(peers, msg.c)
			case i_request:
				msg := m.(InternalRequestMessage)
				req := msg.m
				b := torrent.our_pieces.GetBlockAtPieceAndOffset(req.index, req.begin, req.length)
				if b != nil {
					ret := new(PieceMessage)
					ret.index = req.index
					ret.begin = req.begin
					ret.block = b
					msg.ret <- ret
				} else {
					msg.ret <- nil
				}
			default:
				fmt.Println("Got weird internal request")
			}
		}
	}
	done <- 0
}

func broadcast(channels []chan Message, m Message) {
	fmt.Println("Number of broadcast channels:", len(channels))
	for i := range channels {
		go func(ch chan Message) {
			ch <- m
		}(channels[i])
	}
}
