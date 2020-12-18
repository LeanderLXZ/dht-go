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

	h, err := createNode("12", "0.0.0.0:8004", nil)
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

			if line == "q" {
				break
			}
			if array[0] == "add" {
				if len(array) != 3 {
					fmt.Println("Incorrect input! Please input again.")
					goto JUMPPOINT
				}
				sErr := h.AddKey(array[1], array[2])
				if sErr != nil {
					log.Println("err: ", sErr)
				}
			} else if array[0] == "del" {
				if len(array) != 1 {
					fmt.Println("Incorrect input! Please input again.")
					goto JUMPPOINT
				}
				err := h.DeleteKey(array[1])
				if err != nil {
					log.Println("err: ", err)
				}
			} else if array[0] == "get" {
				if len(array) != 2 {
					fmt.Println("Incorrect input! Please input again.")
					goto JUMPPOINT
				}
				val, err := h.GetValue(array[1])
				if err != nil {
					log.Println("err", err)
				}
				fmt.Printf("Output: ")
				if val != nil {
					fmt.Printf("%s\n", string(val))
				} else {
					fmt.Printf("Key not found\n")
				}
			} else if array[0] == "getloc" {
				if len(array) != 2 {
					fmt.Println("Incorrect input! Please input again.")
					goto JUMPPOINT
				}
				node, err := h.GetLocation(array[1])
				if err != nil {
					log.Println("err: ", err)
				}
				fmt.Printf("Output: ")
				if node != nil {
					fmt.Printf("%d\n", (&big.Int{}).SetBytes(node.NodeId))
				} else {
					fmt.Printf("Key not found\n")
				}
			} else if array[0] == "ft" {
				if len(array) != 1 {
					fmt.Println("Incorrect input! Please input again.")
					goto JUMPPOINT
				}
				fts := h.GetFingerTable()
				if fts != nil {
					fmt.Printf("Output: \n")
					for i := 0; i < len(fts); i++ {
						ft := fts[i]
						fmt.Printf(
							"ID: %d | Remote Node ID: %d\n",
							(&big.Int{}).SetBytes(ft.InitId),
							(&big.Int{}).SetBytes(ft.RemoteNode.NodeId),
						)
					}
					fmt.Println("-----------------------------")
					fmt.Printf("Length of finger table: %d\n", len(fts))
				}
			} else {
				fmt.Println("Incorrect input! Please input again.")
			}
		JUMPPOINT:
			fmt.Println("----------------------------------------------------------------")
			fmt.Print("Input: ")
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	h.Stop()

}
