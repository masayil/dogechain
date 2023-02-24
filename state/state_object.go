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

// stateObject is the internal representation of the account
type stateObject struct {
	account *stypes.Account
	code    []byte

	// status fields, open readable?
	suicide   bool
	deleted   bool
	dirtyCode bool

	// live object radix trie Transaction. Set it only when there is a trie
	radixTxn *iradix.Txn

	// associated transiction Transaction, to update its journal
	transitionTxn *Txn

	// for quick search, inner fileds only
	address  types.Address
	addrHash types.Hash
}

// newStateObject create a new state object
func newStateObject(transitionTxn *Txn, address types.Address, account *stypes.Account) *stateObject {
	if account == nil {
		account = new(stypes.Account)
	}

	if account.Balance == nil {
		account.Balance = new(big.Int)
	}

	if account.CodeHash == nil {
		account.CodeHash = emptyCodeHash
	}

	if account.StorageRoot == (types.Hash{}) {
		account.StorageRoot = emptyStateHash
	}

	return stateObjectWithAddress(transitionTxn, address, account)
}

func stateObjectWithAddress(transitionTxn *Txn, address types.Address, account *stypes.Account) *stateObject {
	return &stateObject{
		account:       account,
		address:       address,
		addrHash:      crypto.Keccak256Hash(address[:]),
		transitionTxn: transitionTxn,
	}
}

func (s *stateObject) Empty() bool {
	return s.Nonce() == 0 && s.Balance().Sign() == 0 && bytes.Equal(s.CodeHash(), emptyCodeHash)
}

// Copy makes a copy of the state object
func (s *stateObject) Copy() *stateObject {
	ss := new(stateObject)

	// copy account
	ss.account = s.account.Copy()

	ss.suicide = s.suicide
	ss.deleted = s.deleted
	ss.dirtyCode = s.dirtyCode
	ss.code = s.code

	if s.radixTxn != nil {
		ss.radixTxn = s.radixTxn.CommitOnly().Txn()
	}

	ss.transitionTxn = s.transitionTxn

	// search key
	ss.address = s.address
	ss.addrHash = s.addrHash

	return ss
}

func (s *stateObject) AddBalance(balance *big.Int) {
	s.SetBalance(new(big.Int).Add(s.Balance(), balance))
}

func (s *stateObject) SubBalance(balance *big.Int) {
	s.SetBalance(new(big.Int).Sub(s.Balance(), balance))
}

func (s *stateObject) SetBalance(balance *big.Int) {
	s.transitionTxn.journal.append(balanceChange{
		account: &s.address,
		prev:    new(big.Int).Set(s.Balance()),
	})
	s.setBalance(balance)
}

func (s *stateObject) setBalance(balance *big.Int) {
	s.account.Balance = balance
}

// Address returns the address of the contract/account
func (s *stateObject) Address() types.Address {
	return s.address
}

func (s *stateObject) AddressHash() types.Hash {
	return s.addrHash
}

// Code returns the contract code associated with this object, if any.
func (s *stateObject) Code() []byte {
	if s.dirtyCode {
		return s.code
	}

	if bytes.Equal(s.CodeHash(), emptyCodeHash) {
		return nil
	}

	code, _ := s.transitionTxn.snapshot.GetCode(types.BytesToHash(s.CodeHash()))

	// cache the code, but it is not dirty
	s.code = code

	return code
}

func (s *stateObject) SetCode(codeHash types.Hash, code []byte) {
	prevcode := s.Code()
	// journal change
	s.transitionTxn.journal.append(codeChange{
		account:  &s.address,
		prevhash: s.CodeHash(),
		prevcode: prevcode,
	})
	s.setCode(codeHash, code)
}

func (s *stateObject) setCode(codeHash types.Hash, code []byte) {
	s.code = code
	s.dirtyCode = true
	s.account.CodeHash = codeHash[:]
}

func (s *stateObject) CodeHash() []byte {
	return s.account.CodeHash
}

func (s *stateObject) Nonce() uint64 {
	return s.account.Nonce
}

func (s *stateObject) SetNonce(nonce uint64) {
	s.transitionTxn.journal.append(nonceChange{
		account: &s.address,
		prev:    s.Nonce(),
	})
	s.setNonce(nonce)
}

func (s *stateObject) setNonce(nonce uint64) {
	s.account.Nonce = nonce
}

func (s *stateObject) Balance() *big.Int {
	return s.account.Balance
}

func (s *stateObject) StorageRoot() types.Hash {
	return s.account.StorageRoot
}
