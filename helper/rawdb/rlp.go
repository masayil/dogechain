package rawdb

import (
	"errors"

	"github.com/dogechain-lab/dogechain/helper/kvdb"
	"github.com/dogechain-lab/dogechain/types"
	"github.com/dogechain-lab/fastrlp"
)

var ErrNotFound = errors.New("not found")

func readRLP(db kvdb.KVReader, key []byte, raw types.RLPUnmarshaler) error {
	data, ok, err := db.Get(key)
	if err != nil {
		return err
	} else if !ok {
		return ErrNotFound
	}

	if obj, ok := raw.(types.RLPStoreUnmarshaler); ok {
		// decode in the store format
		return obj.UnmarshalStoreRLP(data)
	}

	// normal rlp decoding
	return raw.UnmarshalRLP(data)
}

func writeRLP(db kvdb.KVWriter, key []byte, raw types.RLPMarshaler) error {
	var data []byte
	if obj, ok := raw.(types.RLPStoreMarshaler); ok {
		data = obj.MarshalStoreRLPTo(nil)
	} else {
		data = raw.MarshalRLPTo(nil)
	}

	return db.Set(key, data)
}

func readRLP2(db kvdb.Reader, key []byte) (*fastrlp.Value, error) {
	data, ok, err := db.Get(key)
	if err != nil {
		return nil, err
	} else if !ok {
		return nil, ErrNotFound
	}

	return types.RlpUnmarshal(data)
}

func writeRLP2(db kvdb.KVWriter, key []byte, v *fastrlp.Value) error {
	dst := v.MarshalTo(nil)

	return db.Set(key, dst)
}
