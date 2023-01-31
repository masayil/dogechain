package rawdb

import (
	"math/big"

	"github.com/dogechain-lab/dogechain/helper/kvdb"
	"github.com/dogechain-lab/dogechain/types"
	"github.com/umbracle/fastrlp"
)

func ReadBody(db kvdb.KVReader, hash types.Hash) (*types.Body, error) {
	body := new(types.Body)
	err := readRLP(db, bodyKey(hash), body)

	return body, err
}

func WriteBody(db kvdb.KVWriter, hash types.Hash, body *types.Body) error {
	return writeRLP(db, bodyKey(hash), body)
}

func ReadCanonicalHash(db kvdb.KVReader, number uint64) (types.Hash, bool) {
	data, ok, err := db.Get(canonicalHashKey(number))
	if err != nil || !ok {
		return types.Hash{}, false
	}

	return types.BytesToHash(data), true
}

func WriteCanonicalHash(db kvdb.KVWriter, number uint64, hash types.Hash) error {
	return db.Set(canonicalHashKey(number), hash.Bytes())
}

func ReadTotalDifficulty(db kvdb.KVReader, hash types.Hash) (*big.Int, bool) {
	data, ok, err := db.Get(difficultyKey(hash))
	if err != nil || !ok {
		return nil, false
	}

	return new(big.Int).SetBytes(data), true
}

func WriteTotalDifficulty(db kvdb.KVWriter, hash types.Hash, diff *big.Int) error {
	return db.Set(difficultyKey(hash), diff.Bytes())
}

func ReadHeader(db kvdb.KVReader, hash types.Hash) (*types.Header, error) {
	header := new(types.Header)
	err := readRLP(db, headerKey(hash), header)

	return header, err
}

func WriteHeader(db kvdb.KVWriter, hash types.Hash, header *types.Header) error {
	return writeRLP(db, headerKey(hash), header)
}

func ReadReceipts(db kvdb.KVReader, hash types.Hash) ([]*types.Receipt, error) {
	receipts := &types.Receipts{}
	err := readRLP(db, receiptsKey(hash), receipts)

	return *receipts, err
}

func WriteReceipts(db kvdb.KVWriter, hash types.Hash, receipts []*types.Receipt) error {
	v := types.Receipts(receipts)

	return writeRLP(db, receiptsKey(hash), &v)
}

func ReadTxLookup(db kvdb.KVReader, hash types.Hash) (types.Hash, bool) {
	v, err := readRLP2(db, txLookupKey(hash))
	if err != nil {
		return types.Hash{}, false
	}

	blockHash, err := v.GetBytes(nil, 32)
	if err != nil {
		return types.Hash{}, false
	}

	return types.BytesToHash(blockHash), true
}

func WriteTxLookup(db kvdb.KVWriter, hash types.Hash, blockHash types.Hash) error {
	var ar fastrlp.Arena
	v := ar.NewBytes(blockHash.Bytes())

	return writeRLP2(db, txLookupKey(hash), v)
}

func ReadHeadHash(db kvdb.KVReader) (types.Hash, bool) {
	data, ok, err := db.Get(headHashKey)
	if err != nil || !ok {
		return types.Hash{}, false
	}

	return types.BytesToHash(data), true
}

func WriteHeadHash(db kvdb.KVWriter, hash types.Hash) error {
	return db.Set(headHashKey, hash.Bytes())
}

func ReadHeadNumber(db kvdb.KVReader) (uint64, bool) {
	data, ok, err := db.Get(headNumberKey)
	if err != nil || !ok {
		return 0, false
	}

	if len(data) != 8 {
		return 0, false
	}

	return decodeUint(data), true
}

func WriteHeadNumber(db kvdb.KVWriter, number uint64) error {
	return db.Set(headNumberKey, encodeUint(number))
}

func ReadForks(db kvdb.KVReader) ([]types.Hash, error) {
	forks := &Forks{}
	err := readRLP(db, forkEmptyKey, forks)

	return *forks, err
}

func WriteForks(db kvdb.KVWriter, forks []types.Hash) error {
	ff := Forks(forks)

	return writeRLP(db, forkEmptyKey, &ff)
}
