package kvdb

import (
	"encoding/binary"
	"io/ioutil"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
)

func createTestDB(t *testing.T) KVBatchStorage {
	t.Helper()

	tempDir, err := ioutil.TempDir("/tmp", "leveldb-")
	assert.NoError(t, err)

	db, err := NewLevelDBBuilder(
		hclog.NewNullLogger(),
		tempDir,
	).Build()
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		os.RemoveAll(tempDir)
	})

	return db
}

func Test_LevelDB_GetSet(t *testing.T) {
	t.Parallel()

	db := createTestDB(t)
	defer db.Close()

	var (
		key   = []byte("hello")
		value = []byte("world")
	)

	err := db.Set(key, value)
	assert.NoError(t, err)

	v, exist, err := db.Get(key)
	assert.NoError(t, err)
	assert.True(t, exist)
	assert.Equal(t, value, v)
}

func Test_LevelDB_BatchWrite(t *testing.T) {
	t.Parallel()

	db := createTestDB(t)
	defer db.Close()

	var (
		length = 100
		keys   = make([][]byte, 0, length)
		values = make([][]byte, 0, length)
		batch  = db.NewBatch()
	)

	for i := 0; i < length; i++ {
		// random seed
		seed := rand.NewSource(time.Now().Unix())
		r := rand.New(seed)
		// k, v
		key := make([]byte, 32)
		value := make([]byte, 128)
		// random read
		r.Read(key)
		r.Read(value)
		// save value
		keys = append(keys, key)
		values = append(values, value)
		// batch it
		batch.Set(key, value)
	}

	// batch write
	err := batch.Write()
	assert.NoError(t, err)

	for i := 0; i < length; i++ {
		val, exist, err := db.Get(keys[i])
		assert.NoError(t, err)
		assert.True(t, exist)
		assert.Equal(t, values[i], val)
	}
}

func Test_LevelDB_Iterator(t *testing.T) {
	t.Parallel()

	db := createTestDB(t)
	defer db.Close()

	var (
		length = 100
		keys   = make([][]byte, 0, length)
		values = make([][]byte, 0, length)
		batch  = db.NewBatch()
	)

	for i := 0; i < length; i++ {
		// random seed
		seed := rand.NewSource(time.Now().Unix())
		r := rand.New(seed)
		// key
		key := make([]byte, 32)
		r.Read(key)
		// key prefix
		prefix := make([]byte, 4)
		binary.LittleEndian.PutUint32(prefix, uint32(i))
		// append prefix
		key = append(prefix[:], key...)
		// value
		value := make([]byte, 128)
		r.Read(value)
		// k, v
		keys = append(keys, key)
		values = append(values, value)
		// batch it
		batch.Set(key, value)
	}

	if err := batch.Write(); err != nil {
		t.Fatal(err)
	}

	iter := db.NewIterator(nil, nil)
	defer iter.Release()

	count := 0

	for iter.Next() {
		k := iter.Key()
		v := iter.Value()

		assert.Equal(t, keys[count], k)
		assert.Equal(t, values[count], v)

		count++
	}

	assert.Equal(t, length, count)
	assert.NoError(t, iter.Error())
}
