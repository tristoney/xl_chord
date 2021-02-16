package dto

import "crypto/sha1"

type Node struct {
	ID   []byte
	Addr string
}

func NewNodeItem(id string, addr string) *Node {
	h := sha1.New()
	if _, err := h.Write([]byte(id)); err != nil {
		return nil
	}
	val := h.Sum(nil)
	return &Node{
		ID:   val,
		Addr: addr,
	}
}

type Pair struct {
	Key   string
	Value string
}
