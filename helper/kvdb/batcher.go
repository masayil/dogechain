package kvdb

type Batch interface {
	// Set inserts the given value into the key-value data store.
	Set(k, v []byte)
	// Write flushes any accumulated data to disk.
	Write() error
}

// Batcher wraps the NewBatch method of a backing data store.
type Batcher interface {
	// NewBatch creates a write-only database that buffers changes to its host db
	// until a final write is called.
	NewBatch() Batch
}
