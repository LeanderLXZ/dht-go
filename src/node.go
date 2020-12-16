package src

import (
	"crypto/sha1"
	"hash"
	"sync"
	"time"

	"google.golang.org/grpc"
)

// Structure for parameters
type Parameters struct {
	NodeID        string
	Address       string
	HashFunc      func() hash.Hash // Hash function
	HashLen       int              // Length of hash
	Timeout       time.Duration    // Timeout
	MaxIdleTime   time.Duration    // Maximum idle time
	MinStableTime time.Duration    // Minimum stable time
	MaxStableTime time.Duration    // Maximum stable time
	ServerOptions []grpc.ServerOption
	DialOptions   []grpc.DialOption
}
// structure for node
type Node struct {
	*models.Node
	para 			*Parameters

	predecessor 	*models.Node
	predLock		sync.RWMutex

	successor		*models.Node
	succLock		sync.RWMutex

	fingerTable 	fingerTable
	fingerLock		sync.RWMutex

	storage 		Storage
	stLock			sync.RWMutex

	transport 		Transport
	tsLock     		sync.RWMutex

	lastStablized 	time.Time
}
// Get a default parameters settings
func GetDefaultParameters() *Parameters {
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

func(n *Node)join(newNode *model.Node) error {

}
