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

func Dial(addr string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	return grpc.Dial(addr, opts...)
}

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
	poolMtx sync.RWMutex

	server *grpc.Server

	shutdown int32
}

type grpcConn struct {
	addr       string
	client     ChordClient
	conn       *grpc.ClientConn
	lastActive time.Time
}

func NewGrpcConnection(config *Config) (*GrpcConnection, error) {

	addr := config.Addr
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	pool := make(map[string]*grpcConn)

	grp := &GrpcConnection{
		sock:    listener.(*net.TCPListener),
		timeout: config.Timeout,
		maxIdle: config.MaxIdle,
		pool:    pool,
		config:  config,
	}

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

	g.poolMtx.RLock()

	if atomic.LoadInt32(&g.shutdown) == 1 {
		g.poolMtx.Unlock()
		return nil, fmt.Errorf("TCP transport is shutdown")
	}

	cc, ok := g.pool[addr]
	g.poolMtx.RUnlock()
	if ok {
		return cc.client, nil
	}

	var conn *grpc.ClientConn
	var err error
	conn, err = Dial(addr, g.config.DialOpts...)
	if err != nil {
		return nil, err
	}

	client := NewChordClient(conn)
	cc = &grpcConn{addr, client, conn, time.Now()}
	g.poolMtx.Lock()
	if g.pool == nil {
		g.poolMtx.Unlock()
		return nil, errors.New("must instantiate node before using")
	}
	g.pool[addr] = cc
	g.poolMtx.Unlock()

	return client, nil
}

func (g *GrpcConnection) Start() error {
	go g.listen()
	go g.reapOld()
	return nil
}

func (g *GrpcConnection) Stop() error {
	atomic.StoreInt32(&g.shutdown, 1)
	g.poolMtx.Lock()
	g.server.Stop()
	for _, conn := range g.pool {
		conn.Close()
	}
	g.pool = nil
	g.poolMtx.Unlock()
	return nil
}

func (g *GrpcConnection) returnConn(o *grpcConn) {
	o.lastActive = time.Now()
	g.poolMtx.Lock()
	defer g.poolMtx.Unlock()
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
	g.poolMtx.Lock()
	defer g.poolMtx.Unlock()
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
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	return client.GetNextNode(ctx, emptyRequest)
}

func (g *GrpcConnection) GetNextNodeById(node *Node, id []byte) (*Node, error) {
	client, err := g.getConn(node.Addr)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	return client.GetNextNodeById(ctx, &ID{Id: id})
}

func (g *GrpcConnection) GetPreNode(node *Node) (*Node, error) {
	client, err := g.getConn(node.Addr)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	return client.GetPreNode(ctx, emptyRequest)
}

func (g *GrpcConnection) SetPreNode(node *Node, pred *Node) error {
	client, err := g.getConn(node.Addr)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	_, err = client.SetPreNode(ctx, pred)
	return err
}

func (g *GrpcConnection) SetNextNode(node *Node, succ *Node) error {
	client, err := g.getConn(node.Addr)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	_, err = client.SetNextNode(ctx, succ)
	return err
}

func (g *GrpcConnection) CheckPreNode(node *Node) error {
	client, err := g.getConn(node.Addr)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	_, err = client.CheckPreNodeById(ctx, &ID{Id: node.Id})
	return err
}

func (g *GrpcConnection) Inform(node, pred *Node) error {
	client, err := g.getConn(node.Addr)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	_, err = client.Inform(ctx, pred)
	return err

}

func (g *GrpcConnection) AddKey(node *Node, key, value string) error {
	client, err := g.getConn(node.Addr)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	_, err = client.AddKey(ctx, &SetRequest{Key: key, Value: value})
	return err
}

func (g *GrpcConnection) GetValue(node *Node, key string) (*GetResponse, error) {
	client, err := g.getConn(node.Addr)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	return client.GetValue(ctx, &GetRequest{Key: key})
}

func (g *GrpcConnection) DeleteKey(node *Node, key string) error {
	client, err := g.getConn(node.Addr)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	_, err = client.DeleteKey(ctx, &DeleteRequest{Key: key})
	return err
}

func (g *GrpcConnection) DeleteKeys(node *Node, keys []string) error {
	client, err := g.getConn(node.Addr)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	_, err = client.DeleteKeys(
		ctx, &MultiDeleteRequest{Keys: keys},
	)
	return err
}

func (g *GrpcConnection) GetKeys(node *Node, from, to []byte) ([]*KV, error) {
	client, err := g.getConn(node.Addr)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	val, err := client.GetKeys(
		ctx, &GetKeysRequest{From: from, To: to},
	)
	if err != nil {
		return nil, err
	}
	return val.Values, nil
}
