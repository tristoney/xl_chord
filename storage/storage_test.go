package storage

import (
	"crypto/sha1"
	"fmt"
	"github.com/tristoney/xl_chord/dto"
	"github.com/tristoney/xl_chord/util"
	"testing"
)

func TestMappedData(t *testing.T) {
	var TestMap = MappedData{
		data:     make(map[string]dto.Pair),
		HashFunc: sha1.New,
	}
	h := TestMap.HashFunc
	id1 := util.GetHashKey("1", h)
	_ = TestMap.StoreKey(id1, dto.Pair{
		Key:   "1",
		Value: "one",
	})
	_ = TestMap.StoreKey(util.GetHashKey("2", h), dto.Pair{
		Key:   "2",
		Value: "two",
	})
	_ = TestMap.StoreKey(util.GetHashKey("3", h), dto.Pair{
		Key:   "3",
		Value: "three",
	})
	fmt.Println(TestMap.GetKey(util.GetHashKey("2", h)))
	list, _ := TestMap.GetDataAsList()
	fmt.Println(list)
	s, err := TestMap.DeleteKey(util.GetHashKey("2", h))
	if err != nil {
		return
	} else {
		fmt.Println(s)
	}
	fmt.Println(TestMap.GetKey(util.GetHashKey("2", h)))
}