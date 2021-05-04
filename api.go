package xl_chord

import (
	"github.com/tristoney/xl_chord/util/chorderr"
	"github.com/tristoney/xl_chord/util/math"
)

func (n *Node) FindKeyAPI(key string) (string, error) {
	keyID := math.GetHashKey(key, n.Cnf.HashFunc)
	val, err := n.Transport.FindKey(n.Node, keyID, key)
	if err != nil {
		return "", err
	}
	return val, nil
}

func (n *Node) StoreKeyAPI(key, value string) error {
	keyID := math.GetHashKey(key, n.Cnf.HashFunc)
	_, err := n.Transport.StoreKey(n.Node, keyID, key, value)
	if err != nil {
		return err
	}
	return nil
}

func (n *Node) DeleteKeyAPI(key string) (string, bool, error)  {
	keyID := math.GetHashKey(key, n.Cnf.HashFunc)
	val, keyExist, err := n.Transport.DeleteKey(n.Node, keyID, key)
	if err == chorderr.ErrDataNotExist {
		return val, keyExist, nil
	} else {
		return val, keyExist, err
	}
}