package state

import (
	"bytes"
	"math/big"
	"sort"

	"github.com/dogechain-lab/dogechain/chain"
	"github.com/dogechain-lab/dogechain/crypto"
	"github.com/dogechain-lab/dogechain/state/runtime"
	"github.com/dogechain-lab/dogechain/state/snapshot"
	"github.com/dogechain-lab/dogechain/state/stypes"
	"github.com/dogechain-lab/dogechain/types"
	"github.com/dogechain-lab/fastrlp"
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
	GetStorage(addr types.Address, storageRoot types.Hash, key types.Hash) (types.Hash, error)
	GetAccount(addr types.Address) (*stypes.Account, error)
	GetCode(hash types.Hash) ([]byte, bool)
}

type revision struct {
	id           int
	journalIndex int
}

// Txn is a reference of the state
type Txn struct {
	snapshot  snapshotReader
	snapshots []*iradix.Tree
	// current radix trie transaction, caching live objects, including stateobject,
	// log, refund
	txn *iradix.Txn

	// for caching world state
	snap          snapshot.Snapshot
	snapDestructs map[types.Hash]struct{} // deleted and waiting for destruction
	snapAccounts  map[types.Hash][]byte   // live snapshot accounts
	// live snapshot storages map. [accountHash]map[slotHash]hashValue
	// keep the structrue same with persistence layer
	snapStorage map[types.Hash]map[types.Hash][]byte

	// Journal of state modifications. This is the backbone of
	// RevertToSnapshot when enabling cache snapshot.
	journal        *journal
	validRevisions []revision
}

func NewTxn(snapshot Snapshot) *Txn {
	return newTxn(snapshot)
}

func newTxn(snapshot snapshotReader) *Txn {
	i := iradix.New()

	return &Txn{
		snapshot:       snapshot,
		snapshots:      []*iradix.Tree{},
		txn:            i.Txn(),
		journal:        newJournal(),
		validRevisions: []revision{},
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

// Snapshot takes a snapshot at this point in time
func (txn *Txn) Snapshot() int {
	t := txn.txn.CommitOnly()

	id := len(txn.snapshots)
	txn.snapshots = append(txn.snapshots, t)
	// append valid revision for journal
	txn.validRevisions = append(txn.validRevisions, revision{id, txn.journal.length()})

	return id
}

// RevertToSnapshot reverts to a given snapshot
func (txn *Txn) RevertToSnapshot(id int) {
	if id > len(txn.snapshots) {
		panic("")
	}

	tree := txn.snapshots[id]
	txn.txn = tree.Txn()

	// If state snapshotting is active, we should reset to its original value,
	// otherwise the resurrect account or transient update will be persisted
	// into snapshot tree, and make the whole worldstate damage.
	idx := sort.Search(len(txn.validRevisions), func(i int) bool {
		return txn.validRevisions[i].id >= id
	})

	// Find the snapshot in the stack of valid snapshots.
	cachedSnapshot := txn.validRevisions[idx].journalIndex

	// Replay the journal to undo changes
	txn.journal.revert(txn, cachedSnapshot)
	// remove invalidated snapshots
	txn.validRevisions = txn.validRevisions[:idx]
}

func (txn *Txn) clearJournal() {
	if len(txn.journal.entries) > 0 {
		txn.journal = newJournal()
	}

	// Snapshots can be created without journal entries
	txn.validRevisions = txn.validRevisions[:0]
}

func (txn *Txn) getStateObject(addr types.Address) (*stateObject, bool) {
	if obj := txn.getDeletedStateObject(addr); obj != nil && !obj.deleted {
		return obj, true
	}

	return nil, false
}

func (txn *Txn) getDeletedStateObject(addr types.Address) *stateObject {
	// Try to get state from radix tree which holds transient states during block processing first
	if val, exists := txn.txn.Get(addr.Bytes()); exists {
		obj := val.(*stateObject) //nolint:forcetypeassert

		return obj.Copy()
	}

	var (
		account *stypes.Account
	)

	// If no transient objects are available, attempt to use snapshots
	if txn.snap != nil {
		if acc, err := txn.snap.Account(crypto.Keccak256Hash(addr.Bytes())); err == nil { // got
			if acc == nil {
				return nil
			}

			account = acc

			if account.StorageRoot == types.ZeroHash {
				account.StorageRoot = types.EmptyRootHash
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

	return stateObjectWithAddress(txn, addr, account.Copy())
}

// updateSnapAccount updates snap account by object
//
// update live object or revert to some journaled object
func (txn *Txn) updateSnapAccount(object *stateObject) {
	if txn.snap == nil || object == nil {
		return
	}

	if object.suicide {
		delete(txn.snapAccounts, object.addrHash)

		return
	}

	// If state snapshotting is active, cache the data til commit. Note, this
	// update mechanism is not symmetric to the deletion, because whereas it is
	// enough to track account updates at commit time, deletions need tracking
	// at transaction boundary level to ensure we capture state clearing.
	txn.snapAccounts[object.AddressHash()] = snapshot.SlimAccountRLP(
		object.Nonce(),
		object.Balance(),
		object.StorageRoot(),
		object.CodeHash(),
	)
}

func (txn *Txn) upsertAccount(addr types.Address, create bool, f func(object *stateObject)) {
	object, exists := txn.getStateObject(addr)
	if !exists && create {
		object = newStateObject(txn, addr, nil)
	}

	// run the callback to modify the account
	f(object)

	if object != nil {
		txn.txn.Insert(addr.Bytes(), object)
	}

	txn.updateSnapAccount(object)
}

func (txn *Txn) AddSealingReward(addr types.Address, balance *big.Int) {
	txn.upsertAccount(addr, true, func(object *stateObject) {
		if object.suicide {
			// create a only balance object if it suidcide
			*object = *newStateObject(txn, addr, &stypes.Account{
				Balance: new(big.Int).SetBytes(balance.Bytes()),
			})
		} else {
			object.AddBalance(balance)
		}
	})
}

// AddBalance adds balance
func (txn *Txn) AddBalance(addr types.Address, balance *big.Int) {
	// update the account even it add 0
	txn.upsertAccount(addr, true, func(object *stateObject) {
		object.AddBalance(balance)
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

	txn.upsertAccount(addr, true, func(object *stateObject) {
		object.SubBalance(amount)
	})

	return nil
}

// SetBalance sets the balance
func (txn *Txn) SetBalance(addr types.Address, balance *big.Int) {
	txn.upsertAccount(addr, true, func(object *stateObject) {
		object.SetBalance(balance)
	})
}

// GetBalance returns the balance of an address
func (txn *Txn) GetBalance(addr types.Address) *big.Int {
	object, exists := txn.getStateObject(addr)
	if !exists {
		return big.NewInt(0)
	}

	return object.Balance()
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
	txn.upsertAccount(addr, true, func(object *stateObject) {
		if object.radixTxn == nil {
			object.radixTxn = iradix.New().Txn()
		}

		if value == zeroHash {
			object.radixTxn.Insert(key.Bytes(), nil)
		} else {
			object.radixTxn.Insert(key.Bytes(), value.Bytes())
		}
	})
}

// GetState returns the state of the address at a given key
//
// The state might be transient, remember to query the not committed trie
func (txn *Txn) GetState(addr types.Address, slot types.Hash) (types.Hash, error) {
	object, exists := txn.getStateObject(addr)
	if !exists {
		return types.Hash{}, nil
	}

	// Try to get account state from radix tree first
	// Because the latest account state should be in in-memory radix tree
	// if account state update happened in previous transactions of same block
	if object.radixTxn != nil {
		if val, ok := object.radixTxn.Get(slot.Bytes()); ok {
			if val == nil {
				return types.Hash{}, nil
			}
			//nolint:forcetypeassert
			return types.BytesToHash(val.([]byte)), nil
		}
	}

	return txn.getStorageCommitted(object, slot)
}

// Nonce

// IncrNonce increases the nonce of the address
func (txn *Txn) IncrNonce(addr types.Address) {
	txn.upsertAccount(addr, true, func(object *stateObject) {
		object.SetNonce(object.Nonce() + 1)
	})
}

// SetNonce set nonce directly
func (txn *Txn) SetNonce(addr types.Address, nonce uint64) {
	txn.upsertAccount(addr, true, func(object *stateObject) {
		object.SetNonce(nonce)
	})
}

// GetNonce returns the nonce of an addr
func (txn *Txn) GetNonce(addr types.Address) uint64 {
	object, exists := txn.getStateObject(addr)
	if !exists {
		return 0
	}

	return object.Nonce()
}

// Code

// SetCode sets the code for an address
func (txn *Txn) SetCode(addr types.Address, code []byte) {
	txn.upsertAccount(addr, true, func(object *stateObject) {
		object.SetCode(crypto.Keccak256Hash(code), code)
	})
}

func (txn *Txn) GetCode(addr types.Address) []byte {
	object, exists := txn.getStateObject(addr)
	if !exists {
		return nil
	}

	return object.Code()
}

func (txn *Txn) GetCodeSize(addr types.Address) int {
	return len(txn.GetCode(addr))
}

func (txn *Txn) GetCodeHash(addr types.Address) types.Hash {
	object, exists := txn.getStateObject(addr)
	if !exists {
		return types.Hash{}
	}

	return types.BytesToHash(object.CodeHash())
}

// Suicide marks the given account as suicided
func (txn *Txn) Suicide(addr types.Address) bool {
	var suicided bool

	txn.upsertAccount(addr, false, func(object *stateObject) {
		change := suicideChange{
			account:     &addr,
			prevbalance: new(big.Int),
		}
		// cache prev object
		if object != nil {
			change.prev = object.suicide
			change.prevbalance.Set(object.Balance())
		}
		// journal change
		txn.journal.append(change)

		// update value
		if object == nil || object.suicide {
			suicided = false
		} else {
			suicided = true
			object.suicide = true
		}
		if object != nil {
			object.SetBalance(new(big.Int))
		}
	})

	return suicided
}

// HasSuicided returns true if the account suicided
func (txn *Txn) HasSuicided(addr types.Address) bool {
	object, exists := txn.getStateObject(addr)

	return exists && object.suicide
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

// Logs returns and clears all logs held in txn trie
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

func GetCachedCommittedStorage(
	cachedSnap snapshot.Snapshot,
	addrHash types.Hash,
	slot types.Hash,
) (found bool, value types.Hash, err error) {
	// hash slot
	enc, err := cachedSnap.Storage(addrHash, crypto.Keccak256Hash(slot.Bytes()))
	if err != nil {
		return false, types.Hash{}, err
	} else if len(enc) > 0 {
		// The storage value is rlp encoded
		p := fastrlp.Parser{}

		v, err := p.Parse(enc)
		if err != nil {
			return false, types.Hash{}, err
		}

		res := []byte{}
		if res, err = v.GetBytes(res[:0]); err != nil {
			return false, types.Hash{}, err
		}

		return true, types.BytesToHash(res), nil
	}

	return false, types.Hash{}, nil
}

func (txn *Txn) getStorageCommitted(obj *stateObject, slot types.Hash) (types.Hash, error) {
	// query from storage first
	if txn.snap != nil {
		addrHash := obj.AddressHash()
		// If the object was destructed in *this* block (and potentially resurrected),
		// the storage has been cleared out, and we should *not* consult the previous
		// snapshot about any storage values. The only possible alternatives are:
		//   1) resurrect happened, and new slot values were set -- those should
		//      have been handles via pendingStorage above.
		//   2) we don't have new values, and can deliver empty response back
		if _, destructed := txn.snapDestructs[addrHash]; destructed {
			return types.Hash{}, nil
		}

		found, value, err := GetCachedCommittedStorage(txn.snap, addrHash, slot)
		if err != nil {
			return value, err
		} else if found {
			return value, nil
		}
	}

	return txn.snapshot.GetStorage(obj.Address(), obj.StorageRoot(), slot)
}

// GetCommittedState returns the state of the address in the trie
//
// The state is committed (persisted, too).
func (txn *Txn) GetCommittedState(addr types.Address, slot types.Hash) (types.Hash, error) {
	// If the snapshot is unavailable or reading from it fails, load from the database.
	obj, ok := txn.getStateObject(addr)
	if !ok {
		return types.Hash{}, nil
	}

	return txn.getStorageCommitted(obj, slot)
}

func (txn *Txn) TouchAccount(addr types.Address) {
	txn.upsertAccount(addr, true, func(obj *stateObject) {

	})
}

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

	// cache reset change
	var prevdestruct bool

	if txn.snap != nil && prev != nil {
		// destruct object when already deleted
		prevAddrHash := prev.AddressHash()
		_, prevdestruct = txn.snapDestructs[prevAddrHash]
		if !prevdestruct {
			txn.snapDestructs[prevAddrHash] = struct{}{}
		}
	}

	// create a new object no matter exists or not
	obj := newStateObject(txn, addr, nil)

	if prev != nil { // journal reset status
		txn.journal.append(resetObjectChange{prev: prev, prevdestruct: prevdestruct})
	}

	if prev != nil && !prev.deleted {
		obj.SetBalance(prev.Balance())
	}

	// insert it to itrie
	txn.txn.Insert(addr.Bytes(), obj)
}

// CleanDeleteObjects clears deleted objects and invalid its journals.
//
// Byzantium fork is alwals on, so reverting across transactions is no allow.
func (txn *Txn) CleanDeleteObjects(deleteEmptyObjects bool) {
	remove := [][]byte{}

	txn.txn.Root().Walk(func(k []byte, v interface{}) bool {
		a, ok := v.(*stateObject)
		if !ok {
			return false
		}
		if a.suicide || a.Empty() && deleteEmptyObjects {
			remove = append(remove, k)
		}

		return false
	})

	for _, k := range remove {
		v, ok := txn.txn.Get(k)
		if !ok {
			panic("it should not happen")
		}

		obj, ok := v.(*stateObject)

		if !ok {
			panic("it should not happen")
		}

		obj2 := obj.Copy()
		obj2.deleted = true
		txn.txn.Insert(k, obj2)
	}

	// delete refunds
	txn.txn.Delete(refundIndex)

	// Invalidate journal because reverting across transactions is not allowed.
	txn.clearJournal()
}

// func (txn *Txn) Commit(deleteEmptyObjects bool) (Snapshot, []byte) {
func (txn *Txn) Commit(deleteEmptyObjects bool) []*stypes.Object {
	txn.CleanDeleteObjects(deleteEmptyObjects)

	x := txn.txn.Commit()

	// Do a more complex thing for now
	objs := []*stypes.Object{}

	x.Root().Walk(func(k []byte, v interface{}) bool {
		sobj, ok := v.(*stateObject)
		if !ok {
			// We also have logs, avoid those
			return false
		}

		addr := types.BytesToAddress(k)

		// for storage value marshaling
		storeAr := fastrlp.DefaultArenaPool.Get()
		defer fastrlp.DefaultArenaPool.Put(storeAr)

		obj := &stypes.Object{
			Nonce:     sobj.Nonce(),
			Address:   addr,
			Balance:   sobj.Balance(),
			Root:      sobj.StorageRoot(),
			CodeHash:  types.BytesToHash(sobj.CodeHash()),
			DirtyCode: sobj.dirtyCode,
			Code:      sobj.Code(),
		}
		if sobj.deleted {
			obj.Deleted = true

			// If state snapshotting is active, also mark the destruction there.
			// Note, we can't do this only at the end of a block because multiple
			// transactions within the same block might self destruct and then
			// resurrect an account; but the snapshotter needs both events.
			if txn.snap != nil {
				addrHash := sobj.AddressHash()
				// We need to maintain account deletions explicitly (will remain set indefinitely)
				txn.snapDestructs[addrHash] = struct{}{}
				// Clear out any previously updated data (may be recreated via a resurrect)
				delete(txn.snapAccounts, addrHash)
				delete(txn.snapStorage, addrHash)
			}
		} else {
			if sobj.radixTxn != nil { // if it has a trie, we need to iterate it
				sobj.radixTxn.Root().Walk(func(k []byte, v interface{}) bool {
					store := &stypes.StorageObject{Key: k}
					if v == nil {
						store.Deleted = true
					} else {
						// rlp marshal value here, since snapshot use the same encoding rule.
						//nolint:forcetypeassert
						vv := storeAr.NewBytes(bytes.TrimLeft(v.([]byte), "\x00"))
						store.Val = vv.MarshalTo(nil)
					}

					// update snapshots storage value
					if txn.snap != nil {
						var (
							// current key is slot, we need slot hash
							storeHash = crypto.Keccak256Hash(k)
							storage   map[types.Hash][]byte
							addrHash  = sobj.AddressHash()
						)
						// create map when not exists
						if storage = txn.snapStorage[addrHash]; storage == nil {
							storage = make(map[types.Hash][]byte)
							txn.snapStorage[addrHash] = storage
						}
						// update value. v will be nil if it's deleted
						storage[storeHash] = store.Val
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
