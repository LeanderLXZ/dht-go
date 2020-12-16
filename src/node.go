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

func(node *Node)join(newNode rpc.Node) error {

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
//                  Local Methods
// ================================================

// ---------------- Node Operations ---------------

//  -----------------------------------------------
// 					find Successor 
// 			reference from paper fig. 5
// 			ask node n to find the successor of id
//	-----------------------------------------------
func(node *Node)findNextNode(nodeId []byte) (rpc.Node, error){
	node.succLock.RLock()
	defer node.succLock.RUnlock()
	currNode := node.Node
	succNode := node.successor

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
		preNode := node.closestPreNode(nodeId)
		if isEqual(preNode.nodeId, node.NodeID){
			succNode, err = node.getSuccessorRPC(preNode)
			if err != nil {
				return nil, err
			}
			if succNode == nil {
				return preNode, nil
			}
			return succNode, nil
		}

		succNode, err := node.findSuccessorRPC(preNode, nodeId)
		if err != nil {
			return nil, err
		}
		if succ == nil {
			return currNode, nil
		}
		return succNode, nil
	}
	return nil, nil
}

// Get the value given a key
func (node *Node) getValue(key string) ([]byte, error) {

}

//  -----------------------------------------------
// 				cloest predecessor node
// search the local table for the highest predecessor of id
// in fingerTable.
//	-----------------------------------------------
func(node *Node)closestPreNode(nodeId []byte) (){
	node.predLock.RLock()
	defer node.predLock.RUnlock()

	currNode := node.Node

	for i := len(node.finferTable) - 1; i>=0; i-- {
		v := node.fingerTable[i]
		if v == nil || v.Node == nil {
			continue
		}
		if between(v.nodeId, currNode,nodeId, nodeId) {
			return v.Node
		}
	}
	return currNode
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
func(node *Node) GetPreNode(ctx context.Context, r rpc.ER) (rpc.ER, error) {
	node.predLock.RLock()
	preNode := node.predecessor
	node.predLock.RUnlock()
	if preNode == nil {
		return emptyNode, nil
	}
	return preNode, nil
}

// set the predecessor node 
func(node *Node) SetPreNode(ctx context.Context, preNode rpc.Node) (rpc.ER, error) {
	node.predLock.Lock()
	node.preNode = preNode
	node.predLock.Unlock()
	return emptyRequest, nil
}

// get the successor node and return it 为什么不需要传值就能拿node？
func(node *Node) GetNextNode(ctx context.Context, r rpc.ER) (rpc.ER, error) {
	node.succLock.RLock()
	NextNode := node.successor
	node.succLock.RUnlock()
	if NextNode == nil {
		return emptyNode, nil
	}
	return NextNode, nil
}

// set the successor node
func(node *Node) SetPreNode(ctx context.Context, NextNode rpc.Node) (rpc.ER, error) {
	node.succLock.Lock()
	node.successor = NextNode
	node.succLock.Unlock()
	return emptyRequest, nil
}

func(node *Node) CheckPreNodeById(ctx context.Context, preNodeId rpc.NodeId) (rpc.Node, error) {
	preNode, err := node.
}


// ---------------- Key Operations ----------------




// ================================================
//                  RPC Protocols
// ================================================

// ---------------- Node Operations ---------------

// getNextNodeRPC the successor ID of a remote node.
func (node *Node) getNextNodeRPC(node1 *models.Node) (*models.Node, error) {
	return node.connections.GetNextNode(node1)
}

// setNextNodeRPC sets the successor of a given node.
func (node *Node) setNextNodeRPC(node1 *models.Node, nextNode *models.Node) error {
	return node.connections.SetNextNode(node1, nextNode)
}

// findNextNodeRPC finds the successor node of a given ID in the entire ring.
func (node *Node) findNextNodeRPC(node1 *models.Node, nodeId []byte) (*models.Node, error) {
	return node.connections.FindNextNode(node1, nodeId)
}

// getNextNodeRPC the successor ID of a remote node.
func (node *Node) getPredecessorRPC(node1 *models.Node) (*models.Node, error) {
	return node.connections.GetPredecessor(node1)
}

// setPredecessorRPC sets the predecessor of a given node.
func (node *Node) setPredecessorRPC(node1 *models.Node, preNode *models.Node) error {
	return node.connections.SetPredecessor(node1, preNode)
}

// Inform the node to be the previous node of current node
func (node *Node) notifyRPC(node1, predNode *models.Node) error {
	return node.connections.Inform(node1, preNode)
}

// ---------------- Key Operations ----------------

// Get the value given a key
func (node *Node) getValueRPC(node1 *models.Node, key string) (*models.GetResponse, error) {
	return node.connections.GetValue(node1, key)
}

// Add a (key, value) pair
func (node *Node) addKeyRPC(node1 *models.Node, key, value string) error {
	return node.connections.AddKey(node1, key, value)
}

// Get keys from a given range
func (node *Node) GetKeysRPC(node1 *models.Node, start []byte, end []byte) ([]*models.KV, error) {
	return node.connections.GetKeys(node1, start, end)
}

	// Delete a given key
func (node *Node) deleteKeyRPC(node1 *models.Node, key string) error {
	return node.connections.DeleteKey(node1, key)
}

// Delete multiple keys
func (node *Node) deleteKeysRPC(node1 *models.Node, keys []string) error {
	return node.connections.DeleteKeys(node1, keys)
}
