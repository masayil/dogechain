package itrie

import (
	"encoding/hex"
	"errors"
	"fmt"
	"sync"

	"github.com/dogechain-lab/dogechain/state"
	"github.com/dogechain-lab/dogechain/state/schema"
	"github.com/dogechain-lab/dogechain/types"
	"go.uber.org/atomic"
)

type StateDBTransaction interface {
	StateDBReader
	StorageWriter

	GetCode(hash types.Hash) ([]byte, bool)
	SetCode(hash types.Hash, code []byte) error

	Commit() error
	Rollback()
}

type txnKey string
type txnPair struct {
	key    []byte
	value  []byte
	isCode bool
}

var txnPairPool = sync.Pool{
	New: func() interface{} {
		return &txnPair{
			key:    make([]byte, 0),
			value:  make([]byte, 0),
			isCode: false,
		}
	},
}

func (pair *txnPair) Reset() {
	pair.key = pair.key[:0]
	pair.value = pair.value[:0]
	pair.isCode = false
}

type stateDBTxn struct {
	db   map[txnKey]*txnPair
	lock sync.Mutex // for protecting map

	stateDB StateDB
	storage Storage

	cancel *atomic.Bool
}

func (tx *stateDBTxn) Set(k []byte, v []byte) error {
	key := byteKeyToTxnKey(k)

	pair, ok := txnPairPool.Get().(*txnPair)
	if !ok {
		return errors.New("invalid type assertion")
	}

	pair.key = append(pair.key[:], k...)
	pair.value = append(pair.value[:], v...)

	tx.lock.Lock()
	defer tx.lock.Unlock()

	tx.db[key] = pair

	return nil
}

func (tx *stateDBTxn) Delete(k []byte) error {
	key := byteKeyToTxnKey(k)

	tx.lock.Lock()
	defer tx.lock.Unlock()

	delete(tx.db, key)

	return nil
}

func byteKeyToTxnKey(k []byte) txnKey {
	return txnKey(hex.EncodeToString(k))
}

// func (tx *stateDBTxn) Has(k []byte) (bool, error) {
// 	key := byteKeyToTxnKey(k)

// 	tx.lock.Lock()

// 	if _, ok := tx.db[key]; ok {
// 		tx.lock.Unlock()

// 		return true, nil
// 	}

// 	tx.lock.Unlock()

// 	return tx.stateDB.Has(k)
// }

func (tx *stateDBTxn) Get(k []byte) ([]byte, bool, error) {
	key := byteKeyToTxnKey(k)

	tx.lock.Lock()

	v, ok := tx.db[key]
	if ok {
		// copy value
		bufValue := make([]byte, len(v.value))
		copy(bufValue[:], v.value[:])

		// unlock
		tx.lock.Unlock()

		return bufValue, true, nil
	}

	tx.lock.Unlock()

	return tx.stateDB.Get(k)
}

func (tx *stateDBTxn) SetCode(hash types.Hash, v []byte) error {
	// active code key is different from account key (hash)
	key := schema.CodeKey(hash)
	keyStr := byteKeyToTxnKey(key)

	pair, ok := txnPairPool.Get().(*txnPair)
	if !ok {
		return errors.New("invalid type assertion")
	}

	pair.key = append(pair.key[:0], key...)
	pair.value = append(pair.value[:0], v...)
	pair.isCode = true

	tx.lock.Lock()
	defer tx.lock.Unlock()

	tx.db[keyStr] = pair

	return nil
}

func (tx *stateDBTxn) GetCode(hash types.Hash) ([]byte, bool) {
	key := byteKeyToTxnKey(schema.CodeKey(hash))

	tx.lock.Lock()

	if v, ok := tx.db[key]; ok {
		// depth copy
		bufValue := make([]byte, len(v.value))
		copy(bufValue[:], v.value[:])

		// unlock
		tx.lock.Unlock()

		return bufValue, true
	}

	tx.lock.Unlock()

	return tx.stateDB.GetCode(hash)
}

func (tx *stateDBTxn) NewSnapshot() state.Snapshot {
	return tx.stateDB.NewSnapshot()
}

func (tx *stateDBTxn) NewSnapshotAt(root types.Hash) (state.Snapshot, error) {
	if root == types.EmptyRootHash {
		// empty state
		return tx.NewSnapshot(), nil
	}

	// user exclusive transaction to get state
	// use non-commit state
	n, ok, err := GetNode(root.Bytes(), tx)

	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, fmt.Errorf("state not found at hash %s", root)
	}

	t := &Trie{
		root:    n,
		stateDB: tx.stateDB,
	}

	return t, nil
}

func (tx *stateDBTxn) Commit() error {
	if tx.cancel.Load() {
		return ErrStateTransactionIsCancel
	}

	tx.lock.Lock()

	// double check
	if tx.cancel.Load() {
		tx.lock.Unlock()

		return ErrStateTransactionIsCancel
	}

	batch := tx.storage.NewBatch()
	metrics := tx.stateDB.GetMetrics()

	for _, pair := range tx.db {
		err := batch.Set(pair.key, pair.value)

		if err != nil {
			tx.lock.Unlock()

			return err
		}

		if !pair.isCode {
			metrics.transactionWriteNodeSize(len(pair.value))
		}
	}

	tx.lock.Unlock()

	return batch.Write()
}

// clear transaction data, set cancel flag
func (tx *stateDBTxn) Rollback() {
	// cancle by atomic swap value
	if alreadyCancel := tx.cancel.Swap(true); alreadyCancel {
		return
	}

	tx.clear()
}

func (tx *stateDBTxn) clear() {
	tx.stateDB = nil
	tx.storage = nil

	tx.lock.Lock()
	defer tx.lock.Unlock()

	for tk := range tx.db {
		pair := tx.db[tk]
		pair.Reset()

		txnPairPool.Put(pair)
		delete(tx.db, tk)
	}
}
