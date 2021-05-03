package storage

import (
	"github.com/tristoney/xl_chord/dto"
	"github.com/tristoney/xl_chord/util/chorderr"
	"hash"
)

type Storage interface {
	GetKey([]byte) (dto.Pair, error)               // Get the base-64 encoded value of key
	StoreKey([]byte, dto.Pair) error                  // Set the base-64 encoded value to the key
	DeleteKey([]byte) (string, error)               // Delete the k-v pair
	GetDataAsList() ([]*dto.Data, error)		// Get All Data as a slice
}

type MappedData struct {
	data     map[string]dto.Pair
	HashFunc func() hash.Hash
}

func NewMappedData(hashFunc func() hash.Hash) *MappedData {
	return &MappedData{
		data:     make(map[string]dto.Pair),
		HashFunc: hashFunc,
	}
}

func (m *MappedData) GetKey(keyID []byte) (*dto.Pair, error) {
	val, ok := m.data[string(keyID)]
	if !ok {
		return nil, chorderr.ErrDataNotExist
	}
	return &val, nil
}

func (m *MappedData) StoreKey(keyID []byte, pair dto.Pair) error {
	m.data[string(keyID)] = pair
	return nil
}

func (m *MappedData) DeleteKey(keyID []byte) (string, error) {
	val := m.data[string(keyID)]
	delete(m.data, string(keyID))
	return val.Value, nil
}

func (m *MappedData) GetDataAsList() ([]*dto.Data, error) {
	dataList := make([]*dto.Data, 0)
	for key, value := range m.data {
		dataList = append(dataList, &dto.Data{
			KeyID: []byte(key),
			Pair:  value,
		})
	}
	return dataList, nil
}

func (m *MappedData) IsEmpty() bool {
	return len(m.data) == 0
}

//func (m *MappedData) MDelete(keys ...string) ([]string, int, error) {
//	for _, k := range keys {
//		delete(m.data, k)
//	}
//	return keys, len(keys), nil
//}

//func (m *MappedData) Between(from, to string) ([]*dto.Pair, error) {
//	pairs := make([]*dto.Pair, 0)
//	for k, v := range m.data {
//		hashedKey, err := util.HashKey(k, m.HashFunc)
//		if err != nil {
//			continue
//		}
//		if util.RightClosedBetween(hashedKey, []byte(from), []byte(to)) {
//			pair := &dto.Pair{
//				Key:   k,
//				Value: string(v),
//			}
//			pairs = append(pairs, pair)
//		}
//	}
//	return pairs, nil
//}
//
//func (m *MappedData) Smaller(upBound string) ([]*dto.Pair, error) {
//	pairs := make([]*dto.Pair, 0)
//	for k, v := range m.data {
//		hashedKey, err := util.HashKey(k, m.HashFunc)
//		if err != nil {
//			continue
//		}
//		if bytes.Compare(hashedKey, []byte(upBound)) <= 0 {
//			pair := &dto.Pair{
//				Key:   k,
//				Value: string(v),
//			}
//			pairs = append(pairs, pair)
//		}
//	}
//	return pairs, nil
//}
