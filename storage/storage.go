package storage

import (
	"github.com/tristoney/xl_chord/util/chorderr"
	"hash"
)

type Pair struct {
	Key   string
	Value string
}

type Storage interface {
	Get(string) ([]byte, error) // Get the base-64 encoded value of key
	Set(string, string) (string, error)	// Set the base-64 encoded value to the key
	Delete(string) (string, error) // Delete the k-v pair
	Between(string, string) ([]*Pair, error) // Get k-v pairs between the range
	MDelete(...string) ([]string, int, error) // MDelete delete a list of k-v pair
}

type MappedData struct {
	data     map[string][]byte
	hashFunc func() hash.Hash
}

func NewMappedData(hashFunc func() hash.Hash) *MappedData {
	return &MappedData{
		data:     make(map[string][]byte),
		hashFunc: hashFunc,
	}
}

func (m *MappedData) Get(key string) ([]byte, error) {
	val, ok := m.data[key]
	if !ok {
		return nil, chorderr.ErrDataNotExist
	}
	return val, nil
}

func (m *MappedData) Set(key, value string) (string, error) {
	m.data[key] = []byte(value)
	return key, nil
}

func (m *MappedData) Delete(key string) (string, error) {
	delete(m.data, key)
	return key, nil
}

func (m *MappedData) MDelete(keys ...string) ([]string, int, error) {
	for _, k := range keys {
		delete(m.data, k)
	}
	return keys, len(keys), nil
}

func (m *MappedData) Between(from, to string) ([]*Pair, error) {
	pairs := make([]*Pair, 0)
	// todo hash between
	return pairs, nil
}
