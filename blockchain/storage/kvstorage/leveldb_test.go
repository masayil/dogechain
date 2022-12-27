package kvstorage

import (
	"os"
	"testing"

	"github.com/dogechain-lab/dogechain/blockchain/storage"
	"github.com/dogechain-lab/dogechain/helper/kvdb/leveldb"
	"github.com/hashicorp/go-hclog"
)

func newLevelDBStorage(t *testing.T) storage.Storage {
	t.Helper()

	path, err := os.MkdirTemp("/tmp", "minimal_storage")
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		os.RemoveAll(path)
	})

	db, err := leveldb.New(path)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	s := NewKeyValueStorage(hclog.NewNullLogger(), db)

	return s
}

func TestLevelDBStorage(t *testing.T) {
	storage.TestStorage(t, newLevelDBStorage)
}
