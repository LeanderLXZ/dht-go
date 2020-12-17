package main

import (
	"dht"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
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

	id1 := "1"
	sister := dht.CreateNodeById(id1, "0.0.0.0:8001")

	h, err := createNode("4", "0.0.0.0:8002", sister)
	if err != nil {
		log.Fatalln(err)
		return
	}

	shut := make(chan bool)
	var count int
	time.Sleep(2000 * time.Millisecond)
	go func() {
		ticker := time.NewTicker(1000 * time.Millisecond)
		for {
			select {
			case <-ticker.C:
				count++
				key := strconv.Itoa(count)
				value := fmt.Sprintf("Hello, this is node2!")
				sErr := h.AddKey(key, value)
				if sErr != nil {
					log.Println("err: ", sErr)
				}
			case <-shut:
				ticker.Stop()
				return
			}
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	shut <- true
	h.Stop()
}
