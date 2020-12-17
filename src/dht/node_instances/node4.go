package main

import (
	"bufio"
	"dht"
	"fmt"
	"log"
	"math/big"
	"os"
	"os/signal"
	"strings"
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

	h, err := createNode("12", "0.0.0.0:8004", sister)
	if err != nil {
		log.Fatalln(err)
		return
	}
	input := bufio.NewScanner(os.Stdin)
	time.Sleep(4000 * time.Millisecond)
	fmt.Println("----------------------------------------------------------------")
	fmt.Print("Input: ")
	go func() {
		for input.Scan() {
			line := input.Text()
			array := strings.Fields(line)

			if line == "bye" {
				break
			}
			if array[0] == "add" {
				sErr := h.AddKey(array[1], array[2])
				if sErr != nil {
					log.Println("err: ", sErr)
				}
			}
			if array[0] == "delete" {
				err := h.DeleteKey(array[1])
				if err != nil {
					log.Println("err: ", err)
				}
			}
			if array[0] == "get" {
				val, err := h.GetValue(array[1])
				if err != nil {
					log.Println("err", err)
				}
				fmt.Printf("Output: ")
				fmt.Printf("%s\n", string(val))
			}
			if array[0] == "getloc" {
				node, err := h.GetLocation(array[1])
				if err != nil {
					log.Println("err: ", err)
				}
				fmt.Printf("Output: ")
				fmt.Printf("%d\n", (&big.Int{}).SetBytes(node.NodeId))
			}
			fmt.Println("----------------------------------------------------------------")
			fmt.Print("Input: ")
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	h.Stop()

}
