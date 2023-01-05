package state

import (
	"bytes"

	"github.com/dogechain-lab/dogechain/state/stypes"
	"github.com/dogechain-lab/dogechain/types"
	"github.com/dogechain-lab/fastrlp"
	iradix "github.com/hashicorp/go-immutable-radix"
)

var emptyCodeHash = types.EmptyCodeHash.Bytes()

type State interface {
	NewSnapshotAt(types.Hash) (Snapshot, error)
	NewSnapshot() Snapshot
	GetCode(hash types.Hash) ([]byte, bool)
}

type Snapshot interface {
	Get(k []byte) ([]byte, bool)
	Commit(objs []*stypes.Object) (Snapshot, []byte, error)
}

// StateObject is the internal representation of the account
type StateObject struct {
	Account   *stypes.Account
	Code      []byte
	Suicide   bool
	Deleted   bool
	DirtyCode bool
	Txn       *iradix.Txn

	// for quick search
	AddrHash types.Hash
	// for data search
	trTxn *Txn
}

func (s *StateObject) Empty() bool {
	return s.Account.Nonce == 0 && s.Account.Balance.Sign() == 0 && bytes.Equal(s.Account.CodeHash, emptyCodeHash)
}

var stateStateParserPool fastrlp.ParserPool

func (s *StateObject) GetCommittedState(key types.Hash) types.Hash {
	var (
		val []byte
		err error
	)

	// attempt to use snapshots
	if s.trTxn.snap != nil {
		// If the object was destructed in *this* block (and potentially resurrected),
		// the storage has been cleared out, and we should *not* consult the previous
		// snapshot about any storage values. The only possible alternatives are:
		//   1) resurrect happened, and new slot values were set -- those should
		//      have been handles outsite.
		//   2) we don't have new values, and can deliver empty response back
		if _, destructed := s.trTxn.snapDestructs[s.AddrHash]; destructed {
			return types.ZeroHash
		}

		val, err = s.trTxn.snap.Storage(s.AddrHash, key)
	}

	// If the snapshot is unavailable or reading from it fails, load from the database.
	if s.trTxn.snap == nil || err != nil {
		v, ok := s.Account.Trie.Get(key.Bytes())
		if !ok {
			return types.Hash{}
		}

		val = v
	}

	p := stateStateParserPool.Get()
	defer stateStateParserPool.Put(p)

	v, err := p.Parse(val)
	if err != nil {
		return types.Hash{}
	}

	res := []byte{}
	if res, err = v.GetBytes(res[:0]); err != nil {
		return types.Hash{}
	}

	return types.BytesToHash(res)
}

// Copy makes a copy of the state object
func (s *StateObject) Copy() *StateObject {
	ss := new(StateObject)

	// copy account
	ss.Account = s.Account.Copy()

	ss.Suicide = s.Suicide
	ss.Deleted = s.Deleted
	ss.DirtyCode = s.DirtyCode
	ss.Code = s.Code

	if s.Txn != nil {
		ss.Txn = s.Txn.CommitOnly().Txn()
	}

	ss.AddrHash = s.AddrHash

	return ss
}
