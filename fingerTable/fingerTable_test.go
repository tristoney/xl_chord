package fingerTable

import (
	"crypto/sha1"
	"fmt"
	"github.com/tristoney/xl_chord/dto"
	"github.com/tristoney/xl_chord/util/math"
	"testing"
)

func tempNode(addr string) dto.Node {
	node := dto.Node{
		ID:   nil,
		Addr: addr,
	}
	nodeID := math.GetHashKey(node.Addr, sha1.New)
	node.ID = nodeID
	return node
}

func TestFingerTable(t *testing.T) {
	firstNode := tempNode("0.0.0.0:8001")
	fingerTable := NewFirst(firstNode.ID, firstNode, sha1.Size * 8)
	fmt.Println(fingerTable)
	fmt.Println(fingerTable.GetSuccessor())
	secondNode := tempNode("0.0.0.0:8002")
	fingerTable.Put(1, secondNode.ID, secondNode)
	fmt.Println(fingerTable)
	fmt.Println(fingerTable.GetSuccessor())
	thirdNode := tempNode("0.0.0.0:8003")
	fingerTable.Put(1, thirdNode.ID, thirdNode)
	fmt.Println(fingerTable)
	fmt.Println(fingerTable.GetSuccessor())
	fingerTable.SetSuccessor(secondNode)
	fmt.Println(fingerTable)
	fmt.Println(fingerTable.GetSuccessor())
	fmt.Printf("fingerTable length: %d\n", fingerTable.Len())
}
