// Copyright 2019 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package snapshot

import (
	"bytes"
	"math/big"
	"sync"

	"github.com/VictoriaMetrics/fastcache"
	"github.com/dogechain-lab/dogechain/helper/kvdb"
	"github.com/dogechain-lab/dogechain/helper/metrics"
	"github.com/dogechain-lab/dogechain/helper/rawdb"
	"github.com/dogechain-lab/dogechain/helper/rlp"
	"github.com/dogechain-lab/dogechain/state/stypes"
	"github.com/dogechain-lab/dogechain/trie"
	"github.com/dogechain-lab/dogechain/types"
)

// diskLayer is a low level persistent snapshot built on top of a key-value store.
type diskLayer struct {
	diskdb kvdb.KVBatchStorage // Key-value store containing the base snapshot
	triedb *trie.Database      // Trie node cache for reconstruction purposes
	cache  *fastcache.Cache    // Cache to avoid hitting the disk for direct access

	root  types.Hash // Root hash of the base snapshot
	stale bool       // Signals that the layer became stale (state progressed)

	genMarker  []byte                    // Marker for the state that's indexed during initial layer generation
	genPending chan struct{}             // Notification channel when generation is done (test synchronicity)
	genAbort   chan chan *generatorStats // Notification channel to abort generating the snapshot in this layer

	lock sync.RWMutex

	logger      kvdb.Logger
	snapmetrics *Metrics
}

// Root returns  root hash for which this snapshot was made.
func (dl *diskLayer) Root() types.Hash {
	return dl.root
}

// Parent always returns nil as there's no layer below the disk.
func (dl *diskLayer) Parent() snapshot {
	return nil
}

// Stale return whether this layer has become stale (was flattened across) or if
// it's still live.
func (dl *diskLayer) Stale() bool {
	dl.lock.RLock()
	defer dl.lock.RUnlock()

	return dl.stale
}

// Account directly retrieves the account associated with a particular hash in
// the snapshot slim data format.
func (dl *diskLayer) Account(hash types.Hash) (*stypes.Account, error) {
	data, err := dl.AccountRLP(hash)
	if err != nil {
		return nil, err
	}

	if len(data) == 0 { // can be both nil and []byte{}
		return nil, nil
	}

	var slimAccount Account
	if err := rlp.DecodeBytes(data, &slimAccount); err != nil {
		panic(err)
	}

	// copy balance to heap
	balance := new(big.Int).Set(slimAccount.Balance)

	return &stypes.Account{
		Nonce:       slimAccount.Nonce,
		Balance:     balance,
		StorageRoot: types.BytesToHash(slimAccount.Root),
		CodeHash:    slimAccount.CodeHash,
	}, nil
}

// AccountRLP directly retrieves the account RLP associated with a particular
// hash in the snapshot slim data format.
func (dl *diskLayer) AccountRLP(hash types.Hash) ([]byte, error) {
	dl.lock.RLock()
	defer dl.lock.RUnlock()

	// If the layer was flattened into, consider it invalid (any live reference to
	// the original should be marked as unusable).
	if dl.stale {
		return nil, ErrSnapshotStale
	}
	// If the layer is being generated, ensure the requested hash has already been
	// covered by the generator.
	if dl.genMarker != nil && bytes.Compare(hash[:], dl.genMarker) > 0 {
		return nil, ErrNotCoveredYet
	}

	// If we're in the disk layer, all diff layers missed
	metrics.CounterInc(dl.snapmetrics.dirtyAccountMissCount)

	// Try to retrieve the account from the memory cache
	if blob, found := dl.cache.HasGet(nil, hash[:]); found {
		metrics.CounterInc(dl.snapmetrics.cleanAccountHitCount)
		metrics.AddCounter(dl.snapmetrics.cleanAccountReadSize, float64(len(blob)))

		return blob, nil
	}

	// Cache doesn't contain account, pull from disk and cache for later
	blob := rawdb.ReadAccountSnapshot(dl.diskdb, hash)
	dl.cache.Set(hash[:], blob)

	metrics.CounterInc(dl.snapmetrics.cleanAccountMissCount)
	// write or inex
	if n := len(blob); n > 0 {
		metrics.AddCounter(dl.snapmetrics.cleanAccountWriteSize, float64(n))
	} else {
		metrics.CounterInc(dl.snapmetrics.cleanAccountInexCount)
	}

	return blob, nil
}

// Storage directly retrieves the storage data associated with a particular hash,
// within a particular account.
func (dl *diskLayer) Storage(accountHash, storageHash types.Hash) ([]byte, error) {
	dl.lock.RLock()
	defer dl.lock.RUnlock()

	// If the layer was flattened into, consider it invalid (any live reference to
	// the original should be marked as unusable).
	if dl.stale {
		return nil, ErrSnapshotStale
	}

	key := append(accountHash[:], storageHash[:]...)

	// If the layer is being generated, ensure the requested hash has already been
	// covered by the generator.
	if dl.genMarker != nil && bytes.Compare(key, dl.genMarker) > 0 {
		return nil, ErrNotCoveredYet
	}

	// If we're in the disk layer, all diff layers missed
	metrics.CounterInc(dl.snapmetrics.dirtyStorageMissCount)

	// Try to retrieve the storage slot from the memory cache
	if blob, found := dl.cache.HasGet(nil, key); found {
		metrics.CounterInc(dl.snapmetrics.cleanStorageHitCount)
		metrics.AddCounter(dl.snapmetrics.cleanStorageReadSize, float64(len(blob)))

		return blob, nil
	}
	// Cache doesn't contain storage slot, pull from disk and cache for later
	blob := rawdb.ReadStorageSnapshot(dl.diskdb, accountHash, storageHash)
	dl.cache.Set(key, blob)

	metrics.CounterInc(dl.snapmetrics.cleanStorageMissCount)
	// write or inex
	if n := len(blob); n > 0 {
		metrics.AddCounter(dl.snapmetrics.cleanStorageWriteSize, float64(n))
	} else {
		metrics.CounterInc(dl.snapmetrics.cleanStorageInexCount)
	}

	return blob, nil
}

// Update creates a new layer on top of the existing snapshot diff tree with
// the specified data items. Note, the maps are retained by the method to avoid
// copying everything.
func (dl *diskLayer) Update(
	blockHash types.Hash,
	destructs map[types.Hash]struct{},
	accounts map[types.Hash][]byte,
	storage map[types.Hash]map[types.Hash][]byte,
	logger kvdb.Logger,
) *diffLayer {
	return newDiffLayer(dl, blockHash, destructs, accounts, storage, logger, dl.snapmetrics)
}
