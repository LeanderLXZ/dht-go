package src

import (
	"crypto/sha1"
	"hash"
	"sync"
	"time"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	rpc "./rpc.pb.go"
)

// Structure for parameters
type Parameters struct {
	NodeID        string
	Address       string
	HashFunc      func() hash.Hash 		// Hash function
	HashLen       int              		// Length of hash
	Timeout       time.Duration    		// Timeout
	MaxIdleTime   time.Duration    		// Maximum idle time
	MinStableTime time.Duration    		// Minimum stable time
	MaxStableTime time.Duration    		// Maximum stable time
	ServerOptions []grpc.ServerOption 	// grpc server option
	DialOptions   []grpc.DialOption 	// grpc dial option
}

// Get a initial parameters settings
func GetInitialParameters() *Parameters {
	param := Parameters{}
	param.HashFunc = sha1.New
	param.HashLen = param.HashFunc.Size() * 8
	param.DialOptions = make([grpc.DialOption, 0, 5])
	param.DialOptions = append(
		param.DialOptions,
		grpc.WithBloc(),
		grpc.WithTimeout(5*time.Second),
		grpc.FailOnNonTempDialError(true),
		grpc.WithInsecure()
	)
	return param
}

func(n *Node)join(newNode rpc.Node) error {

}

// Structure of Node
type Node struct {
	rpc.Node
	para 			*Parameters

	predecessor 	rpc.Node
	predLock		sync.RWMutex

	successor		rpc.Node
	succLock		sync.RWMutex

	fingerTable 	fingerTable
	fingerLock		sync.RWMutex

	storage 		Storage
	stLock			sync.RWMutex

	connections 	Connections
	tsLock     		sync.RWMutex

	lastStablized 	time.Time
}
// ================================================
//                  RPC Callers
// ================================================

func (n *Node) getSuccessorRPC(node rpc.Node) (rpc.Node, error) {
	return n.transport.GetSuccessor(node)
}

// ================================================
//                  Local Methods
// ================================================

// ---------------- Node Operations ---------------

//  -----------------------------------------------
// 					findSuccessor 
// reference from paper fig. 5
// ask node n to find the successor of id
//	-----------------------------------------------
func(n *Node)findNextNode(nodeId []byte) (rpc.Node, error){
	n.succLock.RLock()
	defer n.succLock.RUnlock()
	currNode := n.Node
	succNode := n.successor
	// if no succNode exists
	if succNode == nil {
		return currNode, nil
	}
	
	var err error
	// ask direct predecessor for node with nodeId, 
	// if not found, then go to find closest preceding node of nodeId in fingertable. 
	if betweenRightIncl(nodeId, currNode.nodeId, succNode.nodeId){
		return succNode, nil
	} else {
		preNode := n.closestPreNode(nodeId)
		return preNode.successor
	}
}

// Get the value given a key
func (node *Node) getValue(key string) ([]byte, error) {

}


func(n *Node)closestPreNode(nodeId []byte) (){

}

// ---------------- Key Operations ----------------

// generate hash key
// input ip address, output hash key (sha1)
func (node *Node) getHashKey(key string) ([]byte, error) {
	hash := node.cnf.Hash()
	if _, err := hash.Write([]byte(key)); err != nil {
		return nil, err
	}
	hashKey := hash.Sum(nil)
	return hashKey, nil
}




// ================================================
//                  Public Methods
// ================================================

// ---------------- Node Operations ---------------


// get the predecessor node and return it
func(n *Node) GetPreNode(ctx context.Context, r rpc.ER) (rpc.ER, error) {
	n.predLock.RLock()
	preNode := n.predecessor
	n.predLock.RUnlock()
	if preNode == nil {
		return emptyNode, nil
	}
	return preNode, nil
}

// set the predecessor node 
func(n *Node) SetPreNode(ctx context.Context, preNode rpc.Node) (rpc.ER, error) {
	n.predLock.Lock()
	n.preNode = preNode
	n.predLock.Unlock()
	return emptyRequest, nil
}

// get the successor node and return it 为什么不需要传值就能拿node？
func(n *Node) GetNextNode(ctx context.Context, r rpc.ER) (rpc.ER, error) {
	n.succLock.RLock()
	NextNode := n.successor
	n.succLock.RUnlock()
	if NextNode == nil {
		return emptyNode, nil
	}
	return NextNode, nil
}

// set the successor node
func(n *Node) SetPreNode(ctx context.Context, NextNode rpc.Node) (rpc.ER, error) {
	n.succLock.Lock()
	n.successor = NextNode
	n.succLock.Unlock()
	return emptyRequest, nil
}

func(n *Node) CheckPreNodeById(ctx context.Context, preNodeId rpc.NodeId) (rpc.Node, error) {
	preNode, err := n.
}

//
func (n *Node) SetPreNodeRPC(node rpc.Node){

}

// ---------------- Key Operations ----------------


