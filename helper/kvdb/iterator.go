package kvdb

type IteratorRange struct {
	Start []byte
	Limit []byte
}

type Iterator interface {
	// First moves the iterator to the first key/value pair. If the iterator
	// only contains one key/value pair then First and Last would moves
	// to the same key/value pair.
	// It returns whether such pair exist.
	First() bool

	// Last moves the iterator to the last key/value pair. If the iterator
	// only contains one key/value pair then First and Last would moves
	// to the same key/value pair.
	// It returns whether such pair exist.
	Last() bool

	// Next moves the iterator to the next key/value pair.
	// It returns false if the iterator is exhausted.
	Next() bool

	// Prev moves the iterator to the previous key/value pair.
	// It returns false if the iterator is exhausted.
	Prev() bool

	// Key returns the key of the current key/value pair, or nil if done.
	// The caller should not modify the contents of the returned slice, and
	// its contents may change on the next call to any 'seeks method'.
	Key() []byte

	// Value returns the value of the current key/value pair, or nil if done.
	// The caller should not modify the contents of the returned slice, and
	// its contents may change on the next call to any 'seeks method'.
	Value() []byte

	// Release releases associated resources. Release should always success
	// and can be called multiple times without causing error.
	Release()

	// Error returns any accumulated error. Exhausting all the key/value pairs
	// is not considered to be an error.
	Error() error
}

// Iteratee wraps the NewIterator methods of a backing data store.
type Iteratee interface {
	NewIterator(*IteratorRange) Iterator

	// // NewIterator creates a binary-alphabetical iterator over a subset
	// // of database content with a particular key prefix, starting at a particular
	// // initial key (or after, if it does not exist).
	// //
	// // Note: This method assumes that the prefix is NOT part of the start, so there's
	// // no need for the caller to prepend the prefix to the start
	// NewIterator(prefix []byte, start []byte) Iterator
}
