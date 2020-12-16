package src

import (
	"errors"
	"fmt"
	"net"
	sync "sync"
	"sync/atomic"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
	emptyNode            = &Node{}
	emptyRequest         = &ER{}
	emptyGetResponse     = &GetResponse{}
	emptySetResponse     = &SetResponse{}
	emptyDeleteResponse  = &DeleteResponse{}
	emptyGetKeysResponse = &GetKeysResponse{}
)

type Connections interface {
	Start() error
	Stop() error

	GetNextNode(*Node) (*Node, error)
	GetNextNodeById(*Node, []byte) (*Node, error)
	GetPreNode(*Node) (*Node, error)
	Inform(*Node, *Node) error
	CheckPreNode(*Node) error
	SetPreNode(*Node, *Node) error
	SetNextNode(*Node, *Node) error

	GetValue(*Node, string) (*GetResponse, error)
	AddKey(*Node, string, string) error
	DeleteKey(*Node, string) error
	GetKeys(*Node, []byte, []byte) ([]*KV, error)
	DeleteKeys(*Node, []string) error
}

type GrpcConnection struct {
	config *Config

	timeout time.Duration
	maxIdle time.Duration

	sock *net.TCPListener

	pool    map[string]*grpcConn
	poolRWM sync.RWMutex

	server *grpc.Server

	shutdown int32
}

type grpcConn struct {
	addr       string
	client     ChordClient
	conn       *grpc.ClientConn
	lastActive time.Time
}

func Dial(addr string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	return grpc.Dial(addr, opts...)
}

func NewGrpcConnection(config *Config) (*GrpcConnection, error) {

	address := config.Addr
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}
	grp := &GrpcConnection{}
	grp.sock = listener.(*net.TCPListener)
	grp.timeout = config.Timeout
	grp.maxIdle = config.MaxIdle
	grp.pool = make(map[string]*grpcConn)
	grp.config = config
	grp.server = grpc.NewServer(config.ServerOpts...)

	return grp, nil
}

func (g *grpcConn) Close() {
	g.conn.Close()
}

func (g *GrpcConnection) registerNode(node *Node) {
	RegisterChordServer(g.server, node)
}

func (g *GrpcConnection) GetServer() *grpc.Server {
	return g.server
}

func (g *GrpcConnection) getConn(addr string) (ChordClient, error) {

	g.poolRWM.RLock()

	if atomic.LoadInt32(&g.shutdown) == 1 {
		g.poolRWM.Unlock()
		return nil, fmt.Errorf("TCP connection is shutdown")
	}

	gc, res := g.pool[addr]
	g.poolRWM.RUnlock()
	if res {
		return gc.client, nil
	}

	var conn *grpc.ClientConn
	var err error
	conn, err = Dial(addr, g.config.DialOpts...)
	if err != nil {
		return nil, err
	}

	client := NewChordClient(conn)
	gc = &grpcConn{addr, client, conn, time.Now()}
	g.poolRWM.Lock()
	if g.pool == nil {
		g.poolRWM.Unlock()
		return nil, errors.New("must instantiate node before using")
	}
	g.pool[addr] = gc
	g.poolRWM.Unlock()

	return client, nil
}

func (g *GrpcConnection) Start() error {
	go g.listen()
	go g.reapOld()
	return nil
}

func (g *GrpcConnection) Stop() error {
	atomic.StoreInt32(&g.shutdown, 1)
	g.poolRWM.Lock()
	g.server.Stop()
	for _, conn := range g.pool {
		conn.Close()
	}
	g.pool = nil
	g.poolRWM.Unlock()
	return nil
}

func (g *GrpcConnection) returnConn(o *grpcConn) {
	o.lastActive = time.Now()
	g.poolRWM.Lock()
	defer g.poolRWM.Unlock()
	if atomic.LoadInt32(&g.shutdown) == 1 {
		o.conn.Close()
		return
	}
	g.pool[o.addr] = o
}

func (g *GrpcConnection) reapOld() {
	ticker := time.NewTicker(60 * time.Second)

	for {
		if atomic.LoadInt32(&g.shutdown) == 1 {
			return
		}
		select {
		case <-ticker.C:
			g.reap()
		}

	}
}

func (g *GrpcConnection) reap() {
	g.poolRWM.Lock()
	defer g.poolRWM.Unlock()
	for host, conn := range g.pool {
		if time.Since(conn.lastActive) > g.maxIdle {
			conn.Close()
			delete(g.pool, host)
		}
	}
}

func (g *GrpcConnection) listen() {
	g.server.Serve(g.sock)
}

func (g *GrpcConnection) GetNextNode(node *Node) (*Node, error) {
	client, err := g.getConn(node.Addr)
	if err == nil {
		state, cancel := context.WithTimeout(context.Background(), g.timeout)
		defer cancel()
		return client.GetNextNode(state, emptyRequest)
	}
	return nil, err
}

func (g *GrpcConnection) GetNextNodeById(node *Node, id []byte) (*Node, error) {
	client, err := g.getConn(node.Addr)
	if err == nil {
		state, cancel := context.WithTimeout(context.Background(), g.timeout)
		defer cancel()
		return client.GetNextNodeById(state, &ID{Id: id})
	}
	return nil, err
}

func (g *GrpcConnection) GetPreNode(node *Node) (*Node, error) {
	client, err := g.getConn(node.Addr)
	if err == nil {
		state, cancel := context.WithTimeout(context.Background(), g.timeout)
		defer cancel()
		return client.GetPreNode(state, emptyRequest)
	}
	return nil, err
}

func (g *GrpcConnection) SetPreNode(node *Node, pred *Node) error {
	client, err := g.getConn(node.Addr)
	if err == nil {
		state, cancel := context.WithTimeout(context.Background(), g.timeout)
		defer cancel()
		_, err = client.SetPreNode(state, pred)
		return err
	}
	return err
}

func (g *GrpcConnection) SetNextNode(node *Node, succ *Node) error {
	client, err := g.getConn(node.Addr)
	if err == nil {
		state, cancel := context.WithTimeout(context.Background(), g.timeout)
		defer cancel()
		_, err = client.SetNextNode(state, succ)
		return err
	}
	return err
}

func (g *GrpcConnection) CheckPreNode(node *Node) error {
	client, err := g.getConn(node.Addr)
	if err == nil {
		state, cancel := context.WithTimeout(context.Background(), g.timeout)
		defer cancel()
		_, err = client.CheckPreNodeById(state, &ID{Id: node.Id})
		return err
	}
	return err
}

func (g *GrpcConnection) Inform(node, pred *Node) error {
	client, err := g.getConn(node.Addr)
	if err == nil {
		state, cancel := context.WithTimeout(context.Background(), g.timeout)
		defer cancel()
		_, err = client.Inform(state, pred)
		return err
	}
	return err
}

func (g *GrpcConnection) AddKey(node *Node, key, value string) error {
	client, err := g.getConn(node.Addr)
	if err == nil {
		state, cancel := context.WithTimeout(context.Background(), g.timeout)
		defer cancel()
		_, err = client.AddKey(state, &SetRequest{Key: key, Value: value})
		return err
	}
	return err
}

func (g *GrpcConnection) GetValue(node *Node, key string) (*GetResponse, error) {
	client, err := g.getConn(node.Addr)
	if err == nil {
		state, cancel := context.WithTimeout(context.Background(), g.timeout)
		defer cancel()
		return client.GetValue(state, &GetRequest{Key: key})
	}
	return nil, err
}

func (g *GrpcConnection) DeleteKey(node *Node, key string) error {
	client, err := g.getConn(node.Addr)
	if err == nil {
		state, cancel := context.WithTimeout(context.Background(), g.timeout)
		defer cancel()
		_, err = client.DeleteKey(state, &DeleteRequest{Key: key})
		return err
	}
	return err
}

func (g *GrpcConnection) DeleteKeys(node *Node, keys []string) error {
	client, err := g.getConn(node.Addr)
	if err == nil {
		state, cancel := context.WithTimeout(context.Background(), g.timeout)
		defer cancel()
		_, err = client.DeleteKeys(
			state, &MultiDeleteRequest{Keys: keys},
		)
		return err
	}
	return err
}

func (g *GrpcConnection) GetKeys(node *Node, from, to []byte) ([]*KV, error) {
	client, err := g.getConn(node.Addr)
	if err != nil {
		return nil, err
	}

	state, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	val, err := client.GetKeys(
		state, &GetKeysRequest{From: from, To: to},
	)
	if err != nil {
		return nil, err
	}
	return val.Values, nil
}
