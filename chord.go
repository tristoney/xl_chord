package xl_chord

import (
	"bytes"
	"context"
	"crypto/sha1"
	"github.com/tristoney/xl_chord/config"
	"github.com/tristoney/xl_chord/dto"
	"github.com/tristoney/xl_chord/fingerTable"
	"github.com/tristoney/xl_chord/proto"
	"github.com/tristoney/xl_chord/storage"
	"github.com/tristoney/xl_chord/util/chorderr"
	"github.com/tristoney/xl_chord/util/log"
	"github.com/tristoney/xl_chord/util/math"
	"math/big"
	"sync"
	"time"
)

var (
	defaultCnf                   *config.Config
	ChordCircleBits              int
	FingerTableSize              int
	SuccessorListSize            int
	ChordRingSize                *big.Int
	NodeStabilizeInterval        time.Duration
	NodeFixFingersInterval       time.Duration
	NodeCheckPredecessorInterval time.Duration
	NodeInitSleepInterval        time.Duration
	ReplicaBackUpInterval        time.Duration
)

func init() {
	defaultCnf = config.DefaultConfig()
	ChordCircleBits = defaultCnf.HashSize
	FingerTableSize = ChordCircleBits
	SuccessorListSize = 8
	two := (&big.Int{}).SetInt64(2)
	m := (&big.Int{}).SetInt64(int64(20))
	ChordRingSize = (&big.Int{}).Exp(two, m, nil)
	NodeStabilizeInterval = 2000 * time.Millisecond
	NodeFixFingersInterval = 500 * time.Millisecond
	NodeCheckPredecessorInterval = 1000 * time.Millisecond
	NodeInitSleepInterval = 2000 * time.Millisecond
	ReplicaBackUpInterval = 10 * time.Minute
}

func ChordAbs(a, b []byte) *big.Int {
	aInt := math.ToBig(a)
	bInt := math.ToBig(b)
	defer func() {

		if r := recover(); r != nil {

			log.Logln(log.ERROR, "Recovered in testPanic2Error")

			//check exactly what the panic was and create error.
			switch r.(type) {
			case error:
				log.Logf(log.ERROR, "panic in ChordAbs, a %d b %d\n", aInt, bInt)
			default:
				log.Logf(log.ERROR, "panic in ChordAbs, a %d b %d\n", aInt, bInt)
			}
		}

	}()
	if aInt.Cmp(bInt) == -1 {
		// a < b
		first := (&big.Int{}).Sub(ChordRingSize, bInt)
		res := (&big.Int{}).Add(first, aInt)
		return res
	} else if aInt.Cmp(bInt) == 0{
		return (&big.Int{}).SetInt64(0)
	} else {
		return (&big.Int{}).Sub(aInt, bInt)
	}
}

type Node struct {
	*dto.Node

	Cnf *config.Config

	Predecessor    *dto.Node
	PredecessorMtx sync.RWMutex // read-write mutex, to avoid concurrently modify of node's Predecessor

	SuccessorList    []*dto.Node
	SuccessorListMtx sync.RWMutex

	FingerTable    fingerTable.FingerTable
	FingerTableMtx sync.RWMutex // read-write mutex, to avoid concurrently modify of node's FingerTable

	Storage    storage.Storage
	StorageMtx sync.RWMutex // read-write mutex, to avoid concurrently modify of node's Storage

	PredecessorStorage    storage.Storage
	SuccessorStorage      storage.Storage
	PredecessorStorageMtx sync.RWMutex
	SuccessorStorageMtx   sync.RWMutex

	Transport    Transport
	TransportMtx sync.RWMutex // read-write mutex, to avoid concurrently modify of node's Transport

	shutDownCh chan struct{} // channel to shutdown the node server

	Joined bool // Joined identifies if the node join the chord ring
}

func NewNode(addr string, cnf *config.Config) (*Node, error) {
	id := math.GetHashKey(addr, cnf.HashFunc)

	cnf.ID = id
	cnf.Addr = addr
	grpcTransport := new(GrpcTransport)
	node := &Node{
		Node: &dto.Node{
			ID:   id,
			Addr: addr,
		},
		Cnf:                   cnf,
		Predecessor:           nil,
		PredecessorMtx:        sync.RWMutex{},
		SuccessorList:         make([]*dto.Node, 0, SuccessorListSize),
		SuccessorListMtx:      sync.RWMutex{},
		FingerTable:           fingerTable.NewFingerTable(id, FingerTableSize),
		FingerTableMtx:        sync.RWMutex{},
		Storage:               storage.NewMappedData(cnf.HashFunc),
		StorageMtx:            sync.RWMutex{},
		PredecessorStorage:    storage.NewMappedData(cnf.HashFunc),
		SuccessorStorage:      storage.NewMappedData(cnf.HashFunc),
		PredecessorStorageMtx: sync.RWMutex{},
		SuccessorStorageMtx:   sync.RWMutex{},
		Transport:             grpcTransport,
		TransportMtx:          sync.RWMutex{},
		shutDownCh:            make(chan struct{}),
		Joined:                false,
	}

	node.Transport.SetSender(node)
	return node, nil
}

func NewFirst(addr string, cnf *config.Config) (*Node, error) {
	id := math.GetHashKey(addr, cnf.HashFunc)
	successor := dto.Node{
		ID:   id,
		Addr: addr,
	}
	successorList := make([]*dto.Node, 0, SuccessorListSize)
	successorList = append(successorList, &successor)
	cnf.ID = id
	cnf.Addr = addr
	grpcTransport := new(GrpcTransport)
	node := &Node{
		Node: &dto.Node{
			ID:   id,
			Addr: addr,
		},
		Cnf: cnf,
		Predecessor: &dto.Node{
			ID:   id,
			Addr: addr,
		},
		PredecessorMtx:        sync.RWMutex{},
		SuccessorList:         successorList,
		SuccessorListMtx:      sync.RWMutex{},
		FingerTable:           fingerTable.NewFirst(id, successor, FingerTableSize),
		FingerTableMtx:        sync.RWMutex{},
		Storage:               storage.NewMappedData(cnf.HashFunc),
		StorageMtx:            sync.RWMutex{},
		PredecessorStorage:    storage.NewMappedData(cnf.HashFunc),
		SuccessorStorage:      storage.NewMappedData(cnf.HashFunc),
		PredecessorStorageMtx: sync.RWMutex{},
		SuccessorStorageMtx:   sync.RWMutex{},
		Transport:             grpcTransport,
		TransportMtx:          sync.RWMutex{},
		shutDownCh:            make(chan struct{}),
		Joined:                true,
	}
	node.Transport.SetSender(node)
	return node, nil
}

func (n *Node) getSuccessor() *dto.Node {
	successor := n.FingerTable.GetSuccessor()
	return &successor
}

func (n *Node) getPredecessor() *dto.Node {
	return n.Predecessor
}

func (n *Node) getSuccessorList() []*dto.Node {
	list := make([]*dto.Node, 0, SuccessorListSize)
	n.SuccessorListMtx.RLock()
	defer n.SuccessorListMtx.RUnlock()
	list = append(list, n.SuccessorList...)
	return list
}

func (n *Node) join(peerAddr string) {
	log.Logf(log.INFO, "Trying to join...")
	peer := &dto.Node{
		ID:   math.GetHashKey(peerAddr, sha1.New),
		Addr: peerAddr,
	}
	node, err := n.Transport.FindSuccessor(peer, n.ID)
	if err != nil {
		log.Logf(log.ERROR, "Join failed, %s\n", err)
		return
	}
	log.Logf(log.INFO, "Found My Successor: %s", node)
	n.updateSuccessorAndSuccessorList(node)
	if !n.Joined {
		n.Joined = true
	}
}

func (n *Node) GracefulShutdown() {
	if n.Joined && !n.Storage.IsEmpty() {
		log.Logf(log.INFO, "Initializing shutdown, moving keys...")
		successor := n.getSuccessor()
		dataList := n.getStorageKeys()
		err := n.Transport.TakeOverKeys(successor, dataList)
		retry := 1
		for {
			if err == nil || retry > 3 {
				break
			}
			err = n.Transport.TakeOverKeys(successor, dataList)
			retry += 1
		}
		log.Logf(log.INFO, "Shutting down...")
	}
	n.Stop()
}

func (n *Node) Stop() {
	close(n.shutDownCh)
	err := n.Transport.Stop()
	retry := 1
	for {
		if err == nil || retry > 3 {
			break
		}
		err = n.Transport.Stop()
		retry += 1
	}
}

func (n *Node) updateSuccessorAndSuccessorList(successor *dto.Node) {
	n.FingerTable.SetSuccessor(*successor)
	//dataList := n.getStorageKeys()
	//go n.replicaFromPredecessor(successor, dataList)
	list, err := n.Transport.GetSuccessorList(successor)
	if err != nil {
		return
	}
	newList := make([]*dto.Node, 0, SuccessorListSize)
	newList = append(newList, n.getSuccessor())
	if len(list) == SuccessorListSize {
		//newList = append(newList, list[:len(list)-1]...)
		for _, node := range list[:len(list)-1] {
			if bytes.Equal(n.ID, node.ID) {
				break
			}
			newList = append(newList, node)
		}
	} else {
		//newList = append(newList, list...)
		for _, node := range list {
			if bytes.Equal(n.ID, node.ID) {
				break
			}
			newList = append(newList, node)
		}
	}
	n.SuccessorListMtx.Lock()
	n.SuccessorList = newList
	n.SuccessorListMtx.Unlock()
}

func (n *Node) stabilize() {
	log.Logln(log.INFO, "Stabilizing...")
	if n.Joined {
		ringAlive := false
		for _, successor := range n.getSuccessorList() {
			if err := n.Transport.CheckAlive(successor); err == nil {
				n.updateSuccessorAndSuccessorList(successor)
				pred, _ := n.Transport.GetPredecessor(successor)
				if pred != nil {
					successor := n.getSuccessor()
					if !bytes.Equal(pred.ID, n.ID) && math.Between(pred.ID, n.ID, successor.ID) {
						log.Logf(log.INFO, "Node %s GetPredResp: Had successor Node %s, it's predecessor is Node %s, Change successor to Node %s",
							n.Node, successor, pred, pred)
						n.updateSuccessorAndSuccessorList(&dto.Node{
							ID:   pred.ID,
							Addr: pred.Addr,
						})
					}
				}
				successor := n.getSuccessor()

				var notifyErr error
				retry := 0
				for {
					notifyErr = n.Transport.Notify(successor, n.Node)
					if retry > 3 || notifyErr == nil {
						break
					}
					retry += 1
				}
				ringAlive = true
				log.Logf(log.INFO, "In Stabilizing, Node %s is Alive", successor)
				break
			} else {
				log.Logf(log.INFO, "Node %s is DEAD", successor)
			}
		}
		if !ringAlive {
			n.shutDownCh <- struct{}{}
			log.Logf(log.ERROR, "No functional successor found in successor list. RING IS DEAD. Initializing shutdown...")
		}
	} else {
		log.Logf(log.INFO, "Not joined yet, gonna sleep again")
	}
}

func (n *Node) fixFingers(next int) int {
	nextNum := (next + 1) % 20
	if n.Joined {
		fingerID := fingerTable.GetID(n.ID, next, 20)
		finger, err := n.Transport.FindSuccessor(n.getSuccessor(), fingerID)
		if err != nil {
			return nextNum
		}
		n.FingerTable.Put(next, fingerID, *finger)
		return nextNum
	} else {
		log.Logf(log.INFO, "Not joined yet, gonna sleep again")
		return nextNum
	}
}

func (n *Node) checkPredecessor() {
	if n.Joined {
		pred := n.getPredecessor()
		if pred == nil {
			log.Logf(log.INFO, "No Predecessor, wait for notify")
			return
		}
		if err := n.Transport.CheckAlive(pred); err != nil {
			log.Logf(log.INFO, "Predecessor Node %s is dead", pred)
			n.setPredecessor(nil)
		} else {
			log.Logf(log.INFO, "Predecessor Node %s is alive", pred)
		}
	} else {
		log.Logf(log.INFO, "Not joined yet, gonna sleep again")
	}
}

func (n *Node) setPredecessor(pred *dto.Node) {
	n.PredecessorMtx.Lock()
	n.Predecessor = pred
	n.PredecessorMtx.Unlock()
	if pred != nil {
		n.checkRedistributeKeys(pred)
		//dataList := n.getStorageKeys()
		//go n.replicaFromSuccessor(pred, dataList)
	}
}

func (n *Node) getStorageKeys() []*dto.Data {
	n.StorageMtx.RLock()
	defer n.StorageMtx.RUnlock()
	dataList, _ := n.Storage.GetDataAsList()
	return dataList
}

func (n *Node) replicaFromSuccessor(predecessor *dto.Node, dataList []*dto.Data) {
	if predecessor != nil {
		retry := 0
		for {
			if retry > 3 {
				break
			}
			retry += 1
			err := n.Transport.BackUpFromSuccessor(predecessor, dataList)
			if err == nil {
				break
			}
		}
	}
}

func (n *Node) replicaFromPredecessor(successor *dto.Node, dataList []*dto.Data) {
	if successor != nil {
		retry := 0
		for {
			if retry > 3 {
				break
			}
			retry += 1
			err := n.Transport.BackUpFromPredecessor(successor, dataList)
			if err == nil {
				break
			}
		}
	}
}

func (n *Node) syncToPredecessorAppend(predecessor *dto.Node, keyID []byte, key, value string) {
	_ = n.Transport.AppendSuccessorReplica(predecessor, keyID, key, value)
}

func (n *Node) syncToSuccessorAppend(successor *dto.Node, keyID []byte, key, value string) {
	_ = n.Transport.AppendPredecessorReplica(successor, keyID, key, value)
}

func (n *Node) removeKeyOfSuccessor(predecessor *dto.Node, keyID []byte) {
	_ = n.Transport.DeleteSuccessorReplicaKey(predecessor, keyID)
}

func (n *Node) removeKeyOfPredecessor(successor *dto.Node, keyID []byte) {
	_ = n.Transport.DeletePredecessorReplicaKey(successor, keyID)
}

func (n *Node) checkRedistributeKeys(pred *dto.Node) {
	dataList := n.getStorageKeys()
	for _, data := range dataList {
		keyID := data.KeyID
		pair := data.Pair
		if !math.IsMyKey(n.ID, pred.ID, keyID) {
			key, err := n.Transport.StoreKey(pred, keyID, pair.Key, pair.Value, true)
			if err != nil {
				return
			}
			n.StorageMtx.Lock()
			deleteKey, _, _ := n.Storage.DeleteKey(keyID)
			n.StorageMtx.Unlock()
			log.Logf(log.INFO, "Transferring key %s, value: %s", key, deleteKey)
		}
	}
}

func (n *Node) closestPrecedingNode(keyID []byte) *dto.Node {
	// todo
	minAbs := math.BigMax(ChordCircleBits / 8)
	node := &dto.Node{
		ID:   n.ID,
		Addr: n.Addr,
	}
	for i := 0; i < n.FingerTable.Len(); i++ {
		entry := n.FingerTable.Get(i)
		entryNode := entry.Successor
		fingerAbs := ChordAbs(entry.ID, keyID)
		if fingerAbs.Cmp(minAbs) < 0 {
			minAbs = fingerAbs
			node = &entryNode
		}
	}
	for _, entry := range n.SuccessorList {
		if bytes.Equal(entry.ID, n.ID) {
			break
		} else {
			fingerAbs := ChordAbs(entry.ID, keyID)
			if fingerAbs.Cmp(minAbs) < 0 {
				minAbs = fingerAbs
				node = entry
			}
		}
	}
	return node
}

func (n *Node) replicaBackup() {
	dataList, _ := n.Storage.GetDataAsList()
	predecessor := n.getPredecessor()
	successor := n.getSuccessor()
	n.replicaFromSuccessor(predecessor, dataList)
	n.replicaFromPredecessor(successor, dataList)
}

func SpawnNode(addr string, peerAddr string) (*Node, error) {
	var node *Node
	var err error
	if peerAddr != "" {
		log.Logln(log.INFO, "Spawn node and join")
		node, err = NewNode(addr, config.DefaultConfig())
		if err != nil {
			return nil, err
		}
	} else {
		log.Logln(log.INFO, "Spawn master node.")
		node, err = NewFirst(addr, config.DefaultConfig())
		if err != nil {
			return nil, err
		}
	}
	if err := node.Transport.Init(node.Cnf, node); err != nil {
		return nil, err
	}
	if err := node.Transport.Start(); err != nil {
		return nil, err
	}
	nodeID := node.ID
	log.Logf(log.INFO, "Node %s(ID: %d) server initiated...", node.Node, math.ToBig(nodeID))
	if peerAddr != "" {
		time.Sleep(NodeInitSleepInterval)
		for {
			if node.Joined {
				break
			}
			node.join(peerAddr)
			time.Sleep(NodeInitSleepInterval)
		}
	}

	go func() {
		ticker := time.NewTicker(NodeStabilizeInterval)
		node.stabilize()
		for {
			select {
			case <-ticker.C:
				node.stabilize()
			case <-node.shutDownCh:
				ticker.Stop()
				return
			}
		}
	}()

	go func() {
		next := 0
		ticker := time.NewTicker(NodeFixFingersInterval)
		next = node.fixFingers(next)
		for {
			select {
			case <-ticker.C:
				next = node.fixFingers(next)
			case <-node.shutDownCh:
				ticker.Stop()
				return
			}
		}
	}()

	go func() {
		ticker := time.NewTicker(NodeCheckPredecessorInterval)
		node.checkPredecessor()
		for {
			select {
			case <-ticker.C:
				node.checkPredecessor()
			case <-node.shutDownCh:
				ticker.Stop()
				return
			}
		}
	}()

	go func() {
		ticker := time.NewTicker(ReplicaBackUpInterval)
		node.replicaBackup()
		for {
			select {
			case <-ticker.C:
				node.replicaBackup()
			case <-node.shutDownCh:
				ticker.Stop()
				return
			}
		}
	}()
	return node, nil
}

// RPC interface implementation

func (n *Node) CheckAlive(ctx context.Context, req *proto.CheckAliveReq) (*proto.CheckAliveResp, error) {
	return &proto.CheckAliveResp{Pong: "Pong!"}, nil
}

func (n *Node) FindSuccessor(ctx context.Context, req *proto.FindSuccessorReq) (*proto.FindSuccessorResp, error) {
	keyID := req.GetId()
	successor := n.getSuccessor()
	if math.Between(keyID, n.ID, successor.ID) {
		return &proto.FindSuccessorResp{
			Successor: &proto.Node{
				Id:   successor.ID,
				Addr: successor.Addr,
			},
			Found: true,
		}, nil
	} else if pred := n.Predecessor; pred != nil {
		if math.Between(keyID, pred.ID, n.ID) {
			return &proto.FindSuccessorResp{
				Successor: &proto.Node{
					Id:   n.ID,
					Addr: n.Addr,
				},
				Found: true,
			}, nil
		} else {
			closestNode := n.closestPrecedingNode(keyID)
			return &proto.FindSuccessorResp{
				Successor: nil,
				Found:     false,
				NextNode: &proto.Node{
					Id:   closestNode.ID,
					Addr: closestNode.Addr,
				},
			}, nil
		}
	} else {
		closestNode := n.closestPrecedingNode(keyID)
		return &proto.FindSuccessorResp{
			Successor: nil,
			Found:     false,
			NextNode: &proto.Node{
				Id:   closestNode.ID,
				Addr: closestNode.Addr,
			},
		}, nil
	}
}

func (n *Node) GetPredecessor(ctx context.Context, req *proto.GetPredecessorReq) (*proto.GetPredecessorResp, error) {
	n.PredecessorMtx.RLock()
	defer n.PredecessorMtx.RUnlock()
	if n.Predecessor == nil {
		return nil, chorderr.ErrPredecessorNotExist
	}
	return &proto.GetPredecessorResp{
		Predecessor: &proto.Node{
			Id:   n.Predecessor.ID,
			Addr: n.Predecessor.Addr,
		},
	}, nil
}

func (n *Node) Notify(ctx context.Context, req *proto.NotifyReq) (*proto.NotifyResp, error) {
	pred := req.GetNode()
	node := &dto.Node{
		ID:   pred.Id,
		Addr: pred.Addr,
	}
	if curPred := n.getPredecessor(); curPred == nil {
		log.Logf(log.INFO, "[Node %s] Notify: Has no Pred, pred is now Node: %s", n.Node, node)
		n.setPredecessor(node)
	} else {
		log.Logf(log.INFO, "[Node %s] Notify: Current Pred: Node %s, possible new Pred: Node %s. Successor is: Node %s", n.Node, curPred, node, n.getSuccessor())
		if !bytes.Equal(curPred.ID, node.ID) && math.Between(node.ID, curPred.ID, n.ID) {
			n.setPredecessor(node)
			log.Logf(log.INFO, "[Node %s] Assign new Pred: Node %s", n.Node, node)
		}
	}
	return &proto.NotifyResp{}, nil
}

func (n *Node) FindSuccessorFinger(ctx context.Context, req *proto.FindSuccessorFingerReq) (*proto.FindSuccessorFingerResp, error) {
	index := req.GetIndex()
	fingerID := req.GetFingerId()
	successor := n.getSuccessor()
	if math.Between(n.ID, successor.ID, fingerID) {
		return &proto.FindSuccessorFingerResp{
			Index:    index,
			FingerId: fingerID,
			Found:    true,
			NextNode: successor.ToProtoNode(),
		}, nil
	} else {
		return &proto.FindSuccessorFingerResp{
			Index:    index,
			FingerId: fingerID,
			Found:    false,
			NextNode: successor.ToProtoNode(),
		}, nil
	}
}

func (n *Node) GetSuccessorList(ctx context.Context, req *proto.GetSuccessorListReq) (*proto.GetSuccessorListResp, error) {
	protoList := make([]*proto.Node, 0)
	for _, node := range n.getSuccessorList() {
		protoList = append(protoList, node.ToProtoNode())
	}
	return &proto.GetSuccessorListResp{SuccessorList: protoList}, nil
}

func (n *Node) StoreKey(ctx context.Context, req *proto.StoreKeyReq) (*proto.StoreKeyResp, error) {
	keyID := req.GetKeyId()
	entry := req.GetEntry()
	pred := n.getPredecessor()
	if pred != nil {
		if math.IsMyKey(n.ID, pred.ID, keyID) {
			// I'm responsible for the key
			_ = n.Storage.StoreKey(keyID, dto.Pair{
				Key:   entry.GetKey(),
				Value: entry.GetValue(),
			})
			go n.syncToPredecessorAppend(pred, keyID, entry.GetKey(), entry.GetValue())
			go func() {
				successor := n.getSuccessor()
				n.syncToSuccessorAppend(successor, keyID, entry.GetKey(), entry.GetValue())
			}()

			return &proto.StoreKeyResp{
				Located: true,
			}, nil
		} else {
			nextNode := n.closestPrecedingNode(keyID)
			return &proto.StoreKeyResp{
				Located:  false,
				NextNode: nextNode.ToProtoNode(),
			}, nil
		}
	} else {
		nextNode := n.closestPrecedingNode(keyID)
		return &proto.StoreKeyResp{
			Located:  false,
			NextNode: nextNode.ToProtoNode(),
		}, nil
	}
}

func (n *Node) FindKey(ctx context.Context, req *proto.FindKeyReq) (*proto.FindKeyResp, error) {
	keyID := req.GetKeyId()
	pred := n.getPredecessor()
	if pred != nil {
		if math.IsMyKey(n.ID, pred.ID, keyID) {
			// I'm responsible for the key
			pair, err := n.Storage.GetKey(keyID)
			if err != nil {
				return nil, err
			}
			return &proto.FindKeyResp{
				Located: true,
				KeyId:   keyID,
				Entry: &proto.Pair{
					Key:   pair.Key,
					Value: pair.Value,
				},
			}, nil
		} else {
			nextNode := n.closestPrecedingNode(keyID)
			return &proto.FindKeyResp{
				Located:  false,
				NextNode: nextNode.ToProtoNode(),
			}, nil
		}
	} else {
		nextNode := n.closestPrecedingNode(keyID)
		return &proto.FindKeyResp{
			Located:  false,
			NextNode: nextNode.ToProtoNode(),
		}, nil
	}
}

func (n *Node) DeleteKey(ctx context.Context, req *proto.DeleteKeyReq) (*proto.DeleteKeyResp, error) {
	keyID := req.GetKeyId()
	pred := n.getPredecessor()
	if pred != nil {
		if math.IsMyKey(n.ID, pred.ID, keyID) {
			// I'm responsible for the key
			val, keyExist, _ := n.Storage.DeleteKey(keyID)
			go n.removeKeyOfSuccessor(pred, keyID)
			go func() {
				successor := n.getSuccessor()
				n.removeKeyOfPredecessor(successor, keyID)
			}()

			return &proto.DeleteKeyResp{
				Located:  true,
				KeyId:    keyID,
				KeyExist: keyExist,
				Value:    val,
			}, nil
		} else {
			nextNode := n.closestPrecedingNode(keyID)
			return &proto.DeleteKeyResp{
				Located:  false,
				NextNode: nextNode.ToProtoNode(),
			}, nil
		}
	} else {
		nextNode := n.closestPrecedingNode(keyID)
		return &proto.DeleteKeyResp{
			Located:  false,
			NextNode: nextNode.ToProtoNode(),
		}, nil
	}
}

func (n *Node) TakeOverKeys(ctx context.Context, req *proto.TakeOverKeysReq) (*proto.TakeOverKeysResp, error) {
	dataList := req.GetData()
	for _, data := range dataList {
		id := data.GetKeyId()
		pair := data.GetEntry()
		_ = n.Storage.StoreKey(id, dto.Pair{
			Key:   pair.GetKey(),
			Value: pair.GetValue(),
		})
	}
	return &proto.TakeOverKeysResp{}, nil
}

func (n *Node) BackUpFromPredecessor(ctx context.Context, req *proto.BackUpFromPredecessorReq) (*proto.BackUpFromPredecessorResp, error) {
	dataList := req.GetData()
	n.PredecessorStorageMtx.Lock()
	defer n.PredecessorStorageMtx.Unlock()
	n.PredecessorStorage.Clean()
	for _, data := range dataList {
		id := data.GetKeyId()
		pair := data.GetEntry()
		_ = n.PredecessorStorage.StoreKey(id, dto.Pair{
			Key:   pair.GetKey(),
			Value: pair.GetValue(),
		})
	}
	return &proto.BackUpFromPredecessorResp{}, nil
}

func (n *Node) BackUpFromSuccessor(ctx context.Context, req *proto.BackUpFromSuccessorReq) (*proto.BackUpFromSuccessorResp, error) {
	dataList := req.GetData()
	n.SuccessorStorageMtx.Lock()
	defer n.SuccessorStorageMtx.Unlock()
	n.SuccessorStorage.Clean()
	for _, data := range dataList {
		id := data.GetKeyId()
		pair := data.GetEntry()
		_ = n.SuccessorStorage.StoreKey(id, dto.Pair{
			Key:   pair.GetKey(),
			Value: pair.GetValue(),
		})
	}
	return &proto.BackUpFromSuccessorResp{}, nil
}

func (n *Node) DeletePredecessorReplicaKey(ctx context.Context, req *proto.DeletePredecessorReplicaKeyReq) (*proto.DeletePredecessorReplicaKeyResp, error) {
	keyID := req.GetKeyId()
	_, _, err := n.PredecessorStorage.DeleteKey(keyID)
	if err != nil {
		return nil, err
	}
	return &proto.DeletePredecessorReplicaKeyResp{}, nil
}

func (n *Node) DeleteSuccessorReplicaKey(ctx context.Context, req *proto.DeleteSuccessorReplicaKeyReq) (*proto.DeleteSuccessorReplicaKeyResp, error) {
	keyID := req.GetKeyId()
	_, _, err := n.PredecessorStorage.DeleteKey(keyID)
	if err != nil {
		return nil, err
	}
	return &proto.DeleteSuccessorReplicaKeyResp{}, nil
}

func (n *Node) AppendPredecessorReplica(ctx context.Context, req *proto.AppendPredecessorReplicaReq) (*proto.AppendPredecessorReplicaResp, error) {
	keyID := req.GetKeyId()
	entry := req.GetEntry()
	pred := n.getPredecessor()
	if pred != nil {
		_ = n.PredecessorStorage.StoreKey(keyID, dto.Pair{
			Key:   entry.GetKey(),
			Value: entry.GetValue(),
		})
		return &proto.AppendPredecessorReplicaResp{}, nil
	}
	return &proto.AppendPredecessorReplicaResp{}, nil
}

func (n *Node) AppendSuccessorReplica(ctx context.Context, req *proto.AppendSuccessorReplicaReq) (*proto.AppendSuccessorReplicaResp, error) {
	keyID := req.GetKeyId()
	entry := req.GetEntry()
	pred := n.getPredecessor()
	if pred != nil {
		_ = n.SuccessorStorage.StoreKey(keyID, dto.Pair{
			Key:   entry.GetKey(),
			Value: entry.GetValue(),
		})
		return &proto.AppendSuccessorReplicaResp{}, nil
	}
	return &proto.AppendSuccessorReplicaResp{}, nil
}
