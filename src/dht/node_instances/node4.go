package main

import (
	"dht"
	"log"
	"math/big"
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

func createID(id string) []byte {
	val := big.NewInt(0)
	val.SetString(id, 10)
	return val.Bytes()
}

func main() {

	joinNode := dht.CreateNodeById("4", "0.0.0.0:8002")

	h, err := createNode("12", "0.0.0.0:8004", joinNode)
	if err != nil {
		log.Fatalln(err)
		return
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	h.Stop()
}
