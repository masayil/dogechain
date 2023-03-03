package server

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"

	"github.com/dogechain-lab/dogechain/blockchain"
	"github.com/dogechain-lab/dogechain/chain"
	"github.com/dogechain-lab/dogechain/consensus"
	"github.com/dogechain-lab/dogechain/helper/progress"
	"github.com/dogechain-lab/dogechain/jsonrpc"
	"github.com/dogechain-lab/dogechain/network"
	"github.com/dogechain-lab/dogechain/state"
	"github.com/dogechain-lab/dogechain/state/runtime"
	"github.com/dogechain-lab/dogechain/state/snapshot"
	"github.com/dogechain-lab/dogechain/state/stypes"
	"github.com/dogechain-lab/dogechain/txpool"
	"github.com/dogechain-lab/dogechain/types"
)

type jsonRPCStore struct {
	blockchain         *blockchain.Blockchain
	restoreProgression *progress.ProgressionWrapper
	txpool             *txpool.TxPool
	executor           *state.Executor

	consensus consensus.Consensus
	server    network.Server
	state     state.State
	snaps     *snapshot.Tree

	metrics *jsonrpcStoreMetrics
}

func NewJSONRPCStore(
	snaps *snapshot.Tree,
	state state.State,
	blockchain *blockchain.Blockchain,
	restoreProgression *progress.ProgressionWrapper,
	txpool *txpool.TxPool,
	executor *state.Executor,
	consensus consensus.Consensus,
	network network.Server,
	metrics *jsonrpcStoreMetrics,
) jsonrpc.JSONRPCStore {
	if metrics == nil {
		metrics = JSONRPCStoreNilMetrics()
	}

	return &jsonRPCStore{
		blockchain:         blockchain,
		restoreProgression: restoreProgression,
		txpool:             txpool,
		executor:           executor,
		consensus:          consensus,
		server:             network,
		state:              state,
		snaps:              snaps,
		metrics:            metrics,
	}
}

// jsonrpc.ethTxPoolStore interface

// GetNonce returns the next nonce for this address
func (j *jsonRPCStore) GetNonce(addr types.Address) uint64 {
	j.metrics.GetNonceInc()

	return j.txpool.GetNonce(addr)
}

// AddTx adds a new transaction to the tx pool
func (j *jsonRPCStore) AddTx(tx *types.Transaction) error {
	j.metrics.AddTxInc()

	return j.txpool.AddTx(tx)
}

// GetPendingTx gets the pending transaction from the transaction pool, if it's present
func (j *jsonRPCStore) GetPendingTx(txHash types.Hash) (*types.Transaction, bool) {
	j.metrics.GetPendingTxInc()

	return j.txpool.GetPendingTx(txHash)
}

// jsonrpc.ethStateStore interface
func (j *jsonRPCStore) GetAccount(stateRoot types.Hash, addr types.Address) (*stypes.Account, error) {
	j.metrics.GetAccountInc()

	return getCommittedAccount(j.snaps, j.state, stateRoot, addr)
}

func (j *jsonRPCStore) GetStorage(stateRoot types.Hash, addr types.Address, slot types.Hash) (types.Hash, error) {
	j.metrics.GetStorageInc()

	return getCommittedStorage(j.snaps, j.state, stateRoot, addr, slot)
}

// GetForksInTime returns the active forks at the given block height
func (j *jsonRPCStore) GetForksInTime(blockNumber uint64) chain.ForksInTime {
	j.metrics.GetForksInTimeInc()

	return j.executor.GetForksInTime(blockNumber)
}

// func (j *jsonRPCStore) GetCode(root types.Hash) ([]byte, error) {
func (j *jsonRPCStore) GetCode(root types.Hash, addr types.Address) ([]byte, error) {
	j.metrics.GetCodeInc()

	account, err := getCommittedAccount(j.snaps, j.state, root, addr)
	if err != nil {
		return nil, err
	}

	if len(account.CodeHash) == 0 || bytes.Equal(account.CodeHash, types.EmptyRootHash[:]) {
		return []byte{}, nil
	}

	code, ok := j.state.GetCode(types.BytesToHash(account.CodeHash))
	if !ok {
		return nil, fmt.Errorf("unable to fetch code")
	}

	return code, nil
}

// jsonrpc.ethBlockchainStore interface

// Header returns the current header of the chain (genesis if empty)
func (j *jsonRPCStore) Header() *types.Header {
	j.metrics.HeaderInc()

	return j.blockchain.Header()
}

// GetHeaderByNumber returns the header by number
func (j *jsonRPCStore) GetHeaderByNumber(n uint64) (*types.Header, bool) {
	j.metrics.GetHeaderByNumberInc()

	return j.blockchain.GetHeaderByNumber(n)
}

// GetHeaderByHash returns the header by hash
func (j *jsonRPCStore) GetHeaderByHash(hash types.Hash) (*types.Header, bool) {
	j.metrics.GetHeaderByHashInc()

	return j.blockchain.GetHeaderByHash(hash)
}

// GetBlockByHash gets a block using the provided hash
func (j *jsonRPCStore) GetBlockByHash(hash types.Hash, full bool) (*types.Block, bool) {
	j.metrics.GetBlockByHashInc()

	return j.blockchain.GetBlockByHash(hash, full)
}

// GetBlockByNumber returns a block using the provided number
func (j *jsonRPCStore) GetBlockByNumber(number uint64, full bool) (*types.Block, bool) {
	j.metrics.GetBlockByNumberInc()

	return j.blockchain.GetBlockByNumber(number, full)
}

// ReadTxLookup returns a block hash in which a given txn was mined
func (j *jsonRPCStore) ReadTxLookup(txnHash types.Hash) (types.Hash, bool) {
	j.metrics.ReadTxLookupInc()

	return j.blockchain.ReadTxLookup(txnHash)
}

// GetReceiptsByHash returns the receipts for a block hash
func (j *jsonRPCStore) GetReceiptsByHash(hash types.Hash) ([]*types.Receipt, error) {
	j.metrics.GetReceiptsByHashInc()

	return j.blockchain.GetReceiptsByHash(hash)
}

// GetAvgGasPrice returns the average gas price
func (j *jsonRPCStore) GetAvgGasPrice() *big.Int {
	j.metrics.GetAvgGasPriceInc()

	return j.blockchain.GetAvgGasPrice()
}

// ApplyTxn applies a transaction object to the blockchain
func (j *jsonRPCStore) ApplyTxn(
	header *types.Header,
	txn *types.Transaction,
) (result *runtime.ExecutionResult, err error) {
	j.metrics.ApplyTxnInc()

	blockCreator, err := j.consensus.GetBlockCreator(header)
	if err != nil {
		return nil, err
	}

	transition, err := j.executor.BeginTxn(header.StateRoot, header, blockCreator)

	if err != nil {
		return
	}

	result, err = transition.Apply(txn)

	return
}

// GetSyncProgression retrieves the current sync progression, if any
func (j *jsonRPCStore) GetSyncProgression() *progress.Progression {
	j.metrics.GetSyncProgressionInc()

	// restore progression
	if restoreProg := j.restoreProgression.GetProgression(); restoreProg != nil {
		return restoreProg
	}

	// consensus sync progression
	if consensusSyncProg := j.consensus.GetSyncProgression(); consensusSyncProg != nil {
		return consensusSyncProg
	}

	return nil
}

// StateAtTransaction returns the execution environment of a certain transaction.
// The transition should not commit, it shall be collected by GC.
func (j *jsonRPCStore) StateAtTransaction(block *types.Block, txIndex int) (*state.Transition, error) {
	j.metrics.StateAtTransactionInc()

	if block.Number() == 0 {
		return nil, errors.New("no transaction in genesis")
	}

	if txIndex < 0 {
		return nil, errors.New("invalid transaction index")
	}

	// get parent header
	parent, exists := j.blockchain.GetParent(block.Header)
	if !exists {
		return nil, fmt.Errorf("parent %s not found", block.ParentHash())
	}

	// block creator
	blockCreator, err := j.consensus.GetBlockCreator(block.Header)
	if err != nil {
		return nil, err
	}

	// begin transition, use parent block
	txn, err := j.executor.BeginTxn(parent.StateRoot, parent, blockCreator)
	if err != nil {
		return nil, err
	}

	if txIndex == 0 {
		return txn, nil
	}

	for idx, tx := range block.Transactions {
		if idx == txIndex {
			return txn, nil
		}

		if _, err := txn.Apply(tx); err != nil {
			return nil, fmt.Errorf("transaction %s failed: %w", tx.Hash(), err)
		}
	}

	return nil, fmt.Errorf("transaction index %d out of range for block %s", txIndex, block.Hash())
}

// jsonrpc.networkStore interface

func (j *jsonRPCStore) PeerCount() int64 {
	j.metrics.PeerCountInc()

	return j.server.PeerCount()
}

// jsonrpc.txPoolStore interface

// GetTxs gets tx pool transactions currently pending for inclusion and currently queued for validation
func (j *jsonRPCStore) GetTxs(inclQueued bool) (
	map[types.Address][]*types.Transaction, map[types.Address][]*types.Transaction,
) {
	j.metrics.GetTxsInc()

	return j.txpool.GetTxs(inclQueued)
}

// GetCapacity returns the current and max capacity of the pool in slots
func (j *jsonRPCStore) GetCapacity() (uint64, uint64) {
	j.metrics.GetCapacityInc()

	return j.txpool.GetCapacity()
}

// jsonrpc.filterManagerStore interface

func (j *jsonRPCStore) SubscribeEvents() blockchain.Subscription {
	j.metrics.SubscribeEventsInc()

	return j.blockchain.SubscribeEvents()
}

func (j *jsonRPCStore) GetDDosContractList() map[string]map[types.Address]int {
	return j.txpool.GetDDosContractList()
}
