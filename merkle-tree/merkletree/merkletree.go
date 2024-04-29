package merkletree

import (
	"merkletree/hash"
	"merkletree/kvstore"
	"merkletree/merkletree/util"
)

type MerkleTree interface {
	Root() []byte
	NewNode([]byte)
	Exist([]byte) bool
	DeleteNode([]byte)
	UpdateNode(old, new []byte)
	GetProof([]byte) [][]byte
}

type MerkleTreeImpl struct {
	db   kvstore.LevelDB
	size uint32
	root uint32
	delq *util.PriorityQueue[uint32]
}

func NewMerkleTreeImpl(storage_path string) *MerkleTreeImpl {
	tr := &MerkleTreeImpl{
		db:   *kvstore.NewLevelDB(storage_path),
		size: 0,
		root: 0,
		delq: util.NewPriorityQueue(func(a, b uint32) bool {
			return a < b
		}),
	}
	tr.db.Put([]byte("size"), util.Int2Byte(tr.size))
	tr.db.Put([]byte("root"), util.Int2Byte(tr.size))
	return tr
}
func InitFromLevelDB(storage_path string) *MerkleTreeImpl {
	db := kvstore.NewLevelDB(storage_path)
	delq := util.NewPriorityQueue(func(a, b uint32) bool {
		return a < b
	})
	size, _ := db.Get([]byte("size"))
	root, _ := db.Get([]byte("root"))
	for i := 1; i <= int(util.Byte2Int(size)); i++ {
		if has, _ := db.Has(util.Int2Byte(uint32(i)*2 - 1)); !has {
			delq.Push(uint32(i)*2 - 1)
		}
	}
	tr := &MerkleTreeImpl{
		db:   *db,
		size: util.Byte2Int(size),
		root: util.Byte2Int(root),
		delq: delq,
	}
	return tr
}
func (tr MerkleTreeImpl) GetNode(content []byte) []byte {
	key := hash.Sha3Slice256(content)
	pos, _ := tr.db.Get(key)
	key = util.ConcatHash(key, pos)
	return key
}

func (tr MerkleTreeImpl) Root() []byte {
	// TODO
	pos, _ := tr.db.Get([]byte("root"))
	hash, _ := tr.db.Get(pos)
	return hash
}
func (tr *MerkleTreeImpl) updateSubproc(fa uint32, log int) {
	//buf := util.Int2Byte(sib)
	//var key []byte
	hasl, _ := tr.db.Has(util.Int2Byte(util.Fa2Lson(fa) << (log - 1)))
	hasr, _ := tr.db.Has(util.Int2Byte(util.Fa2Rson(fa) << (log - 1)))
	if !(hasl || hasr) {
		buf := util.Int2Byte(fa << log)
		key, _ := tr.db.Get(buf)
		tr.db.Delete(buf)
		tr.db.Delete(key)
		key = util.ConcatHash(key, buf)
		tr.db.Delete(key)
	} else {
		key := make([]byte, 0)
		if hasl {
			key, _ = tr.db.Get(util.Int2Byte(util.Fa2Lson(fa) << (log - 1)))
			if !hasr {
				key = util.CopyAppend(key, key...)
			}
		}
		if hasr {
			rkey, _ := tr.db.Get(util.Int2Byte(util.Fa2Rson(fa) << (log - 1)))
			key = util.CopyAppend(key, rkey...)
			if !hasl {
				key = util.CopyAppend(key, key...)
			}
		}
		key = hash.Sha3Slice256(key)
		//key = hash.Sha3Slice256(key)
		buf := util.Int2Byte(fa << log)
		if has, _ := tr.db.Has(buf); has {
			dkey, _ := tr.db.Get(buf)
			tr.db.Delete(dkey)
		}
		tr.db.Put(key, buf)
		tr.db.Put(buf, key)
	}
}
func (tr *MerkleTreeImpl) NewNode(content []byte) {

	var pos uint32
	var buf []byte

	if tr.delq.Empty() {
		tr.size++
		pos = tr.size*2 - 1
		tr.db.Put([]byte("size"), util.Int2Byte(tr.size))
		tr.root = (1 << util.Log2(pos)) // when size is full

		tr.db.Put([]byte("root"), util.Int2Byte(tr.root))
	} else {
		pos = tr.delq.Pop().(uint32)
	}
	buf = util.Int2Byte(pos)
	key := hash.Sha3Slice256(content)
	tr.db.Put(key, buf)
	key = util.ConcatHash(key, buf)
	tr.db.Put(buf, key)
	tr.db.Put(key, content)
	cur := pos
	log := 0
	for cur<<log != tr.root {
		log++
		if util.IsLson(cur) {
			cur = util.Lson2Fa(cur)
		} else {
			cur = util.Rson2Fa(cur)
		}
		tr.updateSubproc(cur, log)
	}
}

func (tr MerkleTreeImpl) Exist(content []byte) bool {
	// TODO
	key := hash.Sha3Slice256(content)
	has, _ := tr.db.Has(key)
	return has
}

func (tr *MerkleTreeImpl) DeleteNode(content []byte) {
	key := hash.Sha3Slice256(content)
	pos, _ := tr.db.Get(key)
	tr.delq.Push(util.Byte2Int(pos))
	tr.db.Delete(key)
	tr.db.Delete(pos)
	key = util.ConcatHash(key, pos)
	tr.db.Delete(key)
	cur := util.Byte2Int(pos)
	log := 0
	for cur<<log != tr.root {
		log++
		if util.IsLson(cur) {
			cur = util.Lson2Fa(cur)
		} else {
			cur = util.Rson2Fa(cur)
		}
		tr.updateSubproc(cur, log)
	}
	log = int(util.Lowcnt(tr.root))
	cur = tr.root >> log
	for {
		hasl, _ := tr.db.Has(util.Int2Byte(util.Fa2Lson(cur) << (log - 1)))
		hasr, _ := tr.db.Has(util.Int2Byte(util.Fa2Rson(cur) << (log - 1)))
		if hasl && hasr {
			break
		}
		buf := util.Int2Byte(cur << log)
		key, _ := tr.db.Get(buf)
		tr.db.Delete(key)
		tr.db.Delete(buf)
		key = util.ConcatHash(key, buf)
		tr.db.Delete(key)
		if hasl {
			cur = util.Fa2Lson(cur)
		} else {
			cur = util.Fa2Rson(cur)
		}
		log--
		tr.root = cur << log
	}

}

func (tr *MerkleTreeImpl) UpdateNode(old, new []byte) {
	key := hash.Sha3Slice256(old)
	pos, _ := tr.db.Get(key)
	tr.db.Delete(key)
	key = util.ConcatHash(key, pos)
	tr.db.Delete(key)
	cur := util.Byte2Int(pos)
	key = hash.Sha3Slice256(new)
	tr.db.Put(key, pos)
	key = util.ConcatHash(key, pos)
	tr.db.Put(pos, key)
	tr.db.Put(key, new)
	log := 0
	for cur<<log != tr.root {
		log++
		if util.IsLson(cur) {
			cur = util.Lson2Fa(cur)
		} else {
			cur = util.Rson2Fa(cur)
		}
		tr.updateSubproc(cur, log)
	}
}

func (tr MerkleTreeImpl) GetProof(content []byte) [][]byte {
	var ret [][]byte
	key := hash.Sha3Slice256(content)
	pos, _ := tr.db.Get(key)
	cur := util.Byte2Int(pos)
	log := 0
	subproc := func(cur, sib uint32) {
		buf := util.Int2Byte(sib)
		if has, _ := tr.db.Has(buf); has {
			sibkey, _ := tr.db.Get(buf)
			ret = util.CopyAppend(ret, sibkey)
		} else {
			buf = util.Int2Byte(cur)
			key, _ := tr.db.Get(buf)
			ret = util.CopyAppend(ret, key)
		}
	}
	for cur<<log != tr.root {
		if util.IsLson(cur) {
			subproc(cur<<log, util.Fa2Rson(util.Lson2Fa(cur))<<log)
			cur = util.Lson2Fa(cur)
		} else {
			subproc(cur<<log, util.Fa2Lson(util.Rson2Fa(cur))<<log)
			cur = util.Rson2Fa(cur)
		}
		log++
	}
	return ret
}
