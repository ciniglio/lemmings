package tracker

import ()

func main() {
	c := NewClient()
	done := make(chan int)
	go c.Run()
	c.AddTorrent("test/test2.torrent")
	c.AddTorrent("test/test.torrent")
	<- done
}
