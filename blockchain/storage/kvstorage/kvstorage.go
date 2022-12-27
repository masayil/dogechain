package kvstorage

import (
	"math/big"

	"github.com/dogechain-lab/dogechain/blockchain/storage"
	"github.com/dogechain-lab/dogechain/helper/kvdb"
	"github.com/dogechain-lab/dogechain/helper/rawdb"
	"github.com/dogechain-lab/dogechain/types"
)

// KeyValueStorage is a generic storage for kv databases
type KeyValueStorage struct {
	db kvdb.KVBatchStorage
}

func NewKeyValueStorage(db kvdb.KVBatchStorage) storage.Storage {
	return &KeyValueStorage{db: db}
}

// -- canonical hash --

// ReadCanonicalHash gets the hash from the number of the canonical chain
func (s *KeyValueStorage) ReadCanonicalHash(n uint64) (types.Hash, bool) {
	return rawdb.ReadCanonicalHash(s.db, n)
}

// WriteCanonicalHash writes a hash for a number block in the canonical chain
func (s *KeyValueStorage) WriteCanonicalHash(n uint64, hash types.Hash) error {
	return rawdb.WriteCanonicalHash(s.db, n, hash)
}

// HEAD //

// ReadHeadHash returns the hash of the head
func (s *KeyValueStorage) ReadHeadHash() (types.Hash, bool) {
	return rawdb.ReadHeadHash(s.db)
}

// ReadHeadNumber returns the number of the head
func (s *KeyValueStorage) ReadHeadNumber() (uint64, bool) {
	return rawdb.ReadHeadNumber(s.db)
}

// WriteHeadHash writes the hash of the head
func (s *KeyValueStorage) WriteHeadHash(h types.Hash) error {
	return rawdb.WriteHeadHash(s.db, h)
}

// WriteHeadNumber writes the number of the head
func (s *KeyValueStorage) WriteHeadNumber(n uint64) error {
	return rawdb.WriteHeadNumber(s.db, n)
}

// FORK //

// WriteForks writes the current forks
func (s *KeyValueStorage) WriteForks(forks []types.Hash) error {
	return rawdb.WriteForks(s.db, forks)
}

// ReadForks read the current forks
func (s *KeyValueStorage) ReadForks() ([]types.Hash, error) {
	return rawdb.ReadForks(s.db)
}

// DIFFICULTY //

// WriteTotalDifficulty writes the difficulty
func (s *KeyValueStorage) WriteTotalDifficulty(hash types.Hash, diff *big.Int) error {
	return rawdb.WriteTotalDifficulty(s.db, hash, diff)
}

// ReadTotalDifficulty reads the difficulty
func (s *KeyValueStorage) ReadTotalDifficulty(hash types.Hash) (*big.Int, bool) {
	return rawdb.ReadTotalDifficulty(s.db, hash)
}

// HEADER //

// WriteHeader writes the header
func (s *KeyValueStorage) WriteHeader(h *types.Header) error {
	return rawdb.WriteHeader(s.db, h.Hash, h)
}

// ReadHeader reads the header
func (s *KeyValueStorage) ReadHeader(hash types.Hash) (*types.Header, error) {
	return rawdb.ReadHeader(s.db, hash)
}

// WriteCanonicalHeader implements the storage interface
func (s *KeyValueStorage) WriteCanonicalHeader(h *types.Header, diff *big.Int) error {
	if err := s.WriteHeader(h); err != nil {
		return err
	}

	if err := s.WriteHeadHash(h.Hash); err != nil {
		return err
	}

	if err := s.WriteHeadNumber(h.Number); err != nil {
		return err
	}

	if err := s.WriteCanonicalHash(h.Number, h.Hash); err != nil {
		return err
	}

	if err := s.WriteTotalDifficulty(h.Hash, diff); err != nil {
		return err
	}

	return nil
}

// BODY //

// WriteBody writes the body
func (s *KeyValueStorage) WriteBody(hash types.Hash, body *types.Body) error {
	return rawdb.WriteBody(s.db, hash, body)
}

// ReadBody reads the body
func (s *KeyValueStorage) ReadBody(hash types.Hash) (*types.Body, error) {
	return rawdb.ReadBody(s.db, hash)
}

// RECEIPTS //

// WriteReceipts writes the receipts
func (s *KeyValueStorage) WriteReceipts(hash types.Hash, receipts []*types.Receipt) error {
	return rawdb.WriteReceipts(s.db, hash, receipts)
}

// ReadReceipts reads the receipts
func (s *KeyValueStorage) ReadReceipts(hash types.Hash) ([]*types.Receipt, error) {
	return rawdb.ReadReceipts(s.db, hash)
}

// TX LOOKUP //

// WriteTxLookup maps the transaction hash to the block hash
func (s *KeyValueStorage) WriteTxLookup(hash types.Hash, blockHash types.Hash) error {
	return rawdb.WriteTxLookup(s.db, hash, blockHash)
}

// ReadTxLookup reads the block hash using the transaction hash
func (s *KeyValueStorage) ReadTxLookup(hash types.Hash) (types.Hash, bool) {
	return rawdb.ReadTxLookup(s.db, hash)
}
