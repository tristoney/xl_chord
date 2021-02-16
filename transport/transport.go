package transport

import (
	"context"
	"github.com/mitchellh/mapstructure"
	"github.com/tristoney/xl_chord/dto"
	"github.com/tristoney/xl_chord/node"
	"github.com/tristoney/xl_chord/proto"
	"github.com/tristoney/xl_chord/util"
	"github.com/tristoney/xl_chord/util/chorderr"
	"google.golang.org/grpc"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// Transport is the component of communication
// Which allows nodes to contact with each other
type Transport interface {
	Start() error // Start start the communication of current node
	Stop() error  // Stop shut the communication of current node

	// interactive communication
	GetPredecessor(*dto.Node) (*dto.Node, error)
	GetSuccessor(*dto.Node) (*dto.Node, error)
	Notify(*dto.Node, *dto.Node) error
	FindSuccessor(*dto.Node, []byte) (*dto.Node, error)
	CheckPredecessor(*dto.Node) error
	SetPredecessor(*dto.Node, *dto.Node) error
	SetSuccessor(*dto.Node, *dto.Node) error

	// storage access
	GetVal(*dto.Node, string) ([]byte, error)
	SetKey(*dto.Node, *dto.Pair) ([]byte, error)
	DeleteKey(*dto.Node, string) (string, []byte, error)
	MultiDelete(*dto.Node, []string) ([]string, error)
	GetKeys(*dto.Node, []byte, []byte) ([]*dto.Pair, error)
}

type GrpcConn struct {
	addr       string            // ip:port
	client     proto.ChordClient // chord client
	conn       *grpc.ClientConn  // client connection
	lastActive time.Time         // last active time
}

// transport via gRPC
type GrpcTransport struct {
	config *node.Config

	timeout time.Duration
	sock    *net.TCPListener

	pool    map[string]*GrpcConn // connection pool
	poolMtx sync.RWMutex         // mutex to avoid concurrently modify

	server   *grpc.Server
	shutDown int32
}

func NewGrpcTransport(config *node.Config) (*GrpcTransport, error) {
	if config == nil {
		return nil, chorderr.ErrInvalidConfig
	}
	// start the listener
	listener, err := net.Listen("tcp", config.Addr)
	if err != nil {
		return nil, err
	}

	return &GrpcTransport{
		config:   config,
		timeout:  config.Timeout,
		sock:     listener.(*net.TCPListener), // use tcp listener
		pool:     make(map[string]*GrpcConn, 0),
		server:   grpc.NewServer(config.ServerOptions...),
		shutDown: 0,
	}, nil
}

func (g *GrpcTransport) registerNode(node *node.Node) {
	proto.RegisterChordServer(g.server, node)
}

func (g *GrpcTransport) GetServer() *grpc.Server {
	return g.server
}

// getConnection gets a connection instance through the connection pool
// if pool contains the connection of addr, return the client
// else create a new connection of addr, store in the pool
func (g *GrpcTransport) getConnection(addr string) (proto.ChordClient, error) {
	g.poolMtx.RLock()

	if atomic.LoadInt32(&g.shutDown) == 1 {
		return nil, chorderr.ErrTransportShutdown
	}

	grpcConn, ok := g.pool[addr]
	g.poolMtx.RUnlock()
	if ok {
		return grpcConn.client, nil
	}

	conn, err := grpc.Dial(addr, g.config.DialOptions...)
	if err != nil {
		return nil, err
	}

	client := proto.NewChordClient(conn)
	grpcConn = &GrpcConn{
		addr:       addr,
		client:     client,
		conn:       conn,
		lastActive: time.Now(),
	}

	g.poolMtx.Lock()
	if g.pool == nil {
		return nil, chorderr.ErrNilPool
	}
	g.pool[addr] = grpcConn
	g.poolMtx.Unlock()
	return client, nil
}

func (g *GrpcConn) Close() error {
	return g.conn.Close()
}

// Listens for inbound connections
func (g *GrpcTransport) listen() error {
	return g.server.Serve(g.sock)
}

func (g *GrpcTransport) Start() error {
	errChan := make(chan error, 1)
	// start the RPC server
	go func() {
		err := g.listen()
		if err != nil {
			errChan <- err
		}
	}()
	close(errChan)
	if len(errChan) > 0 {
		return <-errChan
	}
	return nil
}

func (g *GrpcTransport) Stop() error {
	g.poolMtx.Lock()
	g.server.Stop()
	for _, conn := range g.pool {
		_ = conn.Close()
	}
	g.pool = nil
	g.poolMtx.Unlock()
	return nil
}

func (g *GrpcTransport) GetPredecessor(node *dto.Node) (*dto.Node, error) {
	client, err := g.getConnection(node.Addr)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	resp, err := client.GetPredecessor(ctx, util.NewBaseReq())
	if err != nil {
		return nil, err
	}
	n := resp.GetNode()
	return &dto.Node{
		ID:   n.GetId(),
		Addr: n.GetAddr(),
	}, nil
}

func (g *GrpcTransport) GetSuccessor(node *dto.Node) (*dto.Node, error) {
	client, err := g.getConnection(node.Addr)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	resp, err := client.GetPredecessor(ctx, util.NewBaseReq())
	if err != nil {
		return nil, err
	}
	n := resp.GetNode()
	return &dto.Node{
		ID:   n.GetId(),
		Addr: n.GetAddr(),
	}, nil
}

func (g *GrpcTransport) Notify(node *dto.Node, pred *dto.Node) error {
	client, err := g.getConnection(node.Addr)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	_, err = client.Notify(ctx, &proto.NodeReq{
		Node: &proto.Node{
			Id:   pred.ID,
			Addr: pred.Addr,
		},
		Base: util.NewBaseReq(),
	})
	return err
}

func (g *GrpcTransport) FindSuccessor(node *dto.Node, id []byte) (*dto.Node, error) {
	client, err := g.getConnection(node.Addr)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	resp, err := client.FindSuccessor(ctx, &proto.IDReq{
		Id:   id,
		Base: util.NewBaseReq(),
	})
	if err != nil {
		return nil, err
	}
	n := resp.GetNode()
	return &dto.Node{
		ID:   n.Id,
		Addr: n.GetAddr(),
	}, nil
}

func (g *GrpcTransport) CheckPredecessor(node *dto.Node) error {
	client, err := g.getConnection(node.Addr)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	_, err = client.ChekPredecessor(ctx, &proto.IDReq{
		Id:   node.ID,
		Base: util.NewBaseReq(),
	})
	return err
}

func (g *GrpcTransport) SetPredecessor(node *dto.Node, pred *dto.Node) error {
	client, err := g.getConnection(node.Addr)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	_, err = client.SetPredecessor(ctx, &proto.NodeReq{
		Node: &proto.Node{
			Id:   pred.ID,
			Addr: pred.Addr,
		},
		Base: util.NewBaseReq(),
	})
	return err
}

func (g *GrpcTransport) SetSuccessor(node *dto.Node, successor *dto.Node) error {
	client, err := g.getConnection(node.Addr)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	_, err = client.SetSuccessor(ctx, &proto.NodeReq{
		Node: &proto.Node{
			Id:   successor.ID,
			Addr: successor.Addr,
		},
		Base: util.NewBaseReq(),
	})
	return err
}

func (g *GrpcTransport) GetVal(node *dto.Node, key string) ([]byte, error) {
	client, err := g.getConnection(node.Addr)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	resp, err := client.GetVal(ctx, &proto.GetValReq{
		Key:  key,
		Base: util.NewBaseReq(),
	})
	if err != nil {
		return nil, err
	}
	return resp.GetValue(), nil
}

func (g *GrpcTransport) SetKey(node *dto.Node, pair *dto.Pair) ([]byte, error) {
	client, err := g.getConnection(node.Addr)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	var protoPair proto.Pair
	_ = mapstructure.Decode(pair, &protoPair)
	resp, err := client.SetKey(ctx, &proto.SetKeyReq{
		Pair: &protoPair,
		Base: util.NewBaseReq(),
	})
	if err != nil {
		return nil, err
	}
	return resp.GetId(), nil
}

func (g *GrpcTransport) DeleteKey(node *dto.Node, key string) (string, []byte, error) {
	client, err := g.getConnection(node.Addr)
	if err != nil {
		return "", nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	resp, err := client.DeleteKey(ctx, &proto.DeleteKeyReq{
		Key:  key,
		Base: util.NewBaseReq(),
	})
	if err != nil {
		return "", nil, err
	}
	return key, resp.GetId(), nil
}

func (g *GrpcTransport) MultiDelete(node *dto.Node, keys []string) ([]string, error) {
	client, err := g.getConnection(node.Addr)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	resp, err := client.MultiDelete(ctx, &proto.MultiDeleteReq{
		Keys: keys,
		Base: util.NewBaseReq(),
	})
	if err != nil {
		return nil, err
	}
	return resp.GetKeys(), nil
}

func (g *GrpcTransport) GetKeys(node *dto.Node, from []byte, to []byte) ([]*dto.Pair, error) {
	client, err := g.getConnection(node.Addr)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	resp, err := client.GetKeys(ctx, &proto.GetKeysReq{
		From: from,
		To:   to,
		Base: util.NewBaseReq(),
	})
	if err != nil {
		return nil, err
	}
	pairs := resp.GetPairs()
	dtoPairs := make([]*dto.Pair, 0)
	for _, pair := range pairs {
		var dtoPair dto.Pair
		_ = mapstructure.Decode(pair, &dtoPair)
		dtoPairs = append(dtoPairs, &dtoPair)
	}
	return dtoPairs, nil
}
