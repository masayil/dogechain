// Copyright 2018 The go-ethereum Authors
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
	"github.com/dogechain-lab/dogechain/helper/kvdb/memorydb"
)

// nofreezedb is a database wrapper that disables freezer data retrievals.
type nofreezedb struct {
	kvdb.KVBatchStorage
}

// NewDatabase creates a high level database on top of a given key-value data
// store without a freezer moving immutable chain segments into cold storage.
func NewDatabase(db kvdb.KVBatchStorage) kvdb.Database {
	return &nofreezedb{KVBatchStorage: db}
}

// NewMemoryDatabase creates an ephemeral in-memory key-value database without a
// freezer moving immutable chain segments into cold storage.
func NewMemoryDatabase() kvdb.Database {
	return NewDatabase(memorydb.New())
}

// NewMemoryDatabaseWithCap creates an ephemeral in-memory key-value database
// with an initial starting capacity, but without a freezer moving immutable
// chain segments into cold storage.
func NewMemoryDatabaseWithCap(size int) kvdb.Database {
	return NewDatabase(memorydb.NewWithCap(size))
}

// // NewLevelDBDatabase creates a persistent key-value database without a freezer
// // moving immutable chain segments into cold storage.
// func NewLevelDBDatabase(
// 	file string,
// 	cache int,
// 	handles int,
// 	logger hclog.Logger,
// 	namespace string,
// 	readonly bool,
// ) (kvdb.Database, error) {
// 	if logger == nil {
// 		logger = hclog.NewNullLogger()
// 	}

// 	db, err := leveldb.New(file,
// 		leveldb.SetBloomKeyBits(),
// 	)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return NewDatabase(db), nil
// }

// // NewLevelDBDatabaseWithFreezer creates a persistent key-value database with a
// // freezer moving immutable chain segments into cold storage. The passed ancient
// // indicates the path of root ancient directory where the chain freezer can be
// // opened.
// func NewLevelDBDatabaseWithFreezer(
// 	file string,
// 	cache int,
// 	handles int,
// 	ancient string,
// 	namespace string,
// 	readonly bool,
// ) (kvdb.Database, error) {
// 	kvdb, err := leveldb.New(file, cache, handles, namespace, readonly)
// 	if err != nil {
// 		return nil, err
// 	}
// 	frdb, err := NewDatabaseWithFreezer(kvdb, ancient, namespace, readonly)
// 	if err != nil {
// 		kvdb.Close()
// 		return nil, err
// 	}
// 	return frdb, nil
// }
