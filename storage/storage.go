package storage

import (
	"bytes"
	"github.com/tristoney/xl_chord/dto"
	"github.com/tristoney/xl_chord/util"
	"github.com/tristoney/xl_chord/util/chorderr"
	"hash"
)


type Storage interface {
	Get(string) ([]byte, error) // Get the base-64 encoded value of key
	Set(string, string) (string, error)	// Set the base-64 encoded value to the key
	Delete(string) (string, error) // Delete the k-v pair
	Between(string, string) ([]*dto.Pair, error) // Get k-v pairs between the range
	Smaller(string) ([]*dto.Pair, error) // Get k-v pairs smaller than the up_bound
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

func (m *MappedData) Between(from, to string) ([]*dto.Pair, error) {
	pairs := make([]*dto.Pair, 0)
	for k, v := range m.data {
		hashedKey, err := util.HashKey(k, m.hashFunc)
		if err != nil {
			continue
		}
		if util.RightClosedBetween(hashedKey, []byte(from), []byte(to)) {
			pair := &dto.Pair{
				Key:   k,
				Value: string(v),
			}
			pairs = append(pairs, pair)
		}
	}
	return pairs, nil
}

func (m *MappedData) Smaller(upBound string) ([]*dto.Pair, error) {
	pairs := make([]*dto.Pair, 0)
	for k, v := range m.data {
		hashedKey, err := util.HashKey(k, m.hashFunc)
		if err != nil {
			continue
		}
		if bytes.Compare(hashedKey, []byte(upBound)) <= 0 {
			pair := &dto.Pair{
				Key:   k,
				Value: string(v),
			}
			pairs = append(pairs, pair)
		}
	}
	return pairs, nil
}
