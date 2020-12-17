package main

import (
	"dht"
	"fmt"
	"log"
	"math/big"
	"os"
	"time"
	"bufio"
	"strings"
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

func main()  {
	id1 := "1"
	sister := dht.CreateNodeById(id1,"0.0.0.0:8001")

	h, err := createNode("12", "0.0.0.0:8004", sister)
	if err != nil{
		log.Fatalln(err)
		return
	}
	fmt.Print(">")
	input := bufio.NewScanner(os.Stdin)
	for input.Scan() {
		line := input.Text()
		array := strings.Fields(line)
		fmt.Println(array[0])
        
        if line == "bye" {          
            break
		}
		if array[0] == "add"{
			sErr := h.AddKey(array[1], array[2])
				if sErr != nil {
					log.Println("err: ", sErr)
				}
		}
		if array[0] == "delete"{
			err := h.DeleteKey(array[1])
			if err != nil{
				log.Println("err: ", err)
			}
		}
		if array[0] == "get"{
			val, err := h.GetValue(array[1])
			if err != nil{
				log.Println("err", err)
			}
			fmt.Println(val)
		}
		if array[0] == "getloc"{
			node, err := h.GetLocation(array[1])
			if err != nil{
				log.Println("err: ", err)
			}
			fmt.Println(node.NodeId)
		}
		fmt.Print(">")
	}

}
