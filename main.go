package tracker

import (
)

func main() {
	c := make(chan int)
	go LaunchTorrent("test/test2.torrent", c)
	<- c
}
