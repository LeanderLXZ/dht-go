package src

import (
	"fmt"
	"math/big"
)

type fingerTable []*fingerEntry

func newFingerTable(node *NodeRPC, m int) fingerTable {
	ft := make([]*fingerEntry, m)
	for i := range ft {
		ft[i] = newFingerEntry(FingerMath(node.NodeId, i, m), node)
	}

	return ft
}

type fingerEntry struct {
	InitId     []byte
	RemoteNode *NodeRPC
}

func newFingerEntry(initId []byte, remoteNode *NodeRPC) *fingerEntry {
	return &fingerEntry{
		InitId:     initId,
		RemoteNode: remoteNode,
	}
}

func (node *Node) findNextFinger(next int) int {
	nextHash := FingerMath(node.NodeId, next, node.para.HashLen)
	nextOne, errors := node.findNextNode(nextHash)
	nextNum := (next + 1) % node.para.HashLen
	if errors != nil || nextOne == nil {
		fmt.Println("error: ", errors, nextOne)
		fmt.Printf("finger lookup failed %x %x \n", node.NodeId, nextHash)
		return nextNum
	}

	finger := newFingerEntry(nextHash, nextOne)
	node.fingerLock.Lock()
	node.fingerTable[next] = finger
	node.fingerLock.Unlock()

	return nextNum
}

func FingerMath(node []byte, i int, m int) []byte {

	idInt := (&big.Int{}).SetBytes(node)

	// to get offset
	twoOffset := big.NewInt(2)
	newOffset := big.Int{}
	newOffset.Exp(twoOffset, big.NewInt(int64(i)), nil)

	sum := big.Int{}
	sum.Add(idInt, &newOffset)

	// to get the value of sceiling
	coo := big.Int{}
	coo.Exp(twoOffset, big.NewInt(int64(m)), nil)

	// use Mod
	idInt.Mod(&sum, &coo)

	// total sum
	return idInt.Bytes()
}
