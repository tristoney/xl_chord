package fingerTable

import (
	"fmt"
	"github.com/tristoney/xl_chord/dto"
	"github.com/tristoney/xl_chord/util/math"
	"math/big"
)


// finger struct is the structure of fingerTable's item
// which contains a ID to represent the identifier of the key
// in the chord ring and the successor(k)
type Finger struct {
	ID        []byte
	Successor dto.Node
}

type FingerTable struct {
	NodeID  []byte
	Entries []*Finger
}

func (f Finger) String() string {
	return fmt.Sprintf("%s->%s", math.ToBig(f.ID), f.Successor)
}

func (f FingerTable) String() string {
	entries := "["
	for i, finger := range f.Entries {
		entries += fmt.Sprintf("%d: %s->%s", i, math.ToBig(finger.ID), finger.Successor)
		if i != len(f.Entries) - 1{
			entries += "\n"
		}
	}
	entries += "]"

	return fmt.Sprintf("Node{%s}'s FingerTable\n", math.ToBig(f.NodeID))+ entries
}

// GetID computes the hashed identifier of id(n + 2^i) mod 2^m with
// arbitrary-precision arithmetic
func GetID(n []byte, i, m int) []byte {

	bigN := (&big.Int{}).SetBytes(n) // convert n to bigint (if use sha-1, n is a 200bit number, which cannot be compute in simple int type

	two := big.NewInt(2)
	offset := (&big.Int{}).Exp(two, big.NewInt(int64(i)), nil)
	sum := (&big.Int{}).Add(bigN, offset) // n + 2^i

	mod := (&big.Int{}).Exp(two, big.NewInt(int64(m)), nil) // 2^m

	res := (&big.Int{}).Mod(sum, mod).Bytes() // (n+2^i) mod 2^m

	return res
}

// NewFinger returns a new Finger object with id and node item
func NewFinger(id []byte, node dto.Node) Finger {
	return Finger{
		ID:        id,
		Successor: node,
	}
}

// NewFingerTable returns a new FingerTable of the given node with default successor (nil)
func NewFingerTable(nodeID []byte, hashSize int) FingerTable {
	return FingerTable{nodeID, make([]*Finger, 0, hashSize)}
}

func NewFirst(nodeID []byte, successor dto.Node, hashSize int) FingerTable {
	entries := make([]*Finger, 0, hashSize)
	entries = append(entries, &Finger{
		ID:        GetID(nodeID, 0, hashSize),
		Successor: successor,
	})
	return FingerTable{
		NodeID:  nodeID,
		Entries: entries,
	}
}

func (f *FingerTable) Put(index int, fingerID []byte, node dto.Node) {
	finger := Finger{
		ID:        fingerID,
		Successor: node,
	}
	if len(f.Entries) > index {
		f.Entries[index] = &finger
	} else {
		f.Entries = append(f.Entries, &finger)
	}
}

func (f *FingerTable) GetSuccessor() dto.Node {
	return f.Entries[0].Successor
}

func (f *FingerTable) SetSuccessor(node dto.Node) {
	if len(f.Entries) == 0 {
		f.Entries = append(f.Entries, &Finger{
			ID: f.NodeID,
			Successor: node,
		})
	} else {
		f.Entries[0].Successor = node
	}
}

func (f *FingerTable) Get(index int) *Finger {
	return f.Entries[index]
}

func (f *FingerTable) Len() int {
	return len(f.Entries)
}