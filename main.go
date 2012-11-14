package tracker

import ()

func main() {
	torrent := ReadTorrentFile("test/test.torrent")
	tracker_proxy := CreateTrackerProxy(torrent)
	peers_info := tracker_proxy.GetPeers()
	for _, p := range peers_info {
		CreatePeer(&p, torrent)
	}

}
