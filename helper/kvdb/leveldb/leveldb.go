package leveldb

import (
	"errors"
	"fmt"

	"github.com/dogechain-lab/dogechain/helper/kvdb"
	"github.com/hashicorp/go-hclog"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

const (
	// minCache is the minimum memory allocate to leveldb
	// half write, half read
	minCache = 16 // 16 MiB

	// minHandles is the minimum number of files handles to leveldb open files
	minHandles = 16

	DefaultCache               = 1024  // 1 GiB
	DefaultHandles             = 512   // files handles to leveldb open files
	DefaultBloomKeyBits        = 2048  // bloom filter bits (256 bytes)
	DefaultCompactionTableSize = 4     // 4  MiB
	DefaultCompactionTotalSize = 40    // 40 MiB
	DefaultNoSyncFlag          = false // false - sync write, true - async write
)

type batch struct {
	db    *leveldb.DB
	batch *leveldb.Batch
	size  int // counting batch size
}

func (b *batch) Set(k, v []byte) error {
	b.batch.Put(k, v)
	b.size += len(k) + len(v)

	return nil
}

func (b *batch) Delete(k []byte) error {
	b.batch.Delete(k)
	b.size += len(k)

	return nil
}

// // ValueSize retrieves the amount of data queued up for writing.
// func (b *batch) ValueSize() int {
// 	return b.size
// }

func (b *batch) Write() error {
	return b.db.Write(b.batch, nil)
}

// // Reset resets the batch for reuse.
// func (b *batch) Reset() {
// 	b.batch.Reset()
// 	b.size = 0
// }

// // Replay replays the batch contents.
// func (b *batch) Replay(w kvdb.KVWriter) error {
// 	return b.batch.Replay(&replayer{writer: w})
// }

// // replayer is a small wrapper to implement the correct replay methods.
// type replayer struct {
// 	writer  kvdb.KVWriter
// 	failure error
// }

// // Put inserts the given value into the key-value data store.
// func (r *replayer) Put(key, value []byte) {
// 	// If the replay already failed, stop executing ops
// 	if r.failure != nil {
// 		return
// 	}

// 	r.failure = r.writer.Set(key, value)
// }

// // Delete removes the key from the key-value data store.
// func (r *replayer) Delete(key []byte) {
// 	// If the replay already failed, stop executing ops
// 	if r.failure != nil {
// 		return
// 	}

// 	r.failure = r.writer.Delete(key)
// }

// database is the leveldb implementation of the kv storage
type database struct {
	db *leveldb.DB
}

func (kv *database) NewBatch() kvdb.Batch {
	return &batch{db: kv.db, batch: &leveldb.Batch{}}
}

// bytesPrefixRange returns key range that satisfy
// - the given prefix, and
// - the given seek position
func bytesPrefixRange(prefix, start []byte) *util.Range {
	r := util.BytesPrefix(prefix)
	r.Start = append(r.Start, start...)

	return r
}

func (kv *database) NewIterator(prefix, start []byte) kvdb.Iterator {
	return kv.db.NewIterator(bytesPrefixRange(prefix, start), nil)
}

// Set sets the key-value pair in leveldb storage
func (kv *database) Set(p []byte, v []byte) error {
	return kv.db.Put(p, v, nil)
}

func (kv *database) Delete(p []byte) error {
	return kv.db.Delete(p, nil)
}

func (kv *database) Has(p []byte) (bool, error) {
	return kv.db.Has(p, nil)
}

// Get retrieves the key-value pair in leveldb storage
func (kv *database) Get(p []byte) ([]byte, bool, error) {
	data, err := kv.db.Get(p, nil)
	if err != nil {
		if errors.Is(err, leveldb.ErrNotFound) {
			return nil, false, nil
		} else if errors.Is(err, leveldb.ErrClosed) {
			return nil, false, err
		} else {
			panic(err)
		}
	}

	return data, true, nil
}

// Close closes the leveldb storage instance
func (kv *database) Close() error {
	return kv.db.Close()
}

func max(a, b int) int {
	if a > b {
		return a
	}

	return b
}

type builder struct {
	logger  hclog.Logger
	path    string
	options *opt.Options
}

func (b *builder) SetCacheSize(cacheSize int) Builder {
	cacheSize = max(cacheSize, minCache)

	b.options.BlockCacheCapacity = cacheSize * opt.MiB

	b.logger.Info("leveldb",
		"BlockCacheCapacity", fmt.Sprintf("%d Mib", cacheSize),
	)

	return b
}

func (b *builder) SetHandles(handles int) Builder {
	b.options.OpenFilesCacheCapacity = max(handles, minHandles)

	b.logger.Info("leveldb",
		"OpenFilesCacheCapacity", b.options.OpenFilesCacheCapacity,
	)

	return b
}

func (b *builder) SetBloomKeyBits(bloomKeyBits int) Builder {
	b.options.Filter = filter.NewBloomFilter(bloomKeyBits)

	b.logger.Info("leveldb",
		"BloomFilter bits", bloomKeyBits,
	)

	return b
}

func (b *builder) SetCompactionTableSize(compactionTableSize int) Builder {
	b.options.CompactionTableSize = compactionTableSize * opt.MiB
	b.options.WriteBuffer = b.options.CompactionTableSize * 2

	b.logger.Info("leveldb",
		"CompactionTableSize", fmt.Sprintf("%d Mib", compactionTableSize),
		"WriteBuffer", fmt.Sprintf("%d Mib", b.options.WriteBuffer/opt.MiB),
	)

	return b
}

func (b *builder) SetCompactionTotalSize(compactionTotalSize int) Builder {
	b.options.CompactionTotalSize = compactionTotalSize * opt.MiB

	b.logger.Info("leveldb",
		"CompactionTotalSize", fmt.Sprintf("%d Mib", compactionTotalSize),
	)

	return b
}

func (b *builder) SetNoSync(noSync bool) Builder {
	b.options.NoSync = noSync

	b.logger.Info("leveldb",
		"NoSync", noSync,
	)

	return b
}

func (b *builder) Build() (kvdb.KVBatchStorage, error) {
	db, err := leveldb.OpenFile(b.path, b.options)
	if err != nil {
		return nil, err
	}

	return &database{db: db}, nil
}
