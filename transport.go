package xl_chord

import (
	"bytes"
	"context"
	"github.com/tristoney/xl_chord/config"
	"github.com/tristoney/xl_chord/dto"
	"github.com/tristoney/xl_chord/proto"
	"github.com/tristoney/xl_chord/util/chorderr"
	"github.com/tristoney/xl_chord/util/log"
	"github.com/tristoney/xl_chord/util/math"
	"google.golang.org/grpc"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// Transport is the component of communication
// Which allows nodes to contact with each other
type Transport interface {
	Init(*config.Config, interface{}) error // Init the transport
	Start() error                           // Start start the communication of current node
	Stop() error                            // Stop shut the communication of current node
	SetSender(node *Node)

	// interactive communication
	CheckAlive(node *dto.Node) error
	FindSuccessor(node *dto.Node, id []byte) (*dto.Node, error)
	GetPredecessor(node *dto.Node) (*dto.Node, error)
	Notify(node *dto.Node, pred *dto.Node) error
	FindSuccessorFinger(node *dto.Node, index int32, fingerID []byte) (*dto.Node, error)
	GetSuccessorList(node *dto.Node) ([]*dto.Node, error)

	// storage access
	StoreKey(node *dto.Node, keyID []byte, key, value string) (string, error)
	FindKey(node *dto.Node, keyID []byte, key string) (string, error)
	DeleteKey(node *dto.Node, keyID []byte, key string) (string, bool, error)
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
	config *config.Config

	timeout time.Duration
	maxIdle time.Duration

	sock *net.TCPListener

	pool    map[string]*GrpcConn // connection pool
	poolMtx sync.RWMutex         // mutex to avoid concurrently modify

	server   *grpc.Server
	shutDown int32

	sender *Node // owner of this transport entity
}

func NewGrpcTransport(config *config.Config) (*GrpcTransport, error) {
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

func (g *GrpcTransport) SetSender(node *Node) {
	g.sender = node
}

func (g *GrpcTransport) registerNode(node *Node) {
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

func (g *GrpcTransport) reapOld() {
	ticker := time.NewTicker(60 * time.Second)

	for {
		if atomic.LoadInt32(&g.shutDown) == 1 {
			return
		}
		select {
		case <-ticker.C:
			g.reap()
		}
	}
}

func (g *GrpcTransport) reap() {
	g.poolMtx.Lock()
	defer g.poolMtx.Unlock()
	for host, conn := range g.pool {
		if time.Since(conn.lastActive) > g.maxIdle {
			_ = conn.Close()
			delete(g.pool, host)
		}
	}
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
	go g.reapOld()
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
		log.Logf(log.ERROR, "Node {%s} has failed", string(node.ID))
		return chorderr.ErrNodeFailed
	}
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	pong, err := client.CheckAlive(ctx, &proto.CheckAliveReq{})
	if err != nil || pong == nil {
		log.Logf(log.ERROR, "Node {%s} has failed", string(node.ID))
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
	if resp.Found {
		n := resp.Successor
		return &dto.Node{
			ID:   n.Id,
			Addr: n.GetAddr(),
		}, nil
	} else {
		nextNode := resp.GetNextNode()
		log.Logf(log.INFO, "Did not get successor yet, asking Node %s now...", nextNode)
		node, err = g.FindSuccessor(&dto.Node{
			ID:   nextNode.Id,
			Addr: nextNode.Addr,
		}, id)
		if err != nil {
			return nil, err
		}
		return node, nil
	}
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
	pred := resp.Predecessor
	if pred == nil {
		return nil, chorderr.ErrPredecessorNotExist
	}
	sender := g.sender
	successor := sender.getSuccessor()
	if bytes.Equal(pred.Id, g.sender.ID) && math.Between(pred.Id, sender.ID, successor.ID) {
		log.Logf(log.INFO, "Node %s GetPredResp: Had successor Node %s, it's predecessor is Node %s, Change successor to Node %s",
			sender.Node, successor, pred, pred)
		sender.updateSuccessorAndSuccessorList(&dto.Node{
			ID:   pred.Id,
			Addr: pred.Addr,
		})
	}
	successor = sender.getSuccessor()

	if err := sender.Transport.Notify(successor, sender.Node); err != nil {
		return nil, err
	}
	return &dto.Node{
		ID:   pred.GetId(),
		Addr: pred.GetAddr(),
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

func (g *GrpcTransport) FindSuccessorFinger(node *dto.Node, index int32, fingerID []byte) (*dto.Node, error) {
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
	nextNode := &dto.Node{
		ID:   resp.GetNextNode().GetId(),
		Addr: resp.GetNextNode().GetAddr(),
	}
	if resp.Found {
		log.Logf(log.INFO, "Found node for finger_id %d: Node %s", math.ToBig(fingerID), nextNode)
		return node, nil
	} else {
		log.Logf(log.INFO, "Did not get entry for finger %d (%d) yet, asking Node %s now...", math.ToBig(fingerID), index, nextNode)
		n, err := g.FindSuccessorFinger(nextNode, index, fingerID)
		if err != nil {
			return nil, err
		}
		return n, nil
	}
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

func (g *GrpcTransport) StoreKey(node *dto.Node, keyID []byte, key, value string) (string, error) {
	client, err := g.getConnection(node.Addr)
	if err != nil {
		return "", err
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
		return "", err
	}
	if resp.Located {
		log.Logf(log.DEBUG, "Key %s Stored", key)
		return key, nil
	} else {
		nextNode := resp.GetNextNode()
		n := &dto.Node{
			ID:   nextNode.Id,
			Addr: nextNode.Addr,
		}
		log.Logf(log.DEBUG, "Did not store {key: %s value: %s} yet, asking Node %s now...", key, value, n)
		k, err := g.StoreKey(n, keyID, key, value)
		if err != nil {
			return "", err
		}
		return k, nil
	}
}

func (g *GrpcTransport) FindKey(node *dto.Node, keyID []byte, key string) (string, error) {
	client, err := g.getConnection(node.Addr)
	if err != nil {
		return "", err
	}
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	resp, err := client.FindKey(ctx, &proto.FindKeyReq{
		KeyId: keyID,
	})
	if err != nil {
		return "", err
	}
	if resp.Located {
		entry := resp.GetEntry()
		val := entry.GetValue()
		log.Logf(log.DEBUG, "Value for key{%s} ID{%d} has founded: %s", key, math.ToBig(keyID), val)
		return val, nil
	} else {
		nextNode := resp.GetNextNode()
		n := &dto.Node{
			ID:   nextNode.Id,
			Addr: nextNode.Addr,
		}
		log.Logf(log.DEBUG, "Did not find key{%s} ID{%d} yet, asking Node %s now...", key, math.ToBig(keyID), n)
		v, err := g.FindKey(n, keyID, key)
		if err != nil {
			return "", err
		}
		return v, nil
	}
}

func (g *GrpcTransport) DeleteKey(node *dto.Node, keyID []byte, key string) (string, bool, error) {
	client, err := g.getConnection(node.Addr)
	if err != nil {
		return "", false, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	resp, err := client.DeleteKey(ctx, &proto.DeleteKeyReq{
		KeyId: keyID,
	})
	if err != nil {
		return "", false, err
	}
	if resp.Located {
		val := resp.GetValue()
		exist := resp.GetKeyExist()
		if !exist {
			log.Logf(log.DEBUG, "Key{%s} ID{%d} is not stored", key, math.ToBig(keyID))
		}
		log.Logf(log.DEBUG, "Key{%s} ID{%d} has deleted: %s", key, math.ToBig(keyID), val)
		return val, exist, nil
	} else {
		nextNode := resp.GetNextNode()
		n := &dto.Node{
			ID:   nextNode.Id,
			Addr: nextNode.Addr,
		}
		log.Logf(log.DEBUG, "Did not find key{%s} ID{%d} yet, asking Node %s now...", key, math.ToBig(keyID), n)
		v, exist, err := g.DeleteKey(n, keyID, key)
		if err != nil {
			return "", false, err
		}
		return v, exist, nil
	}
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

func (g *GrpcTransport) Init(config *config.Config, server interface{}) error {
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

	n, ok := server.(*Node)
	if !ok {
		return chorderr.ErrParamError
	}
	n.Transport = g

	proto.RegisterChordServer(g.server, n)

	return nil
}
