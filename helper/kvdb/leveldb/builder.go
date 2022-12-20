package leveldb

import (
	"github.com/dogechain-lab/dogechain/helper/kvdb"
	"github.com/hashicorp/go-hclog"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

type Builder interface {
	// set cache size
	SetCacheSize(int) Builder

	// set handles
	SetHandles(int) Builder

	// set bloom key bits
	SetBloomKeyBits(int) Builder

	// set compaction table size
	SetCompactionTableSize(int) Builder

	// set compaction table total size
	SetCompactionTotalSize(int) Builder

	// set no sync
	SetNoSync(bool) Builder

	// build the storage
	Build() (kvdb.KVBatchStorage, error)
}

// NewBuilder creates the new leveldb storage builder
func NewBuilder(logger hclog.Logger, path string) Builder {
	return &builder{
		logger: logger,
		path:   path,
		options: &opt.Options{
			OpenFilesCacheCapacity:        minHandles,
			CompactionTableSize:           DefaultCompactionTableSize * opt.MiB,
			CompactionTotalSize:           DefaultCompactionTotalSize * opt.MiB,
			BlockCacheCapacity:            minCache * opt.MiB,
			WriteBuffer:                   (DefaultCompactionTableSize * 2) * opt.MiB,
			CompactionTableSizeMultiplier: 1.1, // scale size up 1.1 multiple in next level
			Filter:                        filter.NewBloomFilter(DefaultBloomKeyBits),
			NoSync:                        false,
			BlockSize:                     256 * opt.KiB, // default 4kb, but one key-value pair need 0.5kb
			FilterBaseLg:                  19,            // 512kb
			DisableSeeksCompaction:        true,
		},
	}
}
