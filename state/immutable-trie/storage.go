package itrie

import (
	"fmt"

	"github.com/dogechain-lab/dogechain/helper/hex"
	"github.com/dogechain-lab/dogechain/helper/kvdb"
	"github.com/dogechain-lab/fastrlp"
)

var parserPool fastrlp.ParserPool

type StorageReader kvdb.KVReader
type StorageWriter kvdb.KVWriter
type Batch kvdb.Batch

// type Storage kvdb.KVBatchStorage

// Storage stores the trie
type Storage interface {
	StorageReader
	StorageWriter

	NewBatch() Batch
	Close() error
}

type kvStorageBatch struct {
	batch Batch
}

func (kvBatch *kvStorageBatch) Set(k, v []byte) error {
	kvBatch.batch.Set(k, v)

	return nil
}

func (kvBatch *kvStorageBatch) Delete(k []byte) error {
	return kvBatch.batch.Delete(k)
}

// ValueSize retrieves the amount of data queued up for writing.
func (kvBatch *kvStorageBatch) ValueSize() int {
	return kvBatch.batch.ValueSize()
}

// Write flushes any accumulated data to disk.
func (kvBatch *kvStorageBatch) Write() error {
	return kvBatch.batch.Write()
}

// Reset resets the batch for reuse.
func (kvBatch *kvStorageBatch) Reset() {
	kvBatch.batch.Reset()
}

// Replay replays the batch contents.
func (kvBatch *kvStorageBatch) Replay(w kvdb.KVWriter) error {
	return kvBatch.batch.Replay(w)
}

// wrap generic kvdb storage to implement Storage interface
type kvStorage struct {
	db kvdb.KVBatchStorage
}

func (kv *kvStorage) Has(k []byte) (bool, error) {
	return kv.db.Has(k)
}

func (kv *kvStorage) Get(k []byte) ([]byte, bool, error) {
	return kv.db.Get(k)
}

func (kv *kvStorage) Set(k, v []byte) error {
	return kv.db.Set(k, v)
}

func (kv *kvStorage) Delete(k []byte) error {
	return kv.db.Delete(k)
}

func (kv *kvStorage) NewBatch() Batch {
	return &kvStorageBatch{
		batch: kv.db.NewBatch(),
	}
}

func (kv *kvStorage) Close() error {
	return kv.db.Close()
}

func NewLevelDBStorage(db kvdb.KVBatchStorage) Storage {
	return &kvStorage{db: db}
}

// memkeyvalue is a key-value tuple tagged with a deletion field to allow creating
// memory-database write batches.
type memkeyvalue struct {
	key    []byte
	value  []byte
	delete bool
}

type memStorage struct {
	db map[string][]byte
}

// NewMemoryStorage creates an inmemory trie storage
func NewMemoryStorage() Storage {
	return &memStorage{db: map[string][]byte{}}
}

func (m *memStorage) Set(p []byte, v []byte) error {
	buf := make([]byte, len(v))
	copy(buf[:], v[:])
	m.db[hex.EncodeToHex(p)] = buf

	return nil
}

func (m *memStorage) Delete(p []byte) error {
	delete(m.db, hex.EncodeToHex(p))

	return nil
}

func (m *memStorage) Has(p []byte) (bool, error) {
	_, ok := m.db[hex.EncodeToHex(p)]

	return ok, nil
}

func (m *memStorage) Get(p []byte) ([]byte, bool, error) {
	v, ok := m.db[hex.EncodeToHex(p)]
	if !ok {
		return []byte{}, false, nil
	}

	return v, true, nil
}

func (m *memStorage) NewBatch() Batch {
	return &memBatch{db: m}
}

func (m *memStorage) Close() error {
	return nil
}

// memBatch is a write-only memory batch that commits changes to its host
// database when Write is called. A batch cannot be used concurrently.
type memBatch struct {
	db     *memStorage
	writes []memkeyvalue
	size   int
}

func (m *memBatch) Set(p, v []byte) error {
	buf := make([]byte, len(v))
	copy(buf[:], v[:])
	(*&m.db.db)[hex.EncodeToHex(p)] = buf

	return nil
}

func (m *memBatch) Delete(p []byte) error {
	delete(m.db.db, hex.EncodeToHex(p))

	return nil
}

// // ValueSize retrieves the amount of data queued up for writing.
// func (m *memBatch) ValueSize() int {
// 	return 0
// }

// Write flushes any accumulated data to disk.
func (m *memBatch) Write() error {
	return nil
}

// // Reset resets the batch for reuse.
// func (m *memBatch) Reset() {

// }

// // Replay replays the batch contents.
// func (m *memBatch) Replay(w kvdb.KVWriter) error {
// 	return nil
// }

// GetNode retrieves a node from storage
func GetNode(root []byte, storage StorageReader) (Node, bool, error) {
	data, ok, _ := storage.Get(root)
	if !ok {
		return nil, false, nil
	}

	// NOTE. We dont need to make copies of the bytes because the nodes
	// take the reference from data itself which is a safe copy.
	p := parserPool.Get()
	defer parserPool.Put(p)

	v, err := p.Parse(data)
	if err != nil {
		return nil, false, err
	}

	if v.Type() != fastrlp.TypeArray {
		return nil, false, fmt.Errorf("storage item should be an array")
	}

	n, err := decodeNode(v)

	return n, err == nil, err
}

func decodeNode(v *fastrlp.Value) (Node, error) {
	if v.Type() == fastrlp.TypeBytes {
		vv := &ValueNode{
			hash: true,
		}
		vv.buf = append(vv.buf[:0], v.Raw()...)

		return vv, nil
	}

	var err error

	// TODO remove this once 1.0.4 of ifshort is merged in golangci-lint
	ll := v.Elems() //nolint:ifshort
	if ll == 2 {
		key := v.Get(0)
		if key.Type() != fastrlp.TypeBytes {
			return nil, fmt.Errorf("short key expected to be bytes")
		}

		// this can be either an array (extension node)
		// or bytes (leaf node)
		nc := &ShortNode{}
		nc.key = decodeCompact(key.Raw())

		if hasTerminator(nc.key) {
			// value node
			if v.Get(1).Type() != fastrlp.TypeBytes {
				return nil, fmt.Errorf("short leaf value expected to be bytes")
			}

			vv := &ValueNode{}
			vv.buf = append(vv.buf, v.Get(1).Raw()...)
			nc.child = vv
		} else {
			nc.child, err = decodeNode(v.Get(1))
			if err != nil {
				return nil, err
			}
		}

		return nc, nil
	} else if ll == 17 {
		// full node
		nc := &FullNode{}
		for i := 0; i < 16; i++ {
			if v.Get(i).Type() == fastrlp.TypeBytes && len(v.Get(i).Raw()) == 0 {
				// empty
				continue
			}
			nc.children[i], err = decodeNode(v.Get(i))
			if err != nil {
				return nil, err
			}
		}

		if v.Get(16).Type() != fastrlp.TypeBytes {
			return nil, fmt.Errorf("full node value expected to be bytes")
		}
		if len(v.Get(16).Raw()) != 0 {
			vv := &ValueNode{}
			vv.buf = append(vv.buf[:0], v.Get(16).Raw()...)
			nc.value = vv
		}

		return nc, nil
	}

	return nil, fmt.Errorf("node has incorrect number of leafs")
}
