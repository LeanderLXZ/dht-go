package src

import (
	"hash"
	"time"

	"google.golang.org/grpc"
)

type Parameters struct {
	NodeID        string
	Address       string
	HashFunc      func() hash.Hash // Hash function
	HashLen       int              // Length of hash
	Timeout       time.Duration    // Timeout
	MaxIdleTime   time.Duration    // Maximum idle time
	MinStableTime time.Duration    // Minimum stable time
	MaxStableTime time.Duration    // Maximum stable time
	ServerOption  []grpc.ServerOption
	DialOption    []grpc.DialOption
}
