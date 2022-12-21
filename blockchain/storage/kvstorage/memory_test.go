package kvstorage

import (
	"testing"

	"github.com/dogechain-lab/dogechain/blockchain/storage"
	"github.com/hashicorp/go-hclog"
)

func TestMemoryStorage(t *testing.T) {
	t.Helper()

	f := func(t *testing.T) storage.Storage {
		t.Helper()

		s := NewMemoryStorage(hclog.NewNullLogger())

		return s
	}

	storage.TestStorage(t, f)
}
