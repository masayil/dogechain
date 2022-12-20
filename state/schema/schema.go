package schema

import "github.com/dogechain-lab/dogechain/types"

var (
	// codePrefix is the code prefix for leveldb
	CodePrefix = []byte("code")
	// SnapshotAccountPrefix + account hash -> account trie value
	SnapshotAccountPrefix = []byte("sa")
	// SnapshotStoragePrefix + account hash + storage hash -> storage trie value
	SnapshotStoragePrefix = []byte("ss")

	// SnapshotDisabledKey flags that the snapshot should not be maintained due to initial sync.
	SnapshotDisabledKey = []byte("SnapshotDisabled")

	// SnapshotRootKey tracks the hash of the last snapshot.
	SnapshotRootKey = []byte("SnapshotRoot")

	// SnapshotJournalKey tracks the in-memory diff layers across restarts.
	SnapshotJournalKey = []byte("SnapshotJournal")

	// SnapshotGeneratorKey tracks the snapshot generation marker across restarts.
	SnapshotGeneratorKey = []byte("SnapshotGenerator")

	// SnapshotRecoveryKey tracks the snapshot recovery marker across restarts.
	SnapshotRecoveryKey = []byte("SnapshotRecovery")

	// SnapshotSyncStatusKey tracks the snapshot sync status across restarts.
	SnapshotSyncStatusKey = []byte("SnapshotSyncStatus")

	// SkeletonSyncStatusKey tracks the skeleton sync status across restarts.
	SkeletonSyncStatusKey = []byte("SkeletonSyncStatus")
)

// CodeKey = CodePrefix + hash
func CodeKey(hash types.Hash) []byte {
	return append(CodePrefix, hash.Bytes()...)
}

// SnapshotAccountKey = SnapshotAccountPrefix + hash
func SnapshotAccountKey(hash types.Hash) []byte {
	return append(SnapshotAccountPrefix, hash.Bytes()...)
}

// SnapshotStorageKey = SnapshotStoragePrefix + account hash + storage hash
func SnapshotStorageKey(accountHash, storageHash types.Hash) []byte {
	return append(append(SnapshotStoragePrefix, accountHash.Bytes()...), storageHash.Bytes()...)
}

// SnapshotsStorageKey = SnapshotStoragePrefix + account hash (+ storage hash)
func SnapshotsStorageKey(accountHash types.Hash) []byte {
	return append(SnapshotStoragePrefix, accountHash.Bytes()...)
}
