package transport

import (
	"context"
	"github.com/tristoney/xl_chord/dto"
	"github.com/tristoney/xl_chord/node"
	"github.com/tristoney/xl_chord/proto"
	"github.com/tristoney/xl_chord/util/chorderr"
	"google.golang.org/grpc"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// Transport is the component of communication
// Which allows nodes to contact with each other
type Transport interface {
	Init(*node.Config, interface{}) error // Init the transport
	Start() error                         // Start start the communication of current node
	Stop() error                          // Stop shut the communication of current node

	// interactive communication
	CheckAlive(node *dto.Node) error
	FindSuccessor(node *dto.Node, id []byte) (*dto.Node, error)
	GetPredecessor(node *dto.Node) (*dto.Node, error)
	Notify(node *dto.Node, pred *dto.Node) error
	FindSuccessorFinger(node *dto.Node, index int32, fingerID []byte) (*dto.FindSuccessorFingerResp, error)
	GetSuccessorList(node *dto.Node) ([]*dto.Node, error)

	// storage access
	StoreKey(node *dto.Node, keyID []byte, key, value string) (*dto.StoreKeyResp, error)
	FindKey(node *dto.Node, keyID []byte) (*dto.FindKeyResp, error)
	DeleteKey(node *dto.Node, keyID []byte) (*dto.DeleteKeyResp, error)
	TakeOverKeys(node *dto.Node, data []*dto.Data) error
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

func (g *GrpcTransport) CheckAlive(node *dto.Node) error {
	client, err := g.getConnection(node.Addr)
	if err != nil {
		log.Fatalf("Node {%s} has failed", string(node.ID))
		return chorderr.ErrNodeFailed
	}
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	pong, err := client.CheckAlive(ctx, &proto.CheckAliveReq{})
	if err != nil || pong == nil {
		log.Fatalf("Node {%s} has failed", string(node.ID))
		return chorderr.ErrNodeFailed
	}
	return nil
}

func (g *GrpcTransport) FindSuccessor(node *dto.Node, id []byte) (*dto.Node, error) {
	client, err := g.getConnection(node.Addr)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	resp, err := client.FindSuccessor(ctx, &proto.FindSuccessorReq{
		Id: id,
	})
	if err != nil {
		return nil, err
	}
	n := resp.Successor
	if n == nil {
		return nil, chorderr.ErrNoValidResp
	}
	return &dto.Node{
		ID:   n.Id,
		Addr: n.GetAddr(),
	}, nil

}

func (g *GrpcTransport) GetPredecessor(node *dto.Node) (*dto.Node, error) {
	client, err := g.getConnection(node.Addr)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	resp, err := client.GetPredecessor(ctx, &proto.GetPredecessorReq{})
	if err != nil {
		return nil, err
	}
	n := resp.Predecessor
	if n == nil {
		return nil, chorderr.ErrNoValidResp
	}
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
	_, err = client.Notify(ctx, &proto.NotifyReq{
		Node: &proto.Node{
			Id:   pred.ID,
			Addr: pred.Addr,
		},
	})
	return err
}

func (g *GrpcTransport) FindSuccessorFinger(node *dto.Node, index int32, fingerID []byte) (*dto.FindSuccessorFingerResp, error) {
	client, err := g.getConnection(node.Addr)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	resp, err := client.FindSuccessorFinger(ctx, &proto.FindSuccessorFingerReq{
		Index:    index,
		FingerId: fingerID,
	})
	if err != nil {
		return nil, err
	}
	return &dto.FindSuccessorFingerResp{
		Index:    resp.GetIndex(),
		FingerID: resp.GetFingerId(),
		Found:    resp.GetFound(),
		NextNode: &dto.Node{
			ID:   resp.GetNextNode().GetId(),
			Addr: resp.GetNextNode().GetAddr(),
		},
	}, nil
}

func (g *GrpcTransport) GetSuccessorList(node *dto.Node) ([]*dto.Node, error) {
	client, err := g.getConnection(node.Addr)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	resp, err := client.GetSuccessorList(ctx, &proto.GetSuccessorListReq{})
	if err != nil {
		return nil, err
	}
	successorList := make([]*dto.Node, 0)
	for _, n := range resp.GetSuccessorList() {
		successorList = append(successorList, &dto.Node{
			ID:   n.Id,
			Addr: n.Addr,
		})
	}
	return successorList, nil
}

func (g *GrpcTransport) StoreKey(node *dto.Node, keyID []byte, key, value string) (*dto.StoreKeyResp, error) {
	client, err := g.getConnection(node.Addr)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	resp, err := client.StoreKey(ctx, &proto.StoreKeyReq{
		KeyId: keyID,
		Entry: &proto.Pair{
			Key:   key,
			Value: value,
		},
	})
	if err != nil {
		return nil, err
	}
	return &dto.StoreKeyResp{
		Located: resp.GetLocated(),
		NextNode: &dto.Node{
			ID:   resp.GetNextNode().GetId(),
			Addr: resp.GetNextNode().GetAddr(),
		},
	}, nil
}

func (g *GrpcTransport) FindKey(node *dto.Node, keyID []byte) (*dto.FindKeyResp, error) {
	client, err := g.getConnection(node.Addr)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	resp, err := client.FindKey(ctx, &proto.FindKeyReq{
		KeyId: keyID,
	})
	if err != nil {
		return nil, err
	}
	return &dto.FindKeyResp{
		Located: resp.GetLocated(),
		KeyID:   resp.GetKeyId(),
		Entry: &dto.Pair{
			Key:   resp.GetEntry().GetKey(),
			Value: resp.GetEntry().GetValue(),
		},
		NextNode: &dto.Node{
			ID:   resp.GetNextNode().GetId(),
			Addr: resp.GetNextNode().GetAddr(),
		},
	}, nil
}

func (g *GrpcTransport) DeleteKey(node *dto.Node, keyID []byte) (*dto.DeleteKeyResp, error) {
	client, err := g.getConnection(node.Addr)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	resp, err := client.DeleteKey(ctx, &proto.DeleteKeyReq{
		KeyId: keyID,
	})
	if err != nil {
		return nil, err
	}
	return &dto.DeleteKeyResp{
		Located:  resp.GetLocated(),
		KeyExist: resp.GetKeyExist(),
		KeyID:    resp.GetKeyId(),
		NextNode: &dto.Node{
			ID:   resp.GetNextNode().GetId(),
			Addr: resp.GetNextNode().GetAddr(),
		},
	}, nil
}

func (g *GrpcTransport) TakeOverKeys(node *dto.Node, data []*dto.Data) error {
	client, err := g.getConnection(node.Addr)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	reqData := make([]*proto.Data, 0)
	for _, d := range data {
		reqData = append(reqData, &proto.Data{
			KeyId: d.KeyID,
			Entry: &proto.Pair{
				Key:   d.Key,
				Value: d.Value,
			},
		})
	}
	_, err = client.TakeOverKeys(ctx, &proto.TakeOverKeysReq{
		Data: reqData,
	})
	if err != nil {
		return err
	}
	return nil
}

func (g *GrpcTransport) Init(config *node.Config, server interface{}) error {
	if config == nil {
		return chorderr.ErrInvalidConfig
	}
	// start the listener
	listener, err := net.Listen("tcp", config.Addr)
	if err != nil {
		return err
	}

	g.config = config
	g.timeout = config.Timeout
	g.sock = listener.(*net.TCPListener)
	g.pool = make(map[string]*GrpcConn)
	g.server = grpc.NewServer(config.ServerOptions...)
	g.shutDown = 0

	n, ok := server.(*node.Node)
	if !ok {
		return chorderr.ErrParamError
	}
	n.Transport = g

	proto.RegisterChordServer(g.server, n)

	return nil
}
