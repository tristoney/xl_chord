package xl_chord

import (
	"github.com/tristoney/xl_chord/fingerTable"
	"github.com/tristoney/xl_chord/proto"
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

type Node struct {
	*proto.Node

	cnf *Config

	predecessor *proto.Node
	predMtx     sync.RWMutex // read-write mutex, to avoid concurrently modify of node's predecessor

	successor    *proto.Node
	successorMtx sync.RWMutex // read-write mutex, to avoid concurrently modify of node's successor

	fingerTable *fingerTable.FingerTable
	ftMtx       sync.RWMutex // read-write mutex, to avoid concurrently modify of node's fingerTable

}
