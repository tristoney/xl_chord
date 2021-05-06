package xl_chord

import (
	"fmt"
	"github.com/tristoney/xl_chord/util/math"
)

func (n *Node) Print() {
	fmt.Printf("\n\n\n")
	fmt.Printf("Node{%s} ID{%d}\n", n.Node, math.ToBig(n.ID))
	fmt.Printf("FingerTable:\n %s\n", n.FingerTable.Entries)
	fmt.Printf("Stored Data Map: %s", n.Storage)
	fmt.Printf("Predecesor Replica Data Map: %s", n.PredecessorStorage)
	fmt.Printf("Successor Replica Data Map: %s", n.SuccessorStorage)
	fmt.Printf("SuccessorList: %s\n", n.getSuccessorList())
	fmt.Printf("pred:Node{%s}, succ:Node{%s}\n", n.Predecessor, n.getSuccessor())
}
