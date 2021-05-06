package math

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
		return bytes.Compare(a, key) < 0 || bytes.Compare(b, key) > 0
	case -1:
		return bytes.Compare(a, key) < 0 && bytes.Compare(b, key) > 0
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

	valBig := (&big.Int{}).SetBytes(val)

	two := big.NewInt(2)

	mod := (&big.Int{}).Exp(two, big.NewInt(int64(20)), nil) // 2^m

	res := (&big.Int{}).Mod(valBig, mod).Bytes() // (n+2^i) mod 2^m
	return res
}

func ToBig(n []byte) *big.Int {
	return bigInt.SetBytes(n)
}

func IsMyKey(nodeID, predID, keyID []byte) bool {
	return bytes.Equal(nodeID, keyID) || (!bytes.Equal(keyID, predID) && Between(keyID, predID, nodeID))
}

func BigMax(byteNum int) *big.Int {
	two := (&big.Int{}).SetInt64(2)
	one := (&big.Int{}).SetInt64(1)
	m := (&big.Int{}).SetInt64(int64(20))
	max := (&big.Int{}).Exp(two, m, nil)
	return (&big.Int{}).Sub(max, one)
}

