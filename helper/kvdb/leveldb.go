package kvdb

import (
	"errors"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type levelBatch struct {
	db    *leveldb.DB
	batch *leveldb.Batch
}

func (b *levelBatch) Set(k, v []byte) {
	b.batch.Put(k, v)
}

func (b *levelBatch) Write() error {
	return b.db.Write(b.batch, nil)
}

// levelDBKV is the leveldb implementation of the kv storage
type levelDBKV struct {
	db *leveldb.DB
}

func (kv *levelDBKV) NewBatch() Batch {
	return &levelBatch{db: kv.db, batch: &leveldb.Batch{}}
}

// bytesPrefixRange returns key range that satisfy
// - the given prefix, and
// - the given seek position
func bytesPrefixRange(prefix, start, limit []byte) *util.Range {
	r := util.BytesPrefix(prefix)
	r.Start = append(r.Start, start...)
	r.Limit = limit

	return r
}

func (kv *levelDBKV) NewIterator2(prefix, start, limit []byte) Iterator {
	return kv.db.NewIterator(bytesPrefixRange(prefix, start, limit), nil)
}

func (kv *levelDBKV) NewIterator(Range *IteratorRange) Iterator {
	if Range == nil {
		return kv.db.NewIterator(nil, nil)
	}

	return kv.db.NewIterator(bytesPrefixRange(nil, Range.Start, Range.Limit), nil)
}

// Set sets the key-value pair in leveldb storage
func (kv *levelDBKV) Set(p []byte, v []byte) error {
	return kv.db.Put(p, v, nil)
}

// Get retrieves the key-value pair in leveldb storage
func (kv *levelDBKV) Get(p []byte) ([]byte, bool, error) {
	data, err := kv.db.Get(p, nil)
	if err != nil {
		if errors.Is(err, leveldb.ErrNotFound) {
			return nil, false, nil
		} else if errors.Is(err, leveldb.ErrClosed) {
			return nil, false, nil
		} else {
			panic(err)
		}
	}

	return data, true, nil
}

// Close closes the leveldb storage instance
func (kv *levelDBKV) Close() error {
	return kv.db.Close()
}
