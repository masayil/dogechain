package leveldb

import (
	"fmt"

	"github.com/hashicorp/go-hclog"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

type optionType string

const (
	optionArg optionType = "FuncArgument" // Function argument
)

const (
	bloomKeyBits        = "bloomKeyBits"
	cacheSize           = "cacheSize"
	compactionTableSize = "compactionTableSize"
	compactionTotalSize = "compactionTotalSize"
	handles             = "handles"
	logger              = "logger"
	noSync              = "noSync"
	readOnly            = "readOnly"
)

type (
	optionValue struct {
		Value interface{}
		Type  optionType
	}

	// Option leveldb option
	Option func(map[string]optionValue) error
)

func addArg(key string, value interface{}) Option {
	return func(params map[string]optionValue) error {
		if value == nil {
			return nil
		}

		params[key] = optionValue{value, optionArg}

		return nil
	}
}

func addArgError(err error) Option {
	return func(map[string]optionValue) error {
		return err
	}
}

// SetBloomKeyBits sets bloom filter bits per key
func SetBloomKeyBits(v int) Option {
	if v <= 0 {
		return addArgError(fmt.Errorf("%s value must greater than 0", bloomKeyBits))
	}

	return addArg(bloomKeyBits, v)
}

// SetCacheSize sets the cache size in MiB
func SetCacheSize(v int) Option {
	if v <= 0 {
		return addArgError(fmt.Errorf("%s value must greater than 0 MiB", cacheSize))
	}

	return addArg(cacheSize, v)
}

// SetCompactionTableSize sets compaction table size in MiB and
// a write buffer twice the size
//
// It limits size of 'sorted table' that compaction generates.
func SetCompactionTableSize(v int) Option {
	if v <= 0 {
		return addArgError(fmt.Errorf("%s value must greater than 0 MiB", compactionTableSize))
	}

	return addArg(compactionTableSize, v)
}

// CompactionTotalSize sets total size of compaction table size
// in MiB
//
// It limits total size of 'sorted table' for each level.
func SetCompactionTotalSize(v int) Option {
	if v <= 0 {
		return addArgError(fmt.Errorf("%s value must greater than 0 MiB", compactionTotalSize))
	}

	return addArg(compactionTotalSize, v)
}

// SetHandles sets the handles (file discriptor count)
func SetHandles(v int) Option {
	if v <= 0 {
		return addArgError(fmt.Errorf("%s value must greater than 0", handles))
	}

	return addArg(handles, v)
}

// SetLogger sets the outside logger to it
//
// The default one print out nothing
func SetLogger(v Logger) Option {
	if v == nil {
		v = hclog.NewNullLogger()
	}

	return addArg(logger, v)
}

// NoSync allows completely disable fsync
func SetNoSync(v bool) Option {
	return addArg(noSync, v)
}

func SetReadonly(v bool) Option {
	return addArg(readOnly, v)
}

func defaultLevelDBOptions() *opt.Options {
	return &opt.Options{
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
	}
}

type dbOption struct {
	logger  Logger
	options *opt.Options
}

func handleOptions(o *dbOption, options []Option) error {
	params := map[string]optionValue{}

	for _, option := range options {
		if option != nil {
			if err := option(params); err != nil {
				return err
			}
		}
	}

PARAM_LOOP:
	for k, v := range params {
		//nolint:forcetypeassert
		switch k {
		case bloomKeyBits:
			o.options.Filter = filter.NewBloomFilter(v.Value.(int))
		case cacheSize:
			o.options.BlockCacheCapacity = v.Value.(int) * opt.MiB
		case compactionTableSize:
			o.options.CompactionTableSize = v.Value.(int) * opt.MiB
			o.options.WriteBuffer = o.options.CompactionTableSize * 2
		case compactionTotalSize:
			o.options.CompactionTotalSize = v.Value.(int) * opt.MiB
		case handles:
			o.options.OpenFilesCacheCapacity = v.Value.(int)
		case logger:
			o.logger = v.Value.(Logger)
		case noSync:
			o.options.NoSync = v.Value.(bool)
		case readOnly:
			o.options.ReadOnly = v.Value.(bool)
		default:
			continue PARAM_LOOP
		}

		o.logger.Info("set leveldb option", "key", k, "value", v.Value)
	}

	return nil
}
