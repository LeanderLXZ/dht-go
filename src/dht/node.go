package dht

import (
	"crypto/sha1"
	"fmt"
	"hash"
	"math/big"
	"sync"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// Structure for parameters
type Parameters struct {
	NodeId        string
	Address       string
	HashFunc      func() hash.Hash    // Hash function
	HashLen       int                 // Length of hash
	Timeout       time.Duration       // Timeout
	MaxIdleTime   time.Duration       // Maximum idle time
	MinStableTime time.Duration       // Minimum stable time
	MaxStableTime time.Duration       // Maximum stable time
	ServerOptions []grpc.ServerOption // grpc server option
	DialOptions   []grpc.DialOption   // grpc dial option
}

// hashsize bigger than hash func size
func (para *Parameters) Verify() error {
	return nil
}

// Get a initial parameters settings
func GetInitialParameters() *Parameters {
	para := &Parameters{}
	para.HashFunc = sha1.New
	para.HashLen = para.HashFunc().Size() * 8
	para.DialOptions = make([]grpc.DialOption, 0, 5)
	para.DialOptions = append(
		para.DialOptions,
		grpc.WithBlock(),
		grpc.WithTimeout(5*time.Second),
		grpc.FailOnNonTempDialError(true),
		grpc.WithInsecure(),
	)
	return para
}

// join a new node
func (node *Node) join(newNode *NodeRPC) error {
	// First check if node already present in the circle
	// Join this node to the same chord ring as parent
	var a *NodeRPC
	fmt.Println("----------------------------------------------------------------")
	fmt.Println("[Log] Added a new node!")
	fmt.Printf("My Node ID:\n\t%d\n", (&big.Int{}).SetBytes(node.NodeId))
	// // Ask if our id already exists on the ring.
	if newNode != nil {
		rtNode, err := node.getNextNodeByIdRPC(newNode, node.NodeId)
		if err != nil {
			return err
		}

		if isEqual(rtNode.NodeId, node.NodeId) {
			return ERR_NODE_EXISTS
		}
		a = newNode
	} else {
		a = node.NodeRPC
	}

	succNode, err := node.getNextNodeByIdRPC(a, node.NodeId)
	if err != nil {
		return err
	}
	node.succLock.Lock()
	node.successor = succNode
	node.succLock.Unlock()

	return nil
}

// Structure of Node
type Node struct {
	*NodeRPC
	para *Parameters

	closeCh chan struct{}

	predecessor *NodeRPC
	predLock    sync.RWMutex

	successor *NodeRPC
	succLock  sync.RWMutex

	fingerTable fingerTable
	fingerLock  sync.RWMutex

	dataStorage hash_table
	stLock      sync.RWMutex

	connections Connections
	tsLock      sync.RWMutex

	lastStablized time.Time
}

// ================================================
//                  Local Methods
// ================================================

// ---------------- Node Operations ---------------

// node period to stablize, check the finger table, and check predecessor status
func newNodePeriod(node *Node) {
	go func() {
		ticker := time.NewTicker(1000 * time.Millisecond)
		for {
			select {
			case <-ticker.C:
				node.stabilize()
			case <-node.closeCh:
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
				next = node.findNextFinger(next)
			case <-node.closeCh:
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
				node.CheckPreNode()
			case <-node.closeCh:
				ticker.Stop()
				return
			}
		}
	}()
}

func (node *Node) stabilize() {
	node.succLock.RLock()
	succNode := node.successor
	if succNode == nil {
		node.succLock.RLock()
		return
	}
	node.succLock.RUnlock()

	n, err := node.getPreNodeRPC(succNode)
	if err != nil || n == nil {
		fmt.Println("error getting predecessor, ", err, n)
		return
	}

	if n.NodeId != nil && between(n.NodeId, node.NodeId, succNode.NodeId) {
		node.succLock.Lock()
		node.successor = n
		node.succLock.Unlock()
		fmt.Println("----------------------------------------------------------------")
		fmt.Printf(
			"[Log] Set successor\n\tof: %d\n\tto be: %d\n",
			(&big.Int{}).SetBytes(node.NodeId),
			(&big.Int{}).SetBytes(n.NodeId),
		)
	}
	node.informRPC(succNode, node.NodeRPC)
}

func (node *Node) CheckPreNode() {
	// implement using rpc func
	node.predLock.RLock()
	predNode := node.predecessor
	node.predLock.RUnlock()

	if predNode != nil {
		err := node.connections.CheckPreNode(predNode)
		if err != nil {
			fmt.Println("predecessor failed!", err)
			node.predLock.Lock()
			node.predecessor = nil
			node.predLock.Unlock()
		}
	}
}

//  -----------------------------------------------
// 					findNextNode
// 			reference from paper fig. 5
// 			ask node n to find the successor of id
//	-----------------------------------------------
func (node *Node) findNextNode(nodeId []byte) (*NodeRPC, error) {
	node.succLock.RLock()
	defer node.succLock.RUnlock()
	currNode := node.NodeRPC
	succNode := node.successor

	// if no succNode exists
	if succNode == nil {
		return currNode, nil
	}

	var err error
	// ask direct predecessor for node with nodeId,
	// if not found, then go to find closest preceding node of nodeId in fingertable.
	if betweenRightIncl(nodeId, currNode.NodeId, succNode.NodeId) {
		return succNode, nil
	} else {
		preNode := node.closestPreNode(nodeId)
		if isEqual(preNode.NodeId, node.NodeId) {
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
		if succNode == nil {
			return currNode, nil
		}
		return succNode, nil
	}
	return nil, nil
}

//  -----------------------------------------------
// 				cloest predecessor node
// search the local table for the highest predecessor of id
// in fingerTable.
//	-----------------------------------------------
func (node *Node) closestPreNode(nodeId []byte) *NodeRPC {
	node.predLock.RLock()
	defer node.predLock.RUnlock()

	currNode := node.NodeRPC

	for i := len(node.fingerTable) - 1; i >= 0; i-- {
		v := node.fingerTable[i]
		if v == nil || v.RemoteNode == nil {
			continue
		}
		if between(v.InitId, currNode.NodeId, nodeId) {
			return v.RemoteNode
		}
	}
	return currNode
}

// ---------------- Key Operations ----------------

// generate hash key
// input ip address, output hash key (sha1)
func (node *Node) getHashKey(key string) ([]byte, error) {
	hash := node.para.HashFunc()
	if _, err := hash.Write([]byte(key)); err != nil {
		return nil, err
	}
	hashKey := hash.Sum(nil)
	return hashKey, nil
}

// get the location of a given key
func (node *Node) getLocation(key string) (*NodeRPC, error) {
	nodeId, err := node.getHashKey(key)
	if err != nil {
		return nil, err
	}
	nextNode, err := node.findNextNode(nodeId)
	value, _ := node.getValueRPC(nextNode, key)
	if value != nil {
		fmt.Println("----------------------------------------------------------------")
		fmt.Printf(
			"[Command] Get the location of a key: %s\n\tHashKey: %d\n\tLocation: %d\n",
			key,
			(&big.Int{}).SetBytes(nodeId),
			(&big.Int{}).SetBytes(nextNode.NodeId),
		)
		return nextNode, err
	} else {
		fmt.Println("----------------------------------------------------------------")
		fmt.Printf(
			"[Command] Get the location of a key: %s\n\tHashKey: %d\n\tLocation: %s\n",
			key,
			(&big.Int{}).SetBytes(nodeId),
			"Key not found",
		)
		return nil, nil
	}
}

// Add a (key, value) pair
func (node *Node) addKey(key, value string) error {
	nodeId, err := node.getHashKey(key)
	if err != nil {
		return err
	}
	node1, err := node.findNextNode(nodeId)
	if err != nil {
		return err
	}
	err = node.addKeyRPC(node1, key, value)
	fmt.Println("----------------------------------------------------------------")
	hk, errhk := node.getHashKey(key)
	if errhk != nil {
		return errhk
	} else {
		fmt.Printf(
			"[Command] Add a new key:\n\t(%s, \"%s\")\n\tHashKey: %d\n\tFrom: %d\n\tTo: %d\n",
			key,
			value,
			(&big.Int{}).SetBytes(hk),
			(&big.Int{}).SetBytes(node.NodeId),
			(&big.Int{}).SetBytes(node1.NodeId),
		)
	}
	return err
}

// get the value of a given key
func (node *Node) getValue(key string) ([]byte, error) {
	nodeId, err := node.getHashKey(key)
	if err != nil {
		return nil, err
	}
	node1, err := node.findNextNode(nodeId)
	if err != nil {
		return nil, err
	}
	value, err := node.getValueRPC(node1, key)
	var v string
	if value != nil {
		v = string(value.Value)
	} else {
		v = "Key not found"
	}
	fmt.Println("----------------------------------------------------------------")
	hk, errhk := node.getHashKey(key)
	if errhk != nil {
		return nil, errhk
	} else {
		fmt.Printf(
			"[Command] Get the value of a key: %s\n\tHashKey: %d\n\tFrom: %d\n\tValue: %s\n",
			key,
			(&big.Int{}).SetBytes(hk),
			(&big.Int{}).SetBytes(node1.NodeId),
			v,
		)
	}
	if value != nil {
		return value.Value, nil
	} else {
		return nil, nil
	}
}

// Delete a given key
func (node *Node) deleteKey(key string) error {
	nodeId, err := node.getHashKey(key)
	if err != nil {
		return err
	}
	node1, err := node.findNextNode(nodeId)
	if err != nil {
		return err
	}
	hk, errhk := node.getHashKey(key)
	if errhk != nil {
		return errhk
	} else {
		fmt.Printf(
			"[Command] Delete a key: %s\n\tHashKey: %d\n\tFrom: %d\n",
			key,
			(&big.Int{}).SetBytes(hk),
			(&big.Int{}).SetBytes(node1.NodeId),
		)
	}

	value, err := node.getValueRPC(node1, key)
	if value != nil {
		fmt.Printf("\tState: Succeed\n")
	} else {
		fmt.Printf("\tState: Failed - Key not found\n")
	}
	err = node.deleteKeyRPC(node1, key)
	return err
}

// When a new node is added to the ring, it gets the keys from the next node
func (node *Node) getKeys(preNode, nextNode *NodeRPC) ([]*KeyValuePair, error) {
	if isEqual(node.NodeId, nextNode.NodeId) {
		return nil, nil
	}
	return node.getKeysRPC(
		nextNode, preNode.NodeId, node.NodeId,
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
		node.dataStorage.AddKey(kv.Key, kv.Value)
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
//  -----------------------------------------------
// 					create a new node
// 	create a new node and peridoically stablize it
// 			return error if node already exists
//	-----------------------------------------------
func CreateNode(para *Parameters, newNode *NodeRPC) (*Node, error) {

	fmt.Println("----------------------------------------------------------------")
	fmt.Printf("[Log] Created a new node\n\tAddress: %s\n\tInitial Node ID: %s\n", para.Address, para.NodeId)

	if err := para.Verify(); err != nil {
		return nil, err
	}

	node := &Node{
		NodeRPC:     new(NodeRPC),
		closeCh:     make(chan struct{}),
		para:        para,
		dataStorage: NewDataHash(para.HashFunc),
	}

	var nodeId string
	if para.NodeId != "" {
		nodeId = para.NodeId
	} else {
		nodeId = para.Address
	}
	hashId, err := node.getHashKey(nodeId)
	if err != nil {
		return nil, err
	}
	fmt.Printf("\tNew Node ID: %d\n", (&big.Int{}).SetBytes(hashId))

	node.NodeRPC.NodeId = hashId
	node.NodeRPC.Address = para.Address

	// create fingertable for new node
	fmt.Println("[Log] Created finger table for the node.")
	node.fingerTable = newFingerTable(node.NodeRPC, para.HashLen)

	// start RPC
	Connections, err := NewGrpcConnection(para)
	if err != nil {
		return nil, err
	}
	node.connections = Connections

	RegisterDistributedHashTableServer(Connections.server, node)

	node.connections.Start()

	err = node.join(newNode)
	if err != nil {
		return nil, err
	}

	newNodePeriod(node)

	return node, nil
}

func CreateNodeById(id string, addr string) *NodeRPC {
	h := sha1.New()
	if _, err := h.Write([]byte(id)); err != nil {
		return nil
	}
	hk := h.Sum(nil)

	return &NodeRPC{
		NodeId:  hk,
		Address: addr,
	}
}

// get the predecessor node and return it
func (node *Node) GetPreNode(ctx context.Context, r *EmptyRequest) (*NodeRPC, error) {
	node.predLock.RLock()
	preNode := node.predecessor
	node.predLock.RUnlock()
	if preNode == nil {
		return emptyNode, nil
	}
	return preNode, nil
}

// set the predecessor node
func (node *Node) SetPreNode(ctx context.Context, preNode *NodeRPC) (*EmptyRequest, error) {
	node.predLock.Lock()
	node.predecessor = preNode
	node.predLock.Unlock()
	return emptyRequest, nil
}

// get the successor node and return it
func (node *Node) GetNextNode(ctx context.Context, r *EmptyRequest) (*NodeRPC, error) {
	node.succLock.RLock()
	NextNode := node.successor
	node.succLock.RUnlock()
	if NextNode == nil {
		return emptyNode, nil
	}
	return NextNode, nil
}

// set the successor node
func (node *Node) SetNextNode(ctx context.Context, NextNode *NodeRPC) (*EmptyRequest, error) {
	node.succLock.Lock()
	node.successor = NextNode
	node.succLock.Unlock()
	return emptyRequest, nil
}

func (node *Node) CheckPreNodeById(ctx context.Context, preNodeId *NodeIdRPC) (*EmptyRequest, error) {
	return emptyRequest, nil
}

func (node *Node) GetNextNodeById(ctx context.Context, nodeId *NodeIdRPC) (*NodeRPC, error) {
	succNode, err := node.findNextNode(nodeId.NodeId)
	if err != nil {
		return nil, err
	}

	if succNode == nil {
		return nil, ERR_NO_NEXT_NODE
	}

	return succNode, nil

}

func (node *Node) Inform(ctx context.Context, n *NodeRPC) (*EmptyRequest, error) {
	node.predLock.Lock()
	defer node.predLock.Unlock()
	var prevNode *NodeRPC
	predNode := node.predecessor
	if predNode == nil || between(n.NodeId, predNode.NodeId, node.NodeId) {
		fmt.Println("----------------------------------------------------------------")
		fmt.Printf(
			"[Log] Set predecessor\n\tof: %d\n\tto be: %d\n",
			(&big.Int{}).SetBytes(n.NodeId),
			(&big.Int{}).SetBytes(node.NodeId),
		)
		if node.predecessor != nil {
			prevNode = node.predecessor
		}
		node.predecessor = n

		// transfer keys from parent node
		if prevNode != nil {
			if between(node.predecessor.NodeId, prevNode.NodeId, node.NodeId) {
				node.changeKeys(predNode, node.predecessor)
			}
		}

	}

	return emptyRequest, nil
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

// Get finger table
func (node *Node) GetFingerTable() []*fingerEntry {
	return node.fingerTable
}

// Get NodeId
func (node *Node) GetNodeId() []byte {
	return node.NodeId
}

// ---------------- Rewrite RPC -------------------

func (node *Node) GetValueHT(ctx context.Context, req *GetValueReq) (*GetValueResp, error) {
	node.stLock.RLock()
	defer node.stLock.RUnlock()
	val, err := node.dataStorage.GetValue(req.Key)
	if err != nil {
		return emptyGetValueResp, err
	}
	return &GetValueResp{Value: val}, nil
}

func (node *Node) AddKeyHT(ctx context.Context, req *AddKeyReq) (*AddKeyResp, error) {
	node.stLock.RLock()
	defer node.stLock.RUnlock()
	err := node.dataStorage.AddKey(req.Key, req.Value)
	fmt.Println("----------------------------------------------------------------")
	fmt.Printf("[Log] Received a new (key, value) pair:\n\t(%s, \"%s\")\n", req.Key, req.Value)
	fmt.Printf("My Node ID:\n\t%d\n", (&big.Int{}).SetBytes(node.NodeId))
	// fmt.Printf("My successor:\n\t%d\n", (&big.Int{}).SetBytes(node.successor.NodeId))
	return emptyAddKeyResp, err
}

func (node *Node) GetKeysHT(ctx context.Context, req *GetKeysReq) (*GetKeysResp, error) {
	node.stLock.RLock()
	defer node.stLock.RUnlock()
	val, err := node.dataStorage.Between(req.Start, req.End)
	if err != nil {
		return emptyGetKeysResp, err
	}
	return &GetKeysResp{Kvs: val}, nil
}

func (node *Node) DeleteKeyHT(ctx context.Context, req *DeleteKeyReq) (*DeleteKeyResp, error) {
	node.stLock.RLock()
	defer node.stLock.RUnlock()
	err := node.dataStorage.DeleteKey(req.Key)
	return emptyDeleteKeyResp, err
}

func (node *Node) DeleteKeysHT(ctx context.Context, req *DeleteKeysReq) (*DeleteKeysResp, error) {
	node.stLock.RLock()
	defer node.stLock.RUnlock()
	err := node.dataStorage.DeleteKeys(req.Keys...)
	return emptyDeleteKeysResp, err
}

// ================================================
//                      Stop
// ================================================

func (node *Node) Stop() {
	close(node.closeCh)

	node.succLock.RLock()
	nextNode := node.successor
	node.succLock.RUnlock()

	node.predLock.RLock()
	preNode := node.predecessor
	node.predLock.RUnlock()

	if node.NodeRPC.Address != nextNode.Address && preNode != nil {
		node.changeKeys(preNode, nextNode)
		predErr := node.setPreNodeRPC(nextNode, preNode)
		succErr := node.setNextNodeRPC(preNode, nextNode)
		fmt.Println("Errors occurred and stop: ", predErr, succErr)
	}

	node.connections.Stop()
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
func (node *Node) getNextNodeRPC(node1 *NodeRPC) (*NodeRPC, error) {
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
func (node *Node) informRPC(node1, preNode *NodeRPC) error {
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
func (node *Node) getKeysRPC(node1 *NodeRPC, start []byte, end []byte) ([]*KeyValuePair, error) {
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
