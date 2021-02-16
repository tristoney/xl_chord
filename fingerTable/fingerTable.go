package fingerTable

import (
	"github.com/tristoney/xl_chord/proto"
	"math/big"
)

type FingerTable []*Finger

// finger struct is the structure of fingerTable's item
// which contains a ID to represent the identifier of the key
// in the chord ring and the successor(k)
type Finger struct {
	ID        []byte
	Successor *proto.Node
}

// GetID computes the hashed identifier of id(n + 2^i) mod 2^m with
// arbitrary-precision arithmetic
func GetID(n []byte, i, m int) []byte {

	bigN := (&big.Int{}).SetBytes(n) // convert n to bigint (if use sha-1, n is a 200bit number, which cannot be compute in simple int type

	two := big.NewInt(2)
	offset := (&big.Int{}).Exp(two, big.NewInt(int64(i)), nil)
	sum := (&big.Int{}).Add(bigN, offset)		// n + 2^i

	mod := (&big.Int{}).Exp(two, big.NewInt(int64(m)), nil) // 2^m

	res := (&big.Int{}).Mod(sum, mod).Bytes()	// (n+2^i) mod 2^m

	return res
}

// NewFinger returns a new Finger object with id and node item
func NewFinger(id []byte, node *proto.Node) *Finger {
	return &Finger{
		ID:        id,
		Successor: node,
	}
}

// NewFingerTable returns a new FingerTable of the given node with default successor (nil)
func NewFingerTable(node *proto.Node, hashSize int) FingerTable {
	table := make([]*Finger, hashSize)
	for i := 0; i < hashSize; i++ {
		table[i] = NewFinger(GetID(node.Id, i, hashSize), nil)
	}
	return table
}