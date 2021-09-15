package node

import (
	"context"
	"crypto/sha1"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/tristoney/xl_chord/dto"
	"github.com/tristoney/xl_chord/fingerTable"
	"github.com/tristoney/xl_chord/proto"
	"github.com/tristoney/xl_chord/storage"
	"github.com/tristoney/xl_chord/transport"
	"github.com/tristoney/xl_chord/util"
	"github.com/tristoney/xl_chord/util/chorderr"
	"google.golang.org/grpc"
	"hash"
	"log"
	"math/big"
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
		HashSize:      sha1.Size * 8, // hash.Hash.Size() returns the the number of bytes
		ServerOptions: nil,
		DialOptions:   []grpc.DialOption{grpc.WithBlock(), grpc.WithInsecure(), grpc.FailOnNonTempDialError(true)},
		Timeout:       5 * time.Second,
	}
}

func (c *Config) Validate() error {
	// todo add config check
	return nl
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

	ShutdownCh chan struct{} // channel to shutdown the node server
}

func NewNode(cnf *Config, peer *dto.Node, storage storage.Storage, transport transport.Transport) (*Node, error) {
	if err := cnf.Validate(); err != nil {
		return nil, err
	}
	node := &Node{
		Cnf:        cnf,
		ShutdownCh: make(chan struct{}),
		Storage:    storage,
		Transport:  transport,
	}

	id := cnf.ID
	if cnf.ID == "" {
		id = cnf.Addr
	}
	hashedID, err := util.HashKey(id, cnf.HashFunc)
	if err != nil {
		return nil, err
	}
	hashedIDNum := (&big.Int{}).SetBytes(hashedID)
	log.Printf("New Node ID:%s\tAddr:%s\nHashedID:%s\tHashedIDNum:%#v", cnf.ID, cnf.Addr, string(hashedID), hashedIDNum)

	node.Node = &dto.Node{
		ID:   hashedID,
		Addr: cnf.Addr,
	}
	node.FingerTable = fingerTable.NewFingerTable(node.Node, cnf.HashSize)

	// init transport
	if err := node.Transport.Init(cnf, node); err != nil {
		return nil, err
	}
	if err := node.Transport.Start(); err != nil {
		return nil, err
	}

	// join the chord ring
	if err := node.join(peer); err != nil {
		return nil, err
	}

	// Periodically stabilize every 1 second
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		for {
			select {
			case <-ticker.C:
				node.stabilize()
			case <-node.ShutdownCh:
				ticker.Stop()
				return
			}
		}
	}()

	// Periodically fix fingerTable every 100 ms
	go func() {
		next := 0
		ticker := time.NewTicker(100 * time.Millisecond)
		for {
			select {
			case <-ticker.C:
				next = node.fixFinger(next)
			case <-node.ShutdownCh:
				ticker.Stop()
				return
			}
		}
	}()

	// Periodically check the predecessor alive has failed every 10 second
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		for {
			select {
			case <-ticker.C:
				node.checkPredecessor()
			case <-node.ShutdownCh:
				ticker.Stop()
				return
			}
		}
	}()
	return node, nil
}

// internal methods

// join n join a chord ring contains the peer
func (n *Node) join(peer *dto.Node) error {
	// First we check if n is on the chord ring
	// If peer is nil, which means n is the first
	// node on the ring, use n to call RPC to find successor
	clientNode := n.Node
	if peer != nil {
		peerSuccessor, err := n.Transport.FindSuccessor(peer, n.ID)
		if err != nil {
			return err
		}
		if util.IsEqual(peerSuccessor.ID, n.ID) {
			return chorderr.ErrNodeExist
		}
		clientNode = peer
	}
	successor, err := n.Transport.FindSuccessor(clientNode, n.ID)
	if err != nil {
		return err
	}
	n.SuccessorMtx.Lock()
	n.Successor = successor
	n.SuccessorMtx.Unlock()
	err = n.transferPairs(n.Successor)
	if err != nil {
		return err
	}
	return nil
}

// locate returns the key that should be store
func (n *Node) locate(key string) (*dto.Node, error) {
	id, err := util.HashKey(key, n.Cnf.HashFunc)
	if err != nil {
		return nil, err
	}
	return n.findSuccessor(id)
}

// closestPrecedingNode is the implementation of Fig 5 pseudocode
// returns the largest predecessor of id
func (n *Node) closestPrecedingNode(id []byte) *dto.Node {
	n.FingerTableMtx.RLock()
	defer n.FingerTableMtx.RUnlock()

	curr := n.Node
	m := n.Cnf.HashSize - 1
	for i := m; i >= 0; i-- {
		finger := n.FingerTable[i]
		if finger == nil || finger.Successor == nil {
			continue
			// when the fingerTable is not fixed, it may have nil finger
		}
		if util.Between(finger.ID, curr.ID, id) {
			return finger.Successor
		}
	}
	return curr
}

// findSuccessor is the implementation of Fig 5 pseudocode
// recursively find the successor of id
func (n *Node) findSuccessor(id []byte) (*dto.Node, error) {
	n.SuccessorMtx.RLock()
	defer n.SuccessorMtx.RUnlock()

	curr := n.Node
	successor := n.Successor
	if successor == nil {
		// which means n is the only node in the ring
		return curr, nil
	}

	if util.RightClosedBetween(id, n.ID, successor.ID) {
		return successor, nil
	} else {
		pred := n.closestPrecedingNode(id)
		var err error

		log.Printf("[Node:%s] pass the query to [Node: %s]", curr.Addr, pred.Addr)
		successor, err = n.Transport.FindSuccessor(pred, id)
		if err != nil {
			return nil, err
		}
		if successor == nil {
			return curr, nil
		}
		return successor, nil
	}
}

// internal method to check predecessor is failed or not
func (n *Node) checkPredecessor() {

	// todo restart service
	n.PredecessorMtx.RLock()
	pred := n.Predecessor
	n.PredecessorMtx.RUnlock()
	if pred == nil {
		return
	} else {
		err := n.Transport.CheckPredecessor(pred)
		if err == chorderr.ErrPredecessorFailed {
			n.PredecessorMtx.Lock()
			n.Predecessor = nil
			n.PredecessorMtx.Unlock()
		}
	}
}

// stabilize is the implementation of Fig 6.
func (n *Node) stabilize() {
	n.SuccessorMtx.RLock()
	successor := n.Successor
	if successor == nil {
		n.SuccessorMtx.RUnlock()
		return
	}
	n.SuccessorMtx.RUnlock()

	x, err := n.Transport.GetPredecessor(successor)
	if err != nil || x == nil {
		log.Printf("error %#v when getting predecessor of successor[%s]", err, successor.Addr)
		return
	}
	if x.ID != nil && util.Between(x.ID, n.ID, successor.ID) {
		n.SuccessorMtx.Lock()
		n.Successor = x
		n.SuccessorMtx.Unlock()
	}
	_ = n.Transport.Notify(successor, n.Node)
}

// transferPairs moves keys that ID smaller than currentNode.ID from the parentNode
func (n *Node) transferPairs(parentNode *dto.Node) error {
	if util.IsEqual(n.ID, parentNode.ID) {
		// if n is the only node do nothing
		return nil
	}
	pairs, err := n.Transport.GetKeys(parentNode, nil, n.ID)
	if err != nil {
		return err
	}
	if len(pairs) > 0 {
		log.Printf("TransferPairs From Node:[%s] To Node[%s]", parentNode.Addr, n.Addr)
	}
	keysToTransfer := make([]string, 0)
	n.StorageMtx.Lock()
	defer n.StorageMtx.Unlock()
	for _, v := range pairs {
		if v != nil {
			_, err = n.Storage.Set(v.Key, v.Value)
			if err == nil {
				continue
			}
			keysToTransfer = append(keysToTransfer, v.Key)
		}
	}
	if len(keysToTransfer) > 0 {
		// todo add retry
		_, _ = n.Transport.MultiDelete(parentNode, keysToTransfer)
	}
	return nil
}

func (n *Node) fixFinger(next int) int {
	nextHash := fingerTable.GetID(n.ID, next, n.Cnf.HashSize)
	successor, err := n.findSuccessor(nextHash)
	nextNum := (next + 1) % n.Cnf.HashSize
	if err != nil || successor == nil {
		fmt.Println("error: ", err, successor)
		fmt.Printf("finger lookup failed %x %x \n", n.ID, nextHash)
		// TODO: Check how to handle retry, passing ahead for now
		return nextNum
	}

	finger := fingerTable.NewFinger(nextHash, successor)
	n.FingerTableMtx.Lock()
	n.FingerTable[next] = finger

	n.FingerTableMtx.Unlock()

	return nextNum
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

// implementation of Fig 6
func (n *Node) Notify(ctx context.Context, req *proto.NodeReq) (*proto.BaseResp, error) {
	// check n's predecessor
	n.PredecessorMtx.RLock()
	pred := n.Predecessor
	n.PredecessorMtx.RUnlock()

	node := req.GetNode()
	if pred == nil || util.Between(node.Id, pred.ID, n.ID) {
		n.PredecessorMtx.Lock()
		defer n.PredecessorMtx.Unlock()
		n.Predecessor = &dto.Node{ID: node.Id, Addr: node.Addr}
	}
	return util.NewBaseResp(), nil
}

func (n *Node) FindSuccessor(ctx context.Context, req *proto.IDReq) (*proto.NodeResp, error) {
	successor, err := n.findSuccessor(req.GetId())
	if err != nil {
		return nil, err
	}

	if successor == nil {
		return nil, chorderr.ErrSuccessorNotExist
	}
	return &proto.NodeResp{
		Node: &proto.Node{
			Id:   successor.ID,
			Addr: successor.Addr,
		},
		BaseResp: util.NewBaseResp(),
	}, nil
}

func (n *Node) ChekPredecessor(ctx context.Context, req *proto.IDReq) (*proto.BaseResp, error) {
	n.PredecessorMtx.RLock()
	pred := n.Predecessor
	n.PredecessorMtx.RUnlock()
	if pred == nil {
		return util.NewBaseResp(), nil
	} else {
		err := n.Transport.CheckPredecessor(pred)
		if err == chorderr.ErrPredecessorFailed {
			n.PredecessorMtx.Lock()
			n.Predecessor = nil
			n.PredecessorMtx.Unlock()
			return util.NewBaseResp(), nil
		}
		return nil, err
	}
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
		Keys:         deletedKeys,
		AffectedKeys: int32(cnt),
		BaseResp:     util.NewBaseResp(),
	}, nil
}

func (n *Node) GetKeys(ctx context.Context, req *proto.GetKeysReq) (*proto.GetKeysResp, error) {
	n.StorageMtx.Lock()
	defer n.StorageMtx.Unlock()
	from := req.GetFrom()
	to := req.GetTo()
	if to == nil {
		return &proto.GetKeysResp{
			Pairs:    nil,
			BaseResp: util.NewErrorResp(chorderr.ErrParamError),
		}, chorderr.ErrParamError
	}
	var pairs []*dto.Pair
	var err error
	if from == nil {
		pairs, err = n.Storage.Smaller(string(to))
	} else {
		pairs, err = n.Storage.Between(string(from), string(to))
	}
	if err != nil {
		return &proto.GetKeysResp{
			Pairs:    nil,
			BaseResp: util.NewErrorResp(err),
		}, err
	}
	protoPairs := make([]*proto.Pair, 0)
	for _, pair := range pairs {
		var p proto.Pair
		_ = mapstructure.Decode(pair, &p)
		protoPairs = append(protoPairs, &p)
	}
	return &proto.GetKeysResp{
		Pairs:    protoPairs,
		BaseResp: util.NewBaseResp(),
	}, nil
}

func (n *Node) CheckAlive(ctx context.Context, er *proto.ER) (*proto.Pong, error) {
	return &proto.Pong{Ts: util.GetTimeForNow()}, nil
}