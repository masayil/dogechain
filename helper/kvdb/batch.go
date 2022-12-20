package kvdb

// IdealBatchSize defines the size of the data batches should ideally add in one
// write.
const IdealBatchSize = 100 * 1024

type Batch interface {
	KVWriter

	// // ValueSize retrieves the amount of data queued up for writing.
	// ValueSize() int

	// Write flushes any accumulated data to disk.
	Write() error

	// // Reset resets the batch for reuse.
	// Reset()

	// // Replay replays the batch contents.
	// Replay(w KVWriter) error
}

// Batcher wraps the NewBatch method of a backing data store.
type Batcher interface {
	// NewBatch creates a write-only database that buffers changes to its host db
	// until a final write is called.
	NewBatch() Batch
}
