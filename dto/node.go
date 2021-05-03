package dto

import (
	"crypto/sha1"
	"fmt"
	"github.com/tristoney/xl_chord/util"
)

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

type Data struct {
	KeyID []byte
	Pair
}

func (n Node) String() string {
	return fmt.Sprintf("Node{%s}", n.Addr)
}

func (p *Pair) String() string {
	return fmt.Sprintf("%s: %s", p.Key, p.Value)
}

func (d *Data) String() string {
	return fmt.Sprintf("%s: [%s]", util.ToBig(d.KeyID), d.Pair)
}
