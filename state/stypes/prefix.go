package stypes

import "github.com/dogechain-lab/dogechain/types"

var (
	SnapshotAccountPrefix = []byte("a") // SnapshotAccountPrefix + account hash -> account trie value
	SnapshotStoragePrefix = []byte("o") // SnapshotStoragePrefix + account hash + storage hash -> storage trie value
)

func SnapshotStorageKey(account types.Address) []byte {
	return append(SnapshotStoragePrefix, account.Bytes()...)
}
