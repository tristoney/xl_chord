package xl_chord

import (
	"fmt"
	"testing"
)

func TestFunctional(t *testing.T) {
	node, err := SpawnNode("127.0.0.1:50001", "")
	if err != nil {
		return
	}
	node2, err := SpawnNode("127.0.0.1:50002", "127.0.0.1:8001")
	if err != nil {
		return
	}
	fmt.Println(node2)
	err = node.StoreKeyAPI("1", "one")
	key, _ := node.FindKeyAPI("1")
	fmt.Println(key)
}
