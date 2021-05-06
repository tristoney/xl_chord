package config

import (
	"crypto/sha1"
	"google.golang.org/grpc"
	"hash"
	"time"
)

// Config is the configuration of a Node, include gRPC options and hash function to use
type Config struct {
	ID   []byte
	Addr string

	HashFunc func() hash.Hash
	HashSize int // which is "m" in the paper

	ServerOptions []grpc.ServerOption // gRPC options
	DialOptions   []grpc.DialOption

	Timeout time.Duration // timeout duration of network
	MaxIdle time.Duration
}

func (c *Config) Validate() error {
	// todo add config check
	return nil
}

func DefaultConfig() *Config {
	return &Config{
		Addr:          "",
		HashFunc:      sha1.New,
		HashSize:      sha1.Size * 8, // hash.Hash.Size() returns the the number of bytes
		ServerOptions: nil,
		DialOptions:   []grpc.DialOption{grpc.WithBlock(), grpc.WithInsecure(), grpc.FailOnNonTempDialError(true)},
		Timeout:       5 * time.Second,
	}
}
