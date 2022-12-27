package kvstorage

import (
	"github.com/dogechain-lab/dogechain/blockchain/storage"
	"github.com/dogechain-lab/dogechain/helper/hex"
)

// memoryKV is an in memory implementation of the kv storage
type memoryKV struct {
	db map[string][]byte
}

func NewMemoryStorage() storage.Storage {
	return NewKeyValueStorage(
		&memoryKV{
			make(map[string][]byte),
		},
	)
}

func (m *memoryKV) Has(p []byte) (bool, error) {
	_, ok := m.db[hex.EncodeToHex(p)]

	return ok, nil
}

func (m *memoryKV) Get(p []byte) ([]byte, bool, error) {
	v, ok := m.db[hex.EncodeToHex(p)]
	if !ok {
		return nil, false, nil
	}

	return v, true, nil
}

func (m *memoryKV) Set(p []byte, v []byte) error {
	m.db[hex.EncodeToHex(p)] = v

	return nil
}

func (m *memoryKV) Delete(p []byte) error {
	delete(m.db, hex.EncodeToHex(p))

	return nil
}

func (m *memoryKV) Close() error {
	return nil
}
