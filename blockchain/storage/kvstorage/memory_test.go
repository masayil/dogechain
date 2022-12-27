package kvstorage

import (
	"testing"

	"github.com/dogechain-lab/dogechain/blockchain/storage"
)

func TestMemoryStorage(t *testing.T) {
	t.Helper()

	f := func(t *testing.T) storage.Storage {
		t.Helper()

		s := NewMemoryStorage()

		return s
	}

	storage.TestStorage(t, f)
}
