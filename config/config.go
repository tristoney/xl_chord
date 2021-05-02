package config

import (
	"google.golang.org/grpc"
	"hash"
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

func (c *Config) Validate() error {
	// todo add config check
	return nil
}

