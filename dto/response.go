package dto

type FindSuccessorFingerResp struct {
	Index    int32
	FingerID []byte
	Found    bool
	NextNode *Node
}

type StoreKeyResp struct {
	Located  bool
	NextNode *Node
}

type FindKeyResp struct {
	Located  bool
	KeyID    []byte
	Entry    *Pair
	NextNode *Node
}

type DeleteKeyResp struct {
	Located       bool
	KeyExist      bool
	KeyID         []byte
	NextNode      *Node
}
