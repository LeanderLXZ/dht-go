package main

import (
	"dht"
	"log"
	"os"
	"os/signal"
	"time"
)

func createNode(id string, addr string, sister *dht.NodeRPC) (*dht.Node, error) {

	p := dht.GetInitialParameters()
	p.NodeId = id
	p.Address = addr
	p.Timeout = 10 * time.Millisecond
	p.MaxIdleTime = 1000 * time.Millisecond

	n, err := dht.CreateNode(p, sister)
	return n, err
}

func main() {

	h, err := createNode("1", "0.0.0.0:8001", nil)
	if err != nil {
		log.Fatalln(err)
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-time.After(10 * time.Second)
	<-c
	h.Stop()

}
