package kvdb

import "io"

// KVReader wraps the Get method of a backing data store.
type KVReader interface {
	// Get retrieves the given key if it's present in the key-value data store.
	Get(k []byte) ([]byte, bool, error)
}

// KVWriter wraps the Put method of a backing data store.
type KVWriter interface {
	Set(k, v []byte) error
}

// KVBatchStorage is a batch write for leveldb
type KVBatchStorage interface {
	KVReader
	KVWriter
	Batcher
	Iteratee
	io.Closer
}
