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
	"sync"
	"sync/atomic"

	"github.com/dogechain-lab/dogechain/state/stypes"
	"github.com/dogechain-lab/dogechain/types"
)

// diffLayer represents a collection of modifications made to a state snapshot
// after running a block on top. It contains one sorted list for the account trie
// and one-one list for each storage tries.
//
// The goal of a diff layer is to act as a journal, tracking recent modifications
// made to the state, that have not yet graduated into a semi-immutable state.
type diffLayer struct {
	origin *diskLayer // Base disk layer to directly use on bloom misses
	parent snapshot   // Parent snapshot modified by this one, never nil
	memory uint64     // Approximate guess as to how much memory we use

	root  types.Hash // Root hash to which this snapshot diff belongs to
	stale uint32     // Signals that the layer became stale (state progressed)

	// destructSet is a very special helper marker. If an account is marked as
	// deleted, then it's recorded in this set. However it's allowed that an account
	// is included here but still available in other sets(e.g. storageData). The
	// reason is the diff layer includes all the changes in a *block*. It can
	// happen that in the tx_1, account A is self-destructed while in the tx_2
	// it's recreated. But we still need this marker to indicate the "old" A is
	// deleted, all data in other set belongs to the "new" A.
	// Keyed markers for deleted (and potentially) recreated accounts
	destructSet map[types.Hash]struct{}
	// List of account for iteration. If it exists, it's sorted, otherwise it's nil
	accountList []types.Hash
	// Keyed accounts for direct retrieval (nil means deleted)
	accountData map[types.Hash][]byte
	// List of storage slots for iterated retrievals, one per account. Any existing lists are sorted if non-nil
	storageList map[types.Hash][]types.Hash
	// Keyed storage slots for direct retrieval. one per account (nil means deleted)
	storageData map[types.Hash]map[types.Hash][]byte

	// diffed *bloomfilter.Filter // Bloom filter tracking all the diffed items up to the disk layer

	lock sync.RWMutex
}

// newDiffLayer creates a new diff on top of an existing snapshot, whether that's a low
// level persistent database or a hierarchical diff already.
func newDiffLayer(
	parent snapshot,
	root types.Hash,
	destructs map[types.Hash]struct{},
	accounts map[types.Hash][]byte,
	storage map[types.Hash]map[types.Hash][]byte,
) *diffLayer {
	return nil
}

// Root returns the root hash for which this snapshot was made.
func (dl *diffLayer) Root() types.Hash {
	return dl.root
}

// Parent returns the subsequent layer of a diff layer.
func (dl *diffLayer) Parent() snapshot {
	dl.lock.RLock()
	defer dl.lock.RUnlock()

	return dl.parent
}

// Stale return whether this layer has become stale (was flattened across) or if
// it's still live.
func (dl *diffLayer) Stale() bool {
	return atomic.LoadUint32(&dl.stale) != 0
}

// Account directly retrieves the account associated with a particular hash in
// the snapshot slim data format.
func (dl *diffLayer) Account(hash types.Hash) (*stypes.Account, error) {
	return nil, nil
}

// AccountRLP directly retrieves the account RLP associated with a particular
// hash in the snapshot slim data format.
//
// Note the returned account is not a copy, please don't modify it.
func (dl *diffLayer) AccountRLP(hash types.Hash) ([]byte, error) {
	dl.lock.RLock()
	defer dl.lock.RUnlock()

	return nil, nil
}

// Storage directly retrieves the storage data associated with a particular hash,
// within a particular account. If the slot is unknown to this diff, it's parent
// is consulted.
//
// Note the returned slot is not a copy, please don't modify it.
func (dl *diffLayer) Storage(accountHash, storageHash types.Hash) ([]byte, error) {
	// Check the bloom filter first whether there's even a point in reaching into
	// all the maps in all the layers below
	dl.lock.RLock()
	defer dl.lock.RUnlock()

	return nil, nil
}

// Update creates a new layer on top of the existing snapshot diff tree with
// the specified data items.
func (dl *diffLayer) Update(
	blockRoot types.Hash,
	destructs map[types.Hash]struct{},
	accounts map[types.Hash][]byte,
	storage map[types.Hash]map[types.Hash][]byte,
) *diffLayer {
	return newDiffLayer(dl, blockRoot, destructs, accounts, storage)
}

// AccountList returns a sorted list of all accounts in this diffLayer, including
// the deleted ones.
//
// Note, the returned slice is not a copy, so do not modify it.
func (dl *diffLayer) AccountList() []types.Hash {
	return nil
}

// StorageList returns a sorted list of all storage slot hashes in this diffLayer
// for the given account. If the whole storage is destructed in this layer, then
// an additional flag *destructed = true* will be returned, otherwise the flag is
// false. Besides, the returned list will include the hash of deleted storage slot.
// Note a special case is an account is deleted in a prior tx but is recreated in
// the following tx with some storage slots set. In this case the returned list is
// not empty but the flag is true.
//
// Note, the returned slice is not a copy, so do not modify it.
func (dl *diffLayer) StorageList(accountHash types.Hash) ([]types.Hash, bool) {
	dl.lock.RLock()
	defer dl.lock.RUnlock()

	return nil, false
}
