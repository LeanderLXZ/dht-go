package chord

import (
	"fmt"
	"math/big"

	"github.com/project-schrodinger/models" //这个models我们还没加在框架里，后面需要加
)

type fingerTable []*fingerEntry

func newFingerTable(node *models.Node, m int) fingerTable {
	ft := make([]*fingerEntry, m)
	for i := range ft {
		ft[i] = newFingerEntry(FingerMath(node.Id, i, m), node)
	}

	return ft
}

type fingerEntry struct {
	InitId     []byte
	RemoteNode *models.Node
}

func newFingerEntry(initId []byte, remoteNode *models.Node) *fingerEntry {
	return &fingerEntry{
		InitId:     initId,
		RemoteNode: remoteNode,
	}
}

func (node *Node) findNextFinger(next int) int {
	nextHash := FingerMath(node.Id, next, n.cnf.HashSize)
	succ, err := node.findSuccessor(nextHash)
	nextNum := (next + 1) % node.cnf.HashSize
	if err != nil || succ == nil {
		fmt.Println("error: ", err, succ)
		fmt.Printf("finger lookup failed %x %x \n", node.Id, nextHash)
		return nextNum
	}

	finger := newFingerEntry(nextHash, succ)
	node.ftMtx.Lock()
	node.fingerTable[next] = finger

	node.ftMtx.Unlock()

	return nextNum
}

func FingerMath(node []byte, i int, m int) []byte {

	// Convert the ID to a bigint
	idInt := (&big.Int{}).SetBytes(node)

	// Get the offset
	two := big.NewInt(2)
	offset := big.Int{}
	offset.Exp(two, big.NewInt(int64(i)), nil)

	// Sum
	sum := big.Int{}
	sum.Add(idInt, &offset)

	// Get the ceiling
	ceil := big.Int{}
	ceil.Exp(two, big.NewInt(int64(m)), nil)

	// Apply the mod
	idInt.Mod(&sum, &ceil)

	// Add together
	return idInt.Bytes()
}
