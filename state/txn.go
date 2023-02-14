package state

import (
	"math/big"

	"github.com/dogechain-lab/dogechain/chain"
	"github.com/dogechain-lab/dogechain/crypto"
	"github.com/dogechain-lab/dogechain/state/runtime"
	"github.com/dogechain-lab/dogechain/state/snapshot"
	"github.com/dogechain-lab/dogechain/state/stypes"
	"github.com/dogechain-lab/dogechain/types"
	iradix "github.com/hashicorp/go-immutable-radix"
)

var emptyStateHash = types.StringToHash("0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421")

var (
	// logIndex is the index of the logs in the trie
	logIndex = types.BytesToHash([]byte{2}).Bytes()

	// refundIndex is the index of the refund
	refundIndex = types.BytesToHash([]byte{3}).Bytes()
)

// snapshotReader is snapshot read only APIs
type snapshotReader interface {
	GetStorage(addr types.Address, root types.Hash, key types.Hash) (types.Hash, error)
	GetAccount(addr types.Address) (*stypes.Account, error)
	GetCode(hash types.Hash) ([]byte, bool)
}

// Txn is a reference of the state
type Txn struct {
	snapshot  snapshotReader
	snapshots []*iradix.Tree
	txn       *iradix.Txn

	// for caching world state
	snap          snapshot.Snapshot
	snapDestructs map[types.Hash]struct{}              // deleted and waiting for destruction
	snapAccounts  map[types.Hash][]byte                // live snapshot accounts
	snapStorage   map[types.Hash]map[types.Hash][]byte // live snapshot storages

	// This map holds 'live' objects, which will get modified while processing a state transition.
	stateObjects map[types.Address]*StateObject
}

func NewTxn(snapshot Snapshot) *Txn {
	return newTxn(snapshot)
}

func newTxn(snapshot snapshotReader) *Txn {
	i := iradix.New()

	return &Txn{
		snapshot:     snapshot,
		snapshots:    []*iradix.Tree{},
		txn:          i.Txn(),
		stateObjects: make(map[types.Address]*StateObject),
	}
}

// SetSnap sets up the world state snapshot
func (txn *Txn) SetSnap(
	snap snapshot.Snapshot,
) {
	txn.snap = snap
	if txn.snap != nil {
		txn.snapDestructs = make(map[types.Hash]struct{})
		txn.snapAccounts = make(map[types.Hash][]byte)
		txn.snapStorage = make(map[types.Hash]map[types.Hash][]byte)
	}
}

func (txn *Txn) GetSnapObjects() (
	snapDestructs map[types.Hash]struct{},
	snapAccounts map[types.Hash][]byte,
	snapStorage map[types.Hash]map[types.Hash][]byte,
) {
	return txn.snapDestructs, txn.snapAccounts, txn.snapStorage
}

// CleanSnap cleans current snapshots
func (txn *Txn) CleanSnap() {
	if txn.snap != nil {
		txn.snap, txn.snapDestructs, txn.snapAccounts, txn.snapStorage = nil, nil, nil, nil
	}
}

func (txn *Txn) hashit(src []byte) []byte {
	return crypto.Keccak256(src)
}

// Snapshot takes a snapshot at this point in time
func (txn *Txn) Snapshot() int {
	t := txn.txn.CommitOnly()

	id := len(txn.snapshots)
	txn.snapshots = append(txn.snapshots, t)

	return id
}

// RevertToSnapshot reverts to a given snapshot
func (txn *Txn) RevertToSnapshot(id int) {
	if id > len(txn.snapshots) {
		panic("")
	}

	tree := txn.snapshots[id]
	txn.txn = tree.Txn()
}

// GetAccount returns an account
func (txn *Txn) GetAccount(addr types.Address) (*stypes.Account, bool) {
	object, exists := txn.getStateObject(addr)
	if !exists {
		return nil, false
	}

	return object.Account, true
}

func (txn *Txn) setStateObject(object *StateObject) {
	txn.stateObjects[object.address] = object
}

func (txn *Txn) getStateObject(addr types.Address) (*StateObject, bool) {
	if obj := txn.getDeletedStateObject(addr); obj != nil && !obj.Deleted {
		return obj, true
	}

	return nil, false
}

func (txn *Txn) getDeletedStateObject(addr types.Address) *StateObject {
	// Prefer live objects if any is available
	if obj := txn.stateObjects[addr]; obj != nil {
		return obj
	}

	// Try to get state from radix tree which holds transient states during block processing first
	if val, exists := txn.txn.Get(addr.Bytes()); exists {
		obj := val.(*StateObject) //nolint:forcetypeassert

		return obj.Copy()
	}

	var (
		account  *stypes.Account
		addrHash = txn.hashit(addr.Bytes())
	)

	// If no transient objects are available, attempt to use snapshots
	if txn.snap != nil {
		if acc, err := txn.snap.Account(types.BytesToHash(addrHash)); err == nil { // got
			if acc == nil {
				return nil
			}

			account = acc

			if account.Root == types.ZeroHash {
				account.Root = types.EmptyRootHash
			}

			if len(account.CodeHash) == 0 {
				account.CodeHash = emptyCodeHash
			}
		}
	}

	// If snapshot unavailable or reading from it failed, load from the database
	if account == nil {
		var err error

		account, err = txn.snapshot.GetAccount(addr)
		if err != nil {
			return nil
		} else if account == nil {
			return nil
		}
	}

	// Insert into the live set
	// the account is a new one, no need to copy it
	obj := newStateObject(addr, account)
	txn.setStateObject(obj)

	return obj
}

func (txn *Txn) upsertAccount(addr types.Address, create bool, f func(object *StateObject)) {
	object, exists := txn.getStateObject(addr)
	if !exists && create {
		object = newStateObject(addr, &stypes.Account{})
	}

	// run the callback to modify the account
	f(object)

	if object != nil {
		txn.txn.Insert(addr.Bytes(), object)
	}

	// If state snapshotting is active, cache the data til commit. Note, this
	// update mechanism is not symmetric to the deletion, because whereas it is
	// enough to track account updates at commit time, deletions need tracking
	// at transaction boundary level to ensure we capture state clearing.
	if txn.snap != nil {
		txn.snapAccounts[object.addrHash] = snapshot.SlimAccountRLP(
			object.Account.Nonce,
			object.Account.Balance,
			object.Account.Root,
			object.Account.CodeHash,
		)
	}
}

func (txn *Txn) AddSealingReward(addr types.Address, balance *big.Int) {
	txn.upsertAccount(addr, true, func(object *StateObject) {
		if object.Suicide {
			// create a only balance object if it suidcide
			*object = *newStateObject(addr, &stypes.Account{
				Balance: big.NewInt(0).SetBytes(balance.Bytes()),
			})
		} else {
			object.Account.Balance.Add(object.Account.Balance, balance)
		}
	})
}

// AddBalance adds balance
func (txn *Txn) AddBalance(addr types.Address, balance *big.Int) {
	txn.upsertAccount(addr, true, func(object *StateObject) {
		object.Account.Balance.Add(object.Account.Balance, balance)
	})
}

// SubBalance reduces the balance at address addr by amount
func (txn *Txn) SubBalance(addr types.Address, amount *big.Int) error {
	// If we try to reduce balance by 0, then it's a noop
	if amount.Sign() == 0 {
		return nil
	}

	// Check if we have enough balance to deduce amount from
	if balance := txn.GetBalance(addr); balance.Cmp(amount) < 0 {
		return runtime.ErrNotEnoughFunds
	}

	txn.upsertAccount(addr, true, func(object *StateObject) {
		object.Account.Balance.Sub(object.Account.Balance, amount)
	})

	return nil
}

// SetBalance sets the balance
func (txn *Txn) SetBalance(addr types.Address, balance *big.Int) {
	txn.upsertAccount(addr, true, func(object *StateObject) {
		object.Account.Balance.SetBytes(balance.Bytes())
	})
}

// GetBalance returns the balance of an address
func (txn *Txn) GetBalance(addr types.Address) *big.Int {
	object, exists := txn.getStateObject(addr)
	if !exists {
		return big.NewInt(0)
	}

	return object.Account.Balance
}

func (txn *Txn) EmitLog(addr types.Address, topics []types.Hash, data []byte) {
	log := &types.Log{
		Address: addr,
		Topics:  topics,
	}
	log.Data = append(log.Data, data...)

	var logs []*types.Log

	val, exists := txn.txn.Get(logIndex)
	if !exists {
		logs = []*types.Log{}
	} else {
		logs = val.([]*types.Log) //nolint:forcetypeassert
	}

	logs = append(logs, log)
	txn.txn.Insert(logIndex, logs)
}

// AddLog adds a new log
func (txn *Txn) AddLog(log *types.Log) {
	var logs []*types.Log

	data, exists := txn.txn.Get(logIndex)
	if !exists {
		logs = []*types.Log{}
	} else {
		logs = data.([]*types.Log) //nolint:forcetypeassert
	}

	logs = append(logs, log)
	txn.txn.Insert(logIndex, logs)
}

// State

var zeroHash types.Hash

func (txn *Txn) SetStorage(
	addr types.Address,
	key types.Hash,
	value types.Hash,
	config *chain.ForksInTime,
) runtime.StorageStatus {
	oldValue, err := txn.GetState(addr, key)
	if err != nil {
		return runtime.StorageReadFailed
	} else if oldValue == value {
		return runtime.StorageUnchanged
	}

	current := oldValue // current - storage dirtied by previous lines of this contract

	original, err := txn.GetCommittedState(addr, key) // storage slot before this transaction started
	if err != nil {
		return runtime.StorageReadFailed
	}

	txn.SetState(addr, key, value)

	legacyGasMetering := !config.Istanbul && (config.Petersburg || !config.Constantinople)

	if legacyGasMetering {
		if oldValue == zeroHash {
			return runtime.StorageAdded
		} else if value == zeroHash {
			txn.AddRefund(15000)

			return runtime.StorageDeleted
		}

		return runtime.StorageModified
	}

	if original == current {
		if original == zeroHash { // create slot (2.1.1)
			return runtime.StorageAdded
		}

		if value == zeroHash { // delete slot (2.1.2b)
			txn.AddRefund(15000)

			return runtime.StorageDeleted
		}

		return runtime.StorageModified
	}

	if original != zeroHash { // Storage slot was populated before this transaction started
		if current == zeroHash { // recreate slot (2.2.1.1)
			txn.SubRefund(15000)
		} else if value == zeroHash { // delete slot (2.2.1.2)
			txn.AddRefund(15000)
		}
	}

	if original == value {
		if original == zeroHash { // reset to original nonexistent slot (2.2.2.1)
			// Storage was used as memory (allocation and deallocation occurred within the same contract)
			if config.Istanbul {
				txn.AddRefund(19200)
			} else {
				txn.AddRefund(19800)
			}
		} else { // reset to original existing slot (2.2.2.2)
			if config.Istanbul {
				txn.AddRefund(4200)
			} else {
				txn.AddRefund(4800)
			}
		}
	}

	return runtime.StorageModifiedAgain
}

// SetState change the state of an address
func (txn *Txn) SetState(
	addr types.Address,
	key,
	value types.Hash,
) {
	txn.upsertAccount(addr, true, func(object *StateObject) {
		if object.Txn == nil {
			object.Txn = iradix.New().Txn()
		}

		if value == zeroHash {
			object.Txn.Insert(key.Bytes(), nil)
		} else {
			object.Txn.Insert(key.Bytes(), value.Bytes())
		}
	})
}

// GetState returns the state of the address at a given key
func (txn *Txn) GetState(addr types.Address, slot types.Hash) (types.Hash, error) {
	object, exists := txn.getStateObject(addr)
	if !exists {
		return types.Hash{}, nil
	}

	// Try to get account state from radix tree first
	// Because the latest account state should be in in-memory radix tree
	// if account state update happened in previous transactions of same block
	if object.Txn != nil {
		if val, ok := object.Txn.Get(slot.Bytes()); ok {
			if val == nil {
				return types.Hash{}, nil
			}
			//nolint:forcetypeassert
			return types.BytesToHash(val.([]byte)), nil
		}
	}

	// // get it from snapshot
	// if txn.snap != nil {
	// 	// live object
	// 	v, err := txn.snap.Storage(object.addrHash, crypto.Keccak256Hash(slot.Bytes()))
	// 	if err != nil {
	// 		return types.Hash{}, err
	// 	} else if len(v) > 0 {
	// 		return types.BytesToHash(v), nil
	// 	}
	// }

	// get it from storage
	return txn.snapshot.GetStorage(addr, object.Account.Root, slot)
}

// Nonce

// IncrNonce increases the nonce of the address
func (txn *Txn) IncrNonce(addr types.Address) {
	txn.upsertAccount(addr, true, func(object *StateObject) {
		object.Account.Nonce++
	})
}

// SetNonce reduces the balance
func (txn *Txn) SetNonce(addr types.Address, nonce uint64) {
	txn.upsertAccount(addr, true, func(object *StateObject) {
		object.Account.Nonce = nonce
	})
}

// GetNonce returns the nonce of an addr
func (txn *Txn) GetNonce(addr types.Address) uint64 {
	object, exists := txn.getStateObject(addr)
	if !exists {
		return 0
	}

	return object.Account.Nonce
}

// Code

// SetCode sets the code for an address
func (txn *Txn) SetCode(addr types.Address, code []byte) {
	txn.upsertAccount(addr, true, func(object *StateObject) {
		object.Account.CodeHash = crypto.Keccak256(code)
		object.DirtyCode = true
		object.Code = code
	})
}

func (txn *Txn) GetCode(addr types.Address) []byte {
	object, exists := txn.getStateObject(addr)
	if !exists {
		return nil
	}

	if object.DirtyCode {
		return object.Code
	}

	// TODO: handle error
	code, _ := txn.snapshot.GetCode(types.BytesToHash(object.Account.CodeHash))

	return code
}

func (txn *Txn) GetCodeSize(addr types.Address) int {
	return len(txn.GetCode(addr))
}

func (txn *Txn) GetCodeHash(addr types.Address) types.Hash {
	object, exists := txn.getStateObject(addr)
	if !exists {
		return types.Hash{}
	}

	return types.BytesToHash(object.Account.CodeHash)
}

// Suicide marks the given account as suicided
func (txn *Txn) Suicide(addr types.Address) bool {
	var suicided bool

	txn.upsertAccount(addr, false, func(object *StateObject) {
		if object == nil || object.Suicide {
			suicided = false
		} else {
			suicided = true
			object.Suicide = true
		}
		if object != nil {
			object.Account.Balance = new(big.Int)
		}
	})

	return suicided
}

// HasSuicided returns true if the account suicided
func (txn *Txn) HasSuicided(addr types.Address) bool {
	object, exists := txn.getStateObject(addr)

	return exists && object.Suicide
}

// Refund
func (txn *Txn) AddRefund(gas uint64) {
	refund := txn.GetRefund() + gas
	txn.txn.Insert(refundIndex, refund)
}

func (txn *Txn) SubRefund(gas uint64) {
	refund := txn.GetRefund() - gas
	txn.txn.Insert(refundIndex, refund)
}

func (txn *Txn) Logs() []*types.Log {
	data, exists := txn.txn.Get(logIndex)
	if !exists {
		return nil
	}

	txn.txn.Delete(logIndex)
	//nolint:forcetypeassert
	return data.([]*types.Log)
}

func (txn *Txn) GetRefund() uint64 {
	data, exists := txn.txn.Get(refundIndex)
	if !exists {
		return 0
	}

	//nolint:forcetypeassert
	return data.(uint64)
}

// GetCommittedState returns the state of the address in the trie
func (txn *Txn) GetCommittedState(addr types.Address, key types.Hash) (types.Hash, error) {
	obj, ok := txn.getStateObject(addr)
	if !ok {
		return types.Hash{}, nil
	}

	// // get it from snapshot
	// if txn.snap != nil {
	// 	// live object
	// 	v, err := txn.snap.Storage(crypto.Keccak256Hash(addr.Bytes()),
	// 		crypto.Keccak256Hash(key.Bytes()))
	// 	if err != nil {
	// 		return types.Hash{}, err
	// 	} else if len(v) > 0 {
	// 		return types.BytesToHash(v), nil
	// 	}
	// }

	// obj.trTxn = txn // reference for look up snapshots
	// return obj.GetCommittedState(types.BytesToHash(txn.hashit(key.Bytes())))

	return txn.snapshot.GetStorage(addr, obj.Account.Root, key)
}

func (txn *Txn) TouchAccount(addr types.Address) {
	txn.upsertAccount(addr, true, func(obj *StateObject) {

	})
}

// TODO, check panics with this ones

func (txn *Txn) Exist(addr types.Address) bool {
	_, exists := txn.getStateObject(addr)

	return exists
}

func (txn *Txn) Empty(addr types.Address) bool {
	obj, exists := txn.getStateObject(addr)
	if !exists {
		return true
	}

	return obj.Empty()
}

func (txn *Txn) CreateAccount(addr types.Address) {
	// prev might have been deleted
	prev := txn.getDeletedStateObject(addr)

	var prevdesctruct bool
	if txn.snap != nil {
		// destruct object when already deleted
		_, prevdesctruct = txn.snapDestructs[prev.addrHash]
		if !prevdesctruct {
			txn.snapDestructs[prev.addrHash] = struct{}{}
		}
	}

	// no trie in create account even if it is deleted.
	newobj := newStateObject(addr, &stypes.Account{})

	// TODO: journal CreateObjectChange

	if prev != nil && !prev.Deleted {
		newobj.Account.Balance = prev.Account.Balance
	}

	// cache object after balance update
	txn.setStateObject(newobj)

	// insert it to itrie
	txn.txn.Insert(addr.Bytes(), newobj)
}

func (txn *Txn) CleanDeleteObjects(deleteEmptyObjects bool) {
	remove := [][]byte{}

	txn.txn.Root().Walk(func(k []byte, v interface{}) bool {
		a, ok := v.(*StateObject)
		if !ok {
			return false
		}
		if a.Suicide || a.Empty() && deleteEmptyObjects {
			remove = append(remove, k)

			// delete it from snapshot too
			if txn.snap != nil {
				// We need to maintain account deletions explicitly (will remain set indefinitely)
				txn.snapDestructs[a.addrHash] = struct{}{}
				// Clear out any previously updated account and storage data (may be recreated via a resurrect)
				delete(txn.snapAccounts, a.addrHash)
				delete(txn.snapStorage, a.addrHash)
			}
		}

		return false
	})

	for _, k := range remove {
		v, ok := txn.txn.Get(k)
		if !ok {
			panic("it should not happen")
		}

		obj, ok := v.(*StateObject)

		if !ok {
			panic("it should not happen")
		}

		obj2 := obj.Copy()
		obj2.Deleted = true
		txn.txn.Insert(k, obj2)
	}

	// delete refunds
	txn.txn.Delete(refundIndex)
}

// func (txn *Txn) Commit(deleteEmptyObjects bool) (Snapshot, []byte) {
func (txn *Txn) Commit(deleteEmptyObjects bool) []*stypes.Object {
	txn.CleanDeleteObjects(deleteEmptyObjects)

	x := txn.txn.Commit()

	// Do a more complex thing for now
	objs := []*stypes.Object{}

	x.Root().Walk(func(k []byte, v interface{}) bool {
		a, ok := v.(*StateObject)
		if !ok {
			// We also have logs, avoid those
			return false
		}

		addr := types.BytesToAddress(k)

		obj := &stypes.Object{
			Nonce:     a.Account.Nonce,
			Address:   addr,
			Balance:   a.Account.Balance,
			Root:      a.Account.Root,
			CodeHash:  types.BytesToHash(a.Account.CodeHash),
			DirtyCode: a.DirtyCode,
			Code:      a.Code,
		}
		if a.Deleted {
			obj.Deleted = true

			if txn.snap != nil {
				// We need to maintain account deletions explicitly (will remain set indefinitely)
				txn.snapDestructs[a.addrHash] = struct{}{}
				// Clear out any previously updated data (may be recreated via a resurrect)
				delete(txn.snapAccounts, a.addrHash)
				delete(txn.snapStorage, a.addrHash)
			}
		} else {
			if a.Txn != nil { // if it has a trie, we need to iterate it
				a.Txn.Root().Walk(func(k []byte, v interface{}) bool {
					store := &stypes.StorageObject{Key: k}
					if v == nil {
						store.Deleted = true
					} else {
						store.Val = v.([]byte) //nolint:forcetypeassert
					}
					obj.Storage = append(obj.Storage, store)

					return false
				})
			}
		}

		objs = append(objs, obj)

		return false
	})

	return objs
}
