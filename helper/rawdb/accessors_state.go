// Copyright 2020 The go-ethereum Authors
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
	"github.com/dogechain-lab/dogechain/helper/kvdb"
	"github.com/dogechain-lab/dogechain/types"
)

// ReadCode retrieves the contract code of the provided code hash.
func ReadCode(db kvdb.KVReader, hash types.Hash) []byte {
	// Try with the prefixed code scheme first, if not then try with legacy
	// scheme.
	data := ReadCodeWithPrefix(db, hash)
	if len(data) != 0 {
		return data
	}

	data, _, _ = db.Get(hash.Bytes())

	return data
}

// ReadCodeWithPrefix retrieves the contract code of the provided code hash.
// The main difference between this function and ReadCode is this function
// will only check the existence with latest scheme(with prefix).
func ReadCodeWithPrefix(db kvdb.KVReader, hash types.Hash) []byte {
	data, _, _ := db.Get(codeKey(hash))

	return data
}

// ReadTrieNode retrieves the trie node of the provided hash.
func ReadTrieNode(db kvdb.KVReader, hash types.Hash) []byte {
	data, _, _ := db.Get(hash.Bytes())

	return data
}

// HasCode checks if the contract code corresponding to the
// provided code hash is present in the db.
func HasCode(db kvdb.KVReader, hash types.Hash) bool {
	// Try with the prefixed code scheme first, if not then try with legacy
	// scheme.
	if ok := HasCodeWithPrefix(db, hash); ok {
		return true
	}

	ok, _ := db.Has(hash.Bytes())

	return ok
}

// HasCodeWithPrefix checks if the contract code corresponding to the
// provided code hash is present in the db. This function will only check
// presence using the prefix-scheme.
func HasCodeWithPrefix(db kvdb.KVReader, hash types.Hash) bool {
	ok, _ := db.Has(codeKey(hash))

	return ok
}

// HasTrieNode checks if the trie node with the provided hash is present in db.
func HasTrieNode(db kvdb.KVReader, hash types.Hash) bool {
	ok, _ := db.Has(hash.Bytes())

	return ok
}

// WriteCode writes the provided contract code database.
func WriteCode(db kvdb.KVWriter, hash types.Hash, code []byte) error {
	return db.Set(codeKey(hash), code)
}

// WriteTrieNode writes the provided trie node database.
func WriteTrieNode(db kvdb.KVWriter, hash types.Hash, node []byte) error {
	return db.Set(hash.Bytes(), node)
}

// DeleteCode deletes the specified contract code from the database.
func DeleteCode(db kvdb.KVWriter, hash types.Hash) error {
	return db.Delete(codeKey(hash))
}

// DeleteTrieNode deletes the specified trie node from the database.
func DeleteTrieNode(db kvdb.KVWriter, hash types.Hash) error {
	return db.Delete(hash.Bytes())
}
