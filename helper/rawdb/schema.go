package rawdb

import (
	"encoding/binary"

	"github.com/dogechain-lab/dogechain/types"
)

const (
	SnapshotPrefixLength = 2 // "s*" s means snapshot, * is the first letter of the subkey
)

// snapshot key prefix
var (
	// codePrefix is the code prefix for leveldb
	codePrefix = []byte("code")
	// SnapshotAccountPrefix + account hash -> account trie value
	SnapshotAccountPrefix = []byte("sa")
	// SnapshotStoragePrefix + account hash + storage hash -> storage trie value
	SnapshotStoragePrefix = []byte("ss")
)

// snapshot key
var (
	// snapshotDisabledKey flags that the snapshot should not be maintained due to initial sync.
	snapshotDisabledKey = []byte("SnapshotDisabled")
	// snapshotRootKey tracks the hash of the last snapshot.
	snapshotRootKey = []byte("SnapshotRoot")
	// snapshotJournalKey tracks the in-memory diff layers across restarts.
	snapshotJournalKey = []byte("SnapshotJournal")
	// snapshotGeneratorKey tracks the snapshot generation marker across restarts.
	snapshotGeneratorKey = []byte("SnapshotGenerator")
	// snapshotRecoveryKey tracks the snapshot recovery marker across restarts.
	snapshotRecoveryKey = []byte("SnapshotRecovery")
	// snapshotSyncStatusKey tracks the snapshot sync status across restarts.
	snapshotSyncStatusKey = []byte("SnapshotSyncStatus")
	// skeletonSyncStatusKey tracks the skeleton sync status across restarts.
	skeletonSyncStatusKey = []byte("SkeletonSyncStatus")
)

// blockchain key prefix
var (
	// bodyPrefix is the prefix for bodies
	bodyPrefix = []byte("b")
	// canonicalPrefix is the prefix for the canonical chain numbers
	canonicalPrefix = []byte("c")
	// difficultyPrefix is the difficulty prefix
	difficultyPrefix = []byte("d")
	// headerPrefix is the header prefix
	headerPrefix = []byte("h")
	// receiptsPrefix is the prefix for receipts
	receiptsPrefix = []byte("r")
	// txLookupPrefix is the prefix for transaction lookups
	txLookupPrefix = []byte("l")
)

// blockchain keys
var (
	// headHashKey tracks the latest known header's hash
	headHashKey = []byte("ohash")
	// headNumberKey tracks the latest known header's number
	headNumberKey = []byte("onumber")
	// forkEmptyKey tracks any fork which never exists
	forkEmptyKey = []byte("empty")
)

func encodeUint(n uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b[:], n)

	return b[:]
}

func decodeUint(b []byte) uint64 {
	return binary.BigEndian.Uint64(b[:])
}

// CodeKey = CodePrefix + hash
func CodeKey(hash types.Hash) []byte {
	return append(codePrefix, hash.Bytes()...)
}

// snapshotAccountKey = SnapshotAccountPrefix + hash
func snapshotAccountKey(hash types.Hash) []byte {
	return append(SnapshotAccountPrefix, hash.Bytes()...)
}

// snapshotStorageKey = SnapshotStoragePrefix + account hash + storage hash
func snapshotStorageKey(accountHash, storageHash types.Hash) []byte {
	return append(append(SnapshotStoragePrefix, accountHash.Bytes()...), storageHash.Bytes()...)
}

// SnapshotsStorageKey = SnapshotStoragePrefix + account hash (+ storage hash)
func SnapshotsStorageKey(accountHash types.Hash) []byte {
	return append(SnapshotStoragePrefix, accountHash.Bytes()...)
}

func bodyKey(h types.Hash) []byte {
	return append(bodyPrefix, h.Bytes()...)
}

func canonicalHashKey(n uint64) []byte {
	return append(canonicalPrefix, encodeUint(n)...)
}

func difficultyKey(h types.Hash) []byte {
	return append(difficultyPrefix, h.Bytes()...)
}

func headerKey(h types.Hash) []byte {
	return append(headerPrefix, h.Bytes()...)
}

func receiptsKey(h types.Hash) []byte {
	return append(receiptsPrefix, h.Bytes()...)
}

func txLookupKey(h types.Hash) []byte {
	return append(txLookupPrefix, h.Bytes()...)
}
