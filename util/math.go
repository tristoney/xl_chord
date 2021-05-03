package util

import (
	"bytes"
	"hash"
	"math/big"
)

var bigInt *big.Int

func init() {
	bigInt = &big.Int{}
}

func IsEqual(a, b []byte) bool {
	return bytes.Compare(a, b) == 0
}

// Between returns if key in (a, b) from the ring's perspective
func Between(key, a, b []byte) bool {
	switch bytes.Compare(a, b) {
	case 1:
		return bytes.Compare(a, key) == -1 || bytes.Compare(b, key) >= 0
	case -1:
		return bytes.Compare(a, key) == -1 && bytes.Compare(b, key) >= 0
	case 0:
		return bytes.Compare(a, key) != 0
	}
	return false
}

// RightClosedBetween returns if Between(key, a, b) or key == b
func RightClosedBetween(key, a, b []byte) bool {
	return Between(key, a, b) || bytes.Equal(key, b)
}

// HashKey returns the hashed key
func GetHashKey(key string, hashFunc func() hash.Hash) []byte {
	h := hashFunc()
	if _, err := h.Write([]byte(key)); err != nil {
		return nil
	}
	val := h.Sum(nil)
	return val
}

func ToBig(n []byte) *big.Int {
	return bigInt.SetBytes(n)
}

