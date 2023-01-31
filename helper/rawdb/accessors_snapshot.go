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

package rawdb

import (
	"encoding/binary"
	"log"

	"github.com/dogechain-lab/dogechain/helper/kvdb"
	"github.com/dogechain-lab/dogechain/state/schema"
	"github.com/dogechain-lab/dogechain/types"
)

// ReadSnapshotDisabled retrieves if the snapshot maintenance is disabled.
func ReadSnapshotDisabled(db kvdb.KVReader) bool {
	disabled, _ := db.Has(schema.SnapshotDisabledKey)

	return disabled
}

// WriteSnapshotDisabled stores the snapshot pause flag.
func WriteSnapshotDisabled(db kvdb.KVWriter) {
	if err := db.Set(schema.SnapshotDisabledKey, []byte("42")); err != nil {
		logCrit("Failed to store snapshot disabled flag", "err", err)
	}
}

// DeleteSnapshotDisabled deletes the flag keeping the snapshot maintenance disabled.
func DeleteSnapshotDisabled(db kvdb.KVWriter) {
	if err := db.Delete(schema.SnapshotDisabledKey); err != nil {
		logCrit("Failed to remove snapshot disabled flag", "err", err)
	}
}

// ReadSnapshotRoot retrieves the root of the block whose state is contained in
// the persisted snapshot.
func ReadSnapshotRoot(db kvdb.KVReader) types.Hash {
	data, _, _ := db.Get(schema.SnapshotRootKey)
	if len(data) != types.HashLength {
		return types.Hash{}
	}

	return types.BytesToHash(data)
}

// WriteSnapshotRoot stores the root of the block whose state is contained in
// the persisted snapshot.
func WriteSnapshotRoot(db kvdb.KVWriter, root types.Hash) {
	if err := db.Set(schema.SnapshotRootKey, root[:]); err != nil {
		logCrit("Failed to store snapshot root", "err", err)
	}
}

// DeleteSnapshotRoot deletes the hash of the block whose state is contained in
// the persisted snapshot. Since snapshots are not immutable, this  method can
// be used during updates, so a crash or failure will mark the entire snapshot
// invalid.
func DeleteSnapshotRoot(db kvdb.KVWriter) {
	if err := db.Delete(schema.SnapshotRootKey); err != nil {
		logCrit("Failed to remove snapshot root", "err", err)
	}
}

// ReadAccountSnapshot retrieves the snapshot entry of an account trie leaf.
func ReadAccountSnapshot(db kvdb.KVReader, hash types.Hash) []byte {
	data, _, _ := db.Get(schema.SnapshotAccountKey(hash))

	return data
}

// WriteAccountSnapshot stores the snapshot entry of an account trie leaf.
func WriteAccountSnapshot(db kvdb.KVWriter, hash types.Hash, entry []byte) {
	if err := db.Set(schema.SnapshotAccountKey(hash), entry); err != nil {
		logCrit("Failed to store account snapshot", "err", err)
	}
}

// DeleteAccountSnapshot removes the snapshot entry of an account trie leaf.
func DeleteAccountSnapshot(db kvdb.KVWriter, hash types.Hash) {
	if err := db.Delete(schema.SnapshotAccountKey(hash)); err != nil {
		logCrit("Failed to delete account snapshot", "err", err)
	}
}

// ReadStorageSnapshot retrieves the snapshot entry of an storage trie leaf.
func ReadStorageSnapshot(db kvdb.KVReader, accountHash, storageHash types.Hash) []byte {
	data, _, _ := db.Get(schema.SnapshotStorageKey(accountHash, storageHash))

	return data
}

// WriteStorageSnapshot stores the snapshot entry of an storage trie leaf.
func WriteStorageSnapshot(db kvdb.KVWriter, accountHash, storageHash types.Hash, entry []byte) {
	if err := db.Set(schema.SnapshotStorageKey(accountHash, storageHash), entry); err != nil {
		logCrit("Failed to store storage snapshot", "err", err)
	}
}

// DeleteStorageSnapshot removes the snapshot entry of an storage trie leaf.
func DeleteStorageSnapshot(db kvdb.KVWriter, accountHash, storageHash types.Hash) {
	if err := db.Delete(schema.SnapshotStorageKey(accountHash, storageHash)); err != nil {
		logCrit("Failed to delete storage snapshot", "err", err)
	}
}

// IterateStorageSnapshots returns an iterator for walking the entire storage
// space of a specific account.
func IterateStorageSnapshots(db kvdb.Iteratee, accountHash types.Hash) kvdb.Iterator {
	return NewKeyLengthIterator(
		db.NewIterator(schema.SnapshotsStorageKey(accountHash), nil),
		len(schema.SnapshotStoragePrefix)+2*types.HashLength,
	)
}

// ReadSnapshotJournal retrieves the serialized in-memory diff layers saved at
// the last shutdown. The blob is expected to be max a few 10s of megabytes.
func ReadSnapshotJournal(db kvdb.KVReader) []byte {
	data, _, _ := db.Get(schema.SnapshotJournalKey)

	return data
}

// WriteSnapshotJournal stores the serialized in-memory diff layers to save at
// shutdown. The blob is expected to be max a few 10s of megabytes.
func WriteSnapshotJournal(db kvdb.KVWriter, journal []byte) {
	if err := db.Set(schema.SnapshotJournalKey, journal); err != nil {
		logCrit("Failed to store snapshot journal", "err", err)
	}
}

// DeleteSnapshotJournal deletes the serialized in-memory diff layers saved at
// the last shutdown
func DeleteSnapshotJournal(db kvdb.KVWriter) {
	if err := db.Delete(schema.SnapshotJournalKey); err != nil {
		logCrit("Failed to remove snapshot journal", "err", err)
	}
}

// ReadSnapshotGenerator retrieves the serialized snapshot generator saved at
// the last shutdown.
func ReadSnapshotGenerator(db kvdb.KVReader) []byte {
	data, _, _ := db.Get(schema.SnapshotGeneratorKey)

	return data
}

// WriteSnapshotGenerator stores the serialized snapshot generator to save at
// shutdown.
func WriteSnapshotGenerator(db kvdb.KVWriter, generator []byte) {
	if err := db.Set(schema.SnapshotGeneratorKey, generator); err != nil {
		logCrit("Failed to store snapshot generator", "err", err)
	}
}

// DeleteSnapshotGenerator deletes the serialized snapshot generator saved at
// the last shutdown
func DeleteSnapshotGenerator(db kvdb.KVWriter) {
	if err := db.Delete(schema.SnapshotGeneratorKey); err != nil {
		logCrit("Failed to remove snapshot generator", "err", err)
	}
}

// ReadSnapshotRecoveryNumber retrieves the block number of the last persisted
// snapshot layer.
func ReadSnapshotRecoveryNumber(db kvdb.KVReader) *uint64 {
	data, _, _ := db.Get(schema.SnapshotRecoveryKey)
	if len(data) == 0 {
		return nil
	}

	if len(data) != 8 {
		return nil
	}

	number := binary.BigEndian.Uint64(data)

	return &number
}

// WriteSnapshotRecoveryNumber stores the block number of the last persisted
// snapshot layer.
func WriteSnapshotRecoveryNumber(db kvdb.KVWriter, number uint64) {
	var buf [8]byte

	binary.BigEndian.PutUint64(buf[:], number)

	if err := db.Set(schema.SnapshotRecoveryKey, buf[:]); err != nil {
		logCrit("Failed to store snapshot recovery number", "err", err)
	}
}

// DeleteSnapshotRecoveryNumber deletes the block number of the last persisted
// snapshot layer.
func DeleteSnapshotRecoveryNumber(db kvdb.KVWriter) {
	if err := db.Delete(schema.SnapshotRecoveryKey); err != nil {
		logCrit("Failed to remove snapshot recovery number", "err", err)
	}
}

// ReadSnapshotSyncStatus retrieves the serialized sync status saved at shutdown.
func ReadSnapshotSyncStatus(db kvdb.KVReader) []byte {
	data, _, _ := db.Get(schema.SnapshotSyncStatusKey)

	return data
}

// WriteSnapshotSyncStatus stores the serialized sync status to save at shutdown.
func WriteSnapshotSyncStatus(db kvdb.KVWriter, status []byte) {
	if err := db.Set(schema.SnapshotSyncStatusKey, status); err != nil {
		logCrit("Failed to store snapshot sync status", "err", err)
	}
}

func logCrit(msg string, args ...interface{}) {
	log.Fatal(msg, args)
}
