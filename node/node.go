package node

import (
	"context"
	"crypto/sha1"
	"github.com/mitchellh/mapstructure"
	"github.com/tristoney/xl_chord/dto"
	"github.com/tristoney/xl_chord/fingerTable"
	"github.com/tristoney/xl_chord/proto"
	"github.com/tristoney/xl_chord/storage"
	"github.com/tristoney/xl_chord/transport"
	"github.com/tristoney/xl_chord/util"
	"google.golang.org/grpc"
	"hash"
	"sync"
	"time"
)

// Config is the configuration of a Node, include gRPC options and hash function to use
type Config struct {
	ID   string
	Addr string

	HashFunc func() hash.Hash
	HashSize int // which is "m" in the paper

	ServerOptions []grpc.ServerOption // gRPC options
	DialOptions   []grpc.DialOption

	Timeout time.Duration // timeout duration of network
}

func DefaultConfig() *Config {
	return &Config{
		ID:            "",
		Addr:          "",
		HashFunc:      sha1.New,
		HashSize:      sha1.Size * 8,	// hash.Hash.Size() returns the the number of bytes
		ServerOptions: nil,
		DialOptions:   []grpc.DialOption{grpc.WithBlock(), grpc.WithInsecure(), grpc.FailOnNonTempDialError(true)},
		Timeout:       5 * time.Second,
	}
}

type Node struct {
	*dto.Node

	Cnf *Config

	Predecessor    *dto.Node
	PredecessorMtx sync.RWMutex // read-write mutex, to avoid concurrently modify of node's Predecessor

	Successor    *dto.Node
	SuccessorMtx sync.RWMutex // read-write mutex, to avoid concurrently modify of node's Successor

	FingerTable    fingerTable.FingerTable
	FingerTableMtx sync.RWMutex // read-write mutex, to avoid concurrently modify of node's FingerTable

	Storage    storage.Storage
	StorageMtx sync.RWMutex // read-write mutex, to avoid concurrently modify of node's Storage

	Transport    transport.Transport
	TransportMtx sync.RWMutex // read-write mutex, to avoid concurrently modify of node's Transport
}


func NewNode()  {

}

// RPC interface implementation

func (n *Node) GetPredecessor(ctx context.Context, req *proto.BaseReq) (*proto.NodeResp, error) {
	n.PredecessorMtx.RLock()
	defer n.PredecessorMtx.RUnlock()
	pred := n.Predecessor
	return &proto.NodeResp{
		Node: &proto.Node{
			Id:   pred.ID,
			Addr: pred.Addr,
		},
		BaseResp: &proto.BaseResp{
			ErrNo:   0,
			ErrTips: "",
			Ts:      time.Now().UnixNano(),
		},
	}, nil
}

func (n *Node) GetSuccessor(ctx context.Context, req *proto.BaseReq) (*proto.NodeResp, error) {
	n.SuccessorMtx.RLock()
	defer n.SuccessorMtx.RUnlock()
	successor := n.Successor
	return &proto.NodeResp{
		Node: &proto.Node{
			Id:   successor.ID,
			Addr: successor.Addr,
		},
		BaseResp: &proto.BaseResp{
			ErrNo:   0,
			ErrTips: "",
			Ts:      time.Now().UnixNano(),
		},
	}, nil
}

func (n *Node) Notify(ctx context.Context, req *proto.NodeReq) (*proto.BaseResp, error) {
	panic("implement me")
}

func (n *Node) FindSuccessor(ctx context.Context, req *proto.IDReq) (*proto.NodeResp, error) {
	panic("implement me")
}

func (n *Node) ChekPredecessor(ctx context.Context, req *proto.IDReq) (*proto.BaseResp, error) {
	panic("implement me")
}

func (n *Node) SetPredecessor(ctx context.Context, req *proto.NodeReq) (*proto.BaseResp, error) {
	n.PredecessorMtx.Lock()
	defer n.PredecessorMtx.Unlock()
	pred := req.Node
	n.Predecessor = &dto.Node{
		ID:   pred.Id,
		Addr: pred.Addr,
	}
	return util.NewBaseResp(), nil
}

func (n *Node) SetSuccessor(ctx context.Context, req *proto.NodeReq) (*proto.BaseResp, error) {
	n.SuccessorMtx.Lock()
	defer n.SuccessorMtx.Unlock()
	successor := req.Node
	n.Successor = &dto.Node{
		ID:   successor.Id,
		Addr: successor.Addr,
	}
	return util.NewBaseResp(), nil
}

func (n *Node) GetVal(ctx context.Context, req *proto.GetValReq) (*proto.GetValResp, error) {
	n.StorageMtx.RLock()
	defer n.StorageMtx.RUnlock()
	val, err := n.Storage.Get(req.Key)
	if err != nil {
		errorResp := &proto.GetValResp{
			Value:    nil,
			BaseResp: util.NewErrorResp(err),
		}
		return errorResp, err
	}
	return &proto.GetValResp{
		Value:    val,
		BaseResp: util.NewBaseResp(),
	}, nil
}

func (n *Node) SetKey(ctx context.Context, req *proto.SetKeyReq) (*proto.SetKeyResp, error) {
	n.StorageMtx.Lock()
	defer n.StorageMtx.Unlock()
	pair := req.GetPair()
	key := pair.Key
	value := pair.Value
	storedKey, err := n.Storage.Set(key, value)
	if err != nil {
		return &proto.SetKeyResp{
			Id:       nil,
			BaseResp: util.NewErrorResp(err),
		}, err
	}
	return &proto.SetKeyResp{
		Id:       []byte(storedKey),
		BaseResp: util.NewBaseResp(),
	}, nil
}

func (n *Node) DeleteKey(ctx context.Context, req *proto.DeleteKeyReq) (*proto.DeleteKeyResp, error) {
	n.StorageMtx.Lock()
	defer n.StorageMtx.Unlock()
	key := req.GetKey()
	storedKey, err := n.Storage.Delete(key)
	if err != nil {
		return &proto.DeleteKeyResp{
			Id:       nil,
			BaseResp: util.NewErrorResp(err),
		}, err
	}
	return &proto.DeleteKeyResp{
		Key:      key,
		Id:       []byte(storedKey),
		BaseResp: util.NewBaseResp(),
	}, nil
}

func (n *Node) MultiDelete(ctx context.Context, req *proto.MultiDeleteReq) (*proto.MultiDeleteResp, error) {
	n.StorageMtx.Lock()
	defer n.StorageMtx.Unlock()
	keys := req.GetKeys()
	deletedKeys, cnt, err := n.Storage.MDelete(keys...)
	if err != nil {
		return &proto.MultiDeleteResp{
			Keys:     nil,
			BaseResp: util.NewErrorResp(err),
		}, err
	}
	return &proto.MultiDeleteResp{
		Keys : deletedKeys,
		AffectedKeys: int32(cnt),
		BaseResp: util.NewBaseResp(),
	}, nil
}

func (n *Node) GetKeys(ctx context.Context, req *proto.GetKeysReq) (*proto.GetKeysResp, error) {
	n.StorageMtx.Lock()
	defer n.StorageMtx.Unlock()
	from := req.GetFrom()
	to := req.GetTo()
	pairs, err := n.Storage.Between(string(from), string(to))
	if err != nil {
		return &proto.GetKeysResp{
			Pairs:    nil,
			BaseResp: util.NewErrorResp(err),
		}, err
	}
	protoPairs := make([]*proto.Pair, 0)
	for _, pair := range pairs{
		var p proto.Pair
		_ = mapstructure.Decode(pair, &p)
		protoPairs = append(protoPairs, &p)
	}
	return &proto.GetKeysResp{
		Pairs:    protoPairs,
		BaseResp: util.NewBaseResp(),
	},nil
}
