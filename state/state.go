package state

import (
	"bytes"
	"math/big"

	"github.com/dogechain-lab/dogechain/crypto"
	"github.com/dogechain-lab/dogechain/state/stypes"
	"github.com/dogechain-lab/dogechain/types"
	iradix "github.com/hashicorp/go-immutable-radix"
)

var emptyCodeHash = types.EmptyCodeHash.Bytes()

type State interface {
	NewSnapshotAt(types.Hash) (Snapshot, error)
	NewSnapshot() Snapshot
	GetCode(hash types.Hash) ([]byte, bool)
}

type Snapshot interface {
	snapshotReader

	// Change object state root if there is any update of storage
	Commit(objs []*stypes.Object) (Snapshot, []byte, error)
}

// StateObject is the internal representation of the account
type StateObject struct {
	Account   *stypes.Account
	Code      []byte
	Suicide   bool
	Deleted   bool
	DirtyCode bool
	Txn       *iradix.Txn // Set it only when there is a trie

	// for quick search
	address  types.Address
	addrHash types.Hash
}

// newStateObject create a new state object
func newStateObject(address types.Address, account *stypes.Account) *StateObject {
	if account.Balance == nil {
		account.Balance = new(big.Int)
	}

	if account.CodeHash == nil {
		account.CodeHash = emptyCodeHash
	}

	if account.Root == (types.Hash{}) {
		account.Root = emptyStateHash
	}

	return stateObjectWithAddress(address, account)
}

func stateObjectWithAddress(address types.Address, account *stypes.Account) *StateObject {
	return &StateObject{
		Account:  account,
		address:  address,
		addrHash: crypto.Keccak256Hash(address[:]),
	}
}

func (s *StateObject) Empty() bool {
	return s.Account.Nonce == 0 && s.Account.Balance.Sign() == 0 && bytes.Equal(s.Account.CodeHash, emptyCodeHash)
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

	// search key
	ss.address = s.address
	ss.addrHash = s.addrHash

	return ss
}
