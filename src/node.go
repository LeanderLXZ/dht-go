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

// findNextNode
// reference from paper fig. 5
func(n *node)findNextNode(nodeId []byte) (rpc.Node, error){
	n.succLock.RLock()
	defer n.succLock.RUnlock()

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


// findNextNode
// reference from paper fig. 5
// ask node n to find the successor of id
func(node *Node)findNextNode(nodeId []byte) (rpc.Node, error){
	n.succLock.RLock()
	defer n.succLock.RUnlock()
	currNode := n.Node
	succNode := n.successor

	// if 
	if succNode == nil {
		return currNode, nil
	}

	var err error

	if betweenRightIncl(nodeId, currNode.nodeId, succNode.nodeId){
		return succNode,nil
	} else {
		preNode := n.closestPreNode(nodeId)
	}

// Get the value given a key
func (node *Node) getValue(key string) ([]byte, error) {
	n, err := node.


}


func(node *Node)closestPreNode(nodeId []byte) (){

}

// ================================================
//                  Public Methods
// ================================================

// ---------------- Node Operations ---------------


// get the predecessor node and return it
func(node *Node) GetPreNode(ctx context.Context, r rpc.ER) (rpc.ER, error) {
	n.predLock.RLock()
	preNode := n.predecessor
	n.predLock.RUnlock()
	if preNode == nil {
		return emptyNode, nil
	}
	return preNode, nil
}

// set the predecessor node 
func(node *Node) SetPreNode(ctx context.Context, preNode rpc.Node) (rpc.ER, error) {
	n.predLock.Lock()
	n.preNode = preNode
	n.predLock.Unlock()
	return emptyRequest, nil
}

// get the successor node and return it 为什么不需要传值就能拿node？
func(node *Node) GetNextNode(ctx context.Context, r rpc.ER) (rpc.ER, error) {
	n.succLock.RLock()
	NextNode := n.successor
	n.succLock.RUnlock()
	if NextNode == nil {
		return emptyNode, nil
	}
	return NextNode, nil
}

// set the successor node
func(node *Node) SetPreNode(ctx context.Context, NextNode rpc.Node) (rpc.ER, error) {
	n.succLock.Lock()
	n.successor = NextNode
	n.succLock.Unlock()
	return emptyRequest, nil
}

func(node *Node) CheckPreNodeById(ctx context.Context, preNodeId rpc.NodeId) (rpc.Node, error) {
	preNode, err := n.
}

//
func (node *Node) SetPreNodeRPC(node rpc.Node){

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
func (node *Node) setNextNodeRPC(node *models.Node, succ *models.Node) error {
	return node.connections.SetNextNode(node, succ)
}

// findNextNodeRPC finds the successor node of a given ID in the entire ring.
func (node *Node) findNextNodeRPC(node *models.Node, id []byte) (*models.Node, error) {
	return node.connections.FindNextNode(node, id)
}

// getNextNodeRPC the successor ID of a remote node.
func (node *Node) getPredecessorRPC(node *models.Node) (*models.Node, error) {
	return node.connections.GetPredecessor(node)
}

// setPredecessorRPC sets the predecessor of a given node.
func (node *Node) setPredecessorRPC(node *models.Node, pred *models.Node) error {
	return node.connections.SetPredecessor(node, pred)
}

// Inform the node to be the previous node of current node
func (node *Node) notifyRPC(node1, predNode *models.Node) error {
	return node.connections.Inform(node1, pred)
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
