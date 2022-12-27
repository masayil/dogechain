package kvstorage

import (
	"testing"

	"github.com/dogechain-lab/dogechain/blockchain/storage"
	"github.com/dogechain-lab/dogechain/helper/kvdb/memorydb"
)

func newStorage(t *testing.T) storage.Storage {
	t.Helper()

	db := memorydb.New()

	t.Cleanup(func() {
		db.Close()
	})

	return NewKeyValueStorage(db)
}

func TestLevelDBStorage(t *testing.T) {
	storage.TestStorage(t, newStorage)
}
