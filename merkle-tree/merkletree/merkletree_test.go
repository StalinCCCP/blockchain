package merkletree

import (
	"crypto/rand"
	"merkletree/merkletree/util"
	"testing"
)

func generateRandomBytes(length int) []byte {
	bytes := make([]byte, length)
	_, _ = rand.Read(bytes) // 生成随机字节
	// if err != nil {
	// 	return nil, err
	// }
	return bytes
}
func TestMerkleTree(t *testing.T) {
	tr := NewMerkleTreeImpl("testdb")
	size := 10000
	sel := generateRandomBytes(100)
	tr.NewNode(sel)
	del := generateRandomBytes(100)
	tr.NewNode(del)
	key := tr.GetNode(sel)
	proof := tr.GetProof(sel)
	for _, ele := range proof {
		key = util.ConcatHash(key, ele)
	}
	for k, ele := range tr.Root() {
		if key[k] != ele {
			panic("something wrong: new node small mount")

		}
	}
	var data [][]byte
	for i := 2; i < size; i++ {
		x := generateRandomBytes(100)
		tr.NewNode(x)
		data = append(data, x)
	}
	key = tr.GetNode(sel)
	proof = tr.GetProof(sel)
	for _, ele := range proof {
		key = util.ConcatHash(key, ele)
	}
	for k, ele := range tr.Root() {
		if key[k] != ele {
			panic("something wrong: new node")

		}
	}
	old := sel
	sel = generateRandomBytes(100)
	tr.UpdateNode(old, sel)
	key = tr.GetNode(sel)
	proof = tr.GetProof(sel)
	for _, ele := range proof {
		key = util.ConcatHash(key, ele)
	}
	for k, ele := range tr.Root() {
		if key[k] != ele {
			panic("something wrong: update node")

		}
	}
	tr.DeleteNode(del)
	key = tr.GetNode(sel)
	proof = tr.GetProof(sel)
	for _, ele := range proof {
		key = util.ConcatHash(key, ele)
	}
	for k, ele := range tr.Root() {
		if key[k] != ele {
			panic("something wrong: delete node")

		}
	}
	tr.NewNode(generateRandomBytes(100))
	key = tr.GetNode(sel)
	proof = tr.GetProof(sel)
	for _, ele := range proof {
		key = util.ConcatHash(key, ele)
	}
	for k, ele := range tr.Root() {
		if key[k] != ele {
			panic("something wrong: insert node")

		}
	}
	for _, ele := range data {
		tr.DeleteNode(ele)
	}
	key = tr.GetNode(sel)
	proof = tr.GetProof(sel)
	for _, ele := range proof {
		key = util.ConcatHash(key, ele)
	}
	for k, ele := range tr.Root() {
		if key[k] != ele {
			panic("something wrong: multiple delete node")

		}
	}
	println("all done")
}
