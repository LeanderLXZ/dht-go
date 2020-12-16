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
	emptyNode                = &models.Node{}
	emptyRequest             = &models.ER{}
	emptyGetResponse         = &models.GetResponse{}
	emptySetResponse         = &models.SetResponse{}
	emptyDeleteResponse      = &models.DeleteResponse{}
	emptyRequestKeysResponse = &models.RequestKeysResponse{}
)

func Dial(addr string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	return grpc.Dial(addr, opts...)
}

type Connections interface {
	Start() error
	Stop() error

	GetNextNode(*models.Node) (*models.Node, error)
	FindNextNode(*models.Node, []byte) (*models.Node, error)
	GetPreNode(*models.Node) (*models.Node, error)
	Inform(*models.Node, *models.Node) error
	CheckPreNode(*models.Node) error
	SetPreNode(*models.Node, *models.Node) error
	SetNextNode(*models.Node, *models.Node) error

	GetValue(*models.Node, string) (*models.GetResponse, error)
	AddKey(*models.Node, string, string) error
	DeleteKey(*models.Node, string) error
	RequestKeys(*models.Node, []byte, []byte) ([]*models.KV, error)
	DeleteKeys(*models.Node, []string) error
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
	client     models.ChordClient
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
	models.RegisterChordServer(g.server, node)
}

func (g *GrpcConnection) GetServer() *grpc.Server {
	return g.server
}

func (g *GrpcConnection) getConn(addr string,) (models.ChordClient, error) {

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

	client := models.NewChordClient(conn)
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

func (g *GrpcConnection) GetNextNode(node *models.Node) (*models.Node, error) {
	client, err := g.getConn(node.Addr)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	return client.GetNextNode(ctx, emptyRequest)
}

func (g *GrpcConnection) FindNextNode(node *models.Node, id []byte) (*models.Node, error) {
	client, err := g.getConn(node.Addr)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	return client.FindNextNode(ctx, &models.ID{Id: id})
}

func (g *GrpcConnection) GetPreNode(node *models.Node) (*models.Node, error) {
	client, err := g.getConn(node.Addr)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	return client.GetPreNode(ctx, emptyRequest)
}

func (g *GrpcConnection) SetPreNode(node *models.Node, pred *models.Node) error {
	client, err := g.getConn(node.Addr)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	_, err = client.SetPreNode(ctx, pred)
	return err
}

func (g *GrpcConnection) SetNextNode(node *models.Node, succ *models.Node) error {
	client, err := g.getConn(node.Addr)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	_, err = client.SetNextNode(ctx, succ)
	return err
}

func (g *GrpcConnection) CheckPreNode(node *models.Node) error {
	client, err := g.getConn(node.Addr)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	_, err = client.CheckPreNode(ctx, &models.ID{Id: node.Id})
	return err
}

func (g *GrpcConnection) Inform(node, pred *models.Node) error {
	client, err := g.getConn(node.Addr)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	_, err = client.Inform(ctx, pred)
	return err

}

func (g *GrpcConnection) AddKey(node *models.Node, key, value string) error {
	client, err := g.getConn(node.Addr)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	_, err = client.XSet(ctx, &models.SetRequest{Key: key, Value: value})
	return err
}

func (g *GrpcConnection) GetValue(node *models.Node, key string) (*models.GetResponse, error) {
	client, err := g.getConn(node.Addr)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	return client.XGet(ctx, &models.GetRequest{Key: key})
}

func (g *GrpcConnection) DeleteKey(node *models.Node, key string) error {
	client, err := g.getConn(node.Addr)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	_, err = client.XDelete(ctx, &models.DeleteRequest{Key: key})
	return err
}

func (g *GrpcConnection) DeleteKeys(node *models.Node, keys []string) error {
	client, err := g.getConn(node.Addr)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	_, err = client.XMultiDelete(
		ctx, &models.MultiDeleteRequest{Keys: keys},
	)
	return err
}

func (g *GrpcConnection) RequestKeys(node *models.Node, from, to []byte) ([]*models.KV, error) {
	client, err := g.getConn(node.Addr)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	val, err := client.XRequestKeys(
		ctx, &models.RequestKeysRequest{From: from, To: to},
	)
	if err != nil {
		return nil, err
	}
	return val.Values, nil
}
