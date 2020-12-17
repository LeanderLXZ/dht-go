package main

import (
	"fmt"
	"dht"
	"log"
	"math/big"
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
	p.MaxIdleTime = 100 * time.Millisecond

	n, err := dht.CreateNode(p, sister)
	return n, err
}

func createID(id string) []byte {
	val := big.NewInt(0)
	val.SetString(id, 10)
	return val.Bytes()
}

func main() {

	id1 := "1"
	sister := chord.NewInode(id1, "0.0.0.0:8001")

	h, err := createNode("4", "0.0.0.0:8002", sister)
	if err != nil {
		log.Fatalln(err)
		return
	}

	shut := make(chan bool)
	var count int
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		for {
			select {
			case <-ticker.C:
				count++
				key := strconv.Itoa(count)
				value := fmt.Sprintf(`{"graph_id" : %d, "nodes" : ["node-%d","node-%d","node-%d"]}`, count, count+1, count+2, count+3)
				sErr := h.Set(key, value)
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