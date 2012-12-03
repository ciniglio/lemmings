package tracker

import (
	"fmt"
	"net"
	"strconv"
)

type Torrent struct {
	messages chan Message
}

func LaunchTorrent(torrent_file string, done chan int) (string, Torrent) {
	c := make(chan Message, 1)
	t := Torrent{c}
	go t.runTorrent(torrent_file, done)
	return FindInfoHash(torrent_file), t
}

func (t Torrent) runTorrent(torrent_file string, done chan int) {
	c := t.messages
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
		peer := CreatePeer(p, torrent, c, t)
		if peer != nil {
			go peer.RunPeer()
		}
	}
	peers := make([]chan Message, 0)
	num_unchoked := 0
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
					broadcast(peers, InternalCancelMessage{msg.index,
					        msg.begin,
						len(msg.block),
					})
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
			case i_add_peer:
				msg := m.(InternalAddPeerMessage)
				ip, port, err := net.SplitHostPort(msg.c.RemoteAddr().String())
				if err != nil {
					fmt.Println("Error splitting host and port",
						msg.c.RemoteAddr())
				}
				iport, _ := strconv.Atoi(port)	
				p := torrentPeer{msg.peer_id, 
					ip,
					iport,
				}
				peer := CreatePeer(p, torrent, c, t)
				peer.connection = msg.c
				peer.shook_hands = true
				peer.connected = true
				go peer.RunPeer()
			case i_can_unchoke:
				msg := m.(InternalCanUnchokeMessage)
				if num_unchoked < 5 {
					num_unchoked++
					msg.ret <- true
				} else {
					msg.ret <- false
				}
			case i_will_choke:
				num_unchoked--
			default:
				fmt.Println("Got weird internal request")
			}
		}
	}
	done <- 0
}

func (t Torrent) CanUnchoke() bool {
	c := make(chan bool)
	t.messages <- InternalCanUnchokeMessage{c}
	return <- c
}

func (t Torrent) WillChoke() {
	t.messages <- InternalChokingMessage{}
}

func broadcast(channels []chan Message, m Message) {
	fmt.Println("Number of broadcast channels:", len(channels))
	for i := range channels {
		go func(ch chan Message) {
			ch <- m
		}(channels[i])
	}
}
