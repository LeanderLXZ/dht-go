package main

import (
	"src"
	"log"
	"math/big"
	"os"
	"os/signal"
	"time"
)

func createNode(id string, addr string, sister *src.NodeRPC) (*src.Node, error) {

	p := src.Parameters()
	p.Id = id
	p.Addr = addr
	p.Timeout = 10 * time.Millisecond
	p.MaxIdle = 100 * time.Millisecond

	n, err := src.NewNode(p, sister)
	return n, err
}

func createID(id string) []byte {
	val := big.NewInt(0)
	val.SetString(id, 10)
	return val.Bytes()
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