package src

import (
	"crypto/sha1"
	"hash"
	"sync"
	"time"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	rpc "./rpc.pb.go"
	"math/big"
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

// hashsize bigger than hash func size
func (para *Parameters) Verify() erro {
	return nil
}

// Get a initial parameters settings
func GetInitialParameters() *Parameters {
	para := Parameters{}
	para.HashFunc = sha1.New
	para.HashLen = param.HashFunc.Size() * 8
	para.DialOptions = make([grpc.DialOption, 0, 5])
	para.DialOptions = append(
		param.DialOptions,
		grpc.WithBloc(),
		grpc.WithTimeout(5*time.Second),
		grpc.FailOnNonTempDialError(true),
		grpc.WithInsecure()
	)
	return para
}

func(node *Node)join(newNode rpc.Node) error {

}

// Structure of Node
type Node struct {
	rpc.Node
	para 			*Parameters

	closeCh			chan struct{}

	predecessor 	rpc.Node
	predLock		sync.RWMutex

	successor		rpc.Node
	succLock		sync.RWMutex

	fingerTable 	fingerTable
	fingerLock		sync.RWMutex

	dataStorage 	Storage
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
// 					create a new node 
// 	create a new node and peridoically stablize it
// 			return error if node already exists		
//	-----------------------------------------------
func CreateNode(para *Parameters, newNode rpc.Node) (*Node, error) {
	if err := para.Verify(); err != nil {
		return nil, err
	}

	node := &Node {
		Node:			new(rpc.Node),
		closeCh:		make(chan struct{}),
		para:			parameters,
		dataStorage:	NewMapStore(para.HashFunc),
	}

	var nodeId string
	if para.NodeID != "" {
		nodeId = para.NodeID
	} else {
		nodeId = para.Address
	}
	hashId, err := node.getHashKey(nodeId)
	if err != nil {
		return nil, err
	}

	SInt := (&big.Int{}).SetBytes(hashId)
	// fmt.Printf("new node id %d, \n", SInt)

	node.Node.NodeId = nodeId
	node.Node.Address = para.Address

	// create fingertable for new node
	node.fingerTable = newFingerTable(node.Node, para.HashLen)

	// start RPC
	connect, err = NewGrpcConnection(para)
	if err != nil {
		return nil, err
	}
	node.connections = connect

	rpc.RegisterDistributedHashTableServer(connect.server, node)

	node.connections.Start()

	if err := node.join(newNode); err != nil {
		return nil, err
	}

	newNodePeriod(node)

	return node, nil
}

// node period to stablize, check the finger table, and check predecessor status
func newNodePeriod(node *Node) {
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		for {
			select {
			case <-ticker.C:
				node.stabilize()
			case <-node.shutdownCh:
				ticker.Stop()
				return
			}
		}
	}()
	go func() {
		next := 0
		ticker := time.NewTicker(100 * time.Millisecond)
		for {
			select {
			case <-ticker.C:
				next = node.fixFinger(next)
			case <-node.shutdownCh:
				ticker.Stop()
				return
			}
		}
	}()
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		for {
			select {
			case <-ticker.C:
				node.checkPredecessor()
			case <-node.shutdownCh:
				ticker.Stop()
				return
			}
		}
	}()
}

//  -----------------------------------------------
// 					findNextNode
// 			reference from paper fig. 5
// 			ask node n to find the successor of id
//	-----------------------------------------------
func(node *Node) findNextNode(nodeId []byte) (rpc.Node, error){
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
			succNode, err = node.getNextNodeRPC(preNode)
			if err != nil {
				return nil, err
			}
			if succNode == nil {
				return preNode, nil
			}
			return succNode, nil
		}

		succNode, err := node.getNextNodeByIdRPC(preNode, nodeId)
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
func(node *Node) closestPreNode(nodeId []byte) (){
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

// get the location of a given key
func (node *Node) getLocation(key string) (*NodeRPC, error) {
	nodeId, err := node.hashKey(key)
	if err != nil {
		return nil, err
	}
	nextNode, err := node.findNextNode(nodeId)
	return nextNode, err
}

// Add a (key, value) pair
func (node *Node) addKey(key, value string) error {
	node1, err := node.getLocation(key)
	if err != nil {
		return err
	}
	err = node.addKeyRPC(node1, key, value)
	return err
}

// get the value of a given key
func (node *Node) getValue(key string) ([]byte, error) {
	node1, err := node.getLocation(key)
	if err != nil {
		return nil, err
	}
	value, err := node.getValueRPC(node1, key)
	if err != nil {
		return nil, err
	}
	return value.Value, nil
}

// Delete a given key
func (node *Node) deleteKey(key string) error {
	node1, err := node.getLocation(key)
	if err != nil {
		return err
	}
	err = node.deleteKeyRPC(node1, key)
	return err
}

// When a new node is added to the ring, it gets the keys from the next node
func (node *Node) getKeys(preNode, nextNode *NodeRPC) ([]KeyValuePair, error) {
	if isEqual(node.nodeId, nextNode.nodeId) {
		return nil, nil
	}
	return node.getKeysRPC(
		nextNode, preNode.nodeId, node.nodeId,
	)
}

// delete given keys from the given model
func (node *Node) deleteKeys(node1 *NodeRPC, keys []string) error {
	return node.deleteKeysRPC(node1, keys)
}

// change the location of the keys
func (node *Node) changeKeys(preNode, nextNode *NodeRPC) {

	keys, err := node.getKeys(preNode, nextNode)
	if len(keys) > 0 {
		fmt.Println("Changing the location of key: ", keys, err)
	}
	keysDeleted := make([]string, 0, 10)
	for _, kv := range keys {
		if kv == nil {
			continue
		}
		node.storage.AddKey(kv.Key, kv.Value)
		keysDeleted = append(keysDeleted, kv.Key)
	}
	// delete keys from next node
	if len(keysDeleted) > 0 {
		node.deleteKeys(nextNode, keysDeleted)
	}

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

func (node *Node) GetLocation(key string) (*NodeRPC, error) {
	return node.getLocation(key)
}

func (node *Node) GetValue(key string) ([]byte, error) {
	return node.getValue(key)
}

func (node *Node) AddKey(key, value string) error {
	return node.addKey(key, value)
}

// Delete a given key
func (node *Node) DeleteKey(key string) error {
	return node.deleteKey(key)
}

// ================================================
//                  RPC Protocols
// ================================================

// ---------------- Node Operations ---------------

// Get the previous node of current node
func (node *Node) getPreNodeRPC(node1 *NodeRPC) (*NodeRPC, error) {
	return node.connections.GetPreNode(node1)
}

// Set the previous node for a given node
func (node *Node) setPreNodeRPC(node1 *NodeRPC, preNode *NodeRPC) error {
	return node.connections.SetPreNode(node1, preNode)
}

// Get the next node of current node
func (node *Node) GetNextNodeRPC(node1 *NodeRPC) (*NodeRPC, error) {
	return node.connections.GetNextNode(node1)
}

// Get the next node given an id
func (node *Node) getNextNodeByIdRPC(node1 *NodeRPC, nodeId []byte) (*NodeRPC, error) {
	return node.connections.GetNextNodeById(node1, nodeId)
}

// Set the next node of a given node
func (node *Node) setNextNodeRPC(node1 *NodeRPC, nextNode *NodeRPC) error {
	return node.connections.SetNextNode(node1, nextNode)
}

// Inform the node to be the previous node of current node
func (node *Node) informRPC(node1, predNode *NodeRPC) error {
	return node.connections.Inform(node1, preNode)
}

// ---------------- Key Operations ----------------

// Get the value given a key
func (node *Node) getValueRPC(node1 *NodeRPC, key string) (*GetValueResp, error) {
	return node.connections.GetValue(node1, key)
}

// Add a (key, value) pair
func (node *Node) addKeyRPC(node1 *NodeRPC, key, value string) error {
	return node.connections.AddKey(node1, key, value)
}

// Get keys from a given range
func (node *Node) GetKeysRPC(node1 *NodeRPC, start []byte, end []byte) ([]KeyValuePair, error) {
	return node.connections.GetKeys(node1, start, end)
}

// Delete a given key
func (node *Node) deleteKeyRPC(node1 *NodeRPC, key string) error {
	return node.connections.DeleteKey(node1, key)
}

// Delete multiple keys
func (node *Node) deleteKeysRPC(node1 *NodeRPC, keys []string) error {
	return node.connections.DeleteKeys(node1, keys)
}
