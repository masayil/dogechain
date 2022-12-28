package reverify

import (
	"fmt"
	"path/filepath"

	"github.com/hashicorp/go-hclog"

	"github.com/dogechain-lab/dogechain/chain"
	"github.com/dogechain-lab/dogechain/helper/kvdb/leveldb"
	itrie "github.com/dogechain-lab/dogechain/state/immutable-trie"
)

const (
	_blockchainDir = "blockchain"
	_stateDir      = "trie"
	_consensusDir  = "consensus"
)

func stateDir(dataDir string) string {
	return filepath.Join(dataDir, _stateDir)
}

func blockchainDir(dataDir string) string {
	return filepath.Join(dataDir, _blockchainDir)
}

func consensusDir(dataDir string) string {
	return filepath.Join(dataDir, _consensusDir)
}

func ReverifyChain(
	logger hclog.Logger,
	chain *chain.Chain,
	dataDir string,
	startHeight uint64,
) error {
	chainDB, err := leveldb.New(
		blockchainDir(dataDir),
		leveldb.SetLogger(logger),
	)
	if err != nil {
		return err
	}

	defer chainDB.Close()

	hub := &DBHub{chainDB: chainDB}

	trieDB, err := leveldb.New(
		stateDir(dataDir),
		leveldb.SetLogger(logger),
	)
	if err != nil {
		return err
	}

	defer trieDB.Close()

	blockchain, consensus, err := createBlockchain(
		hub,
		chainDB,
		logger,
		chain,
		itrie.NewStateDB(trieDB, hclog.NewNullLogger(), itrie.NilMetrics()),
		dataDir,
	)
	if err != nil {
		logger.Error("failed to create blockchain")

		return err
	}

	defer func() {
		consensus.Close()
		blockchain.Close()
	}()

	hash, ok := blockchain.GetHeaderHash()
	if ok {
		logger.Info("current blockchain hash", "hash", hash)
	}

	currentHeight, ok := blockchain.GetHeaderNumber()
	if ok {
		logger.Info("current blockchain height", "Height", currentHeight)
	}

	for i := startHeight; i <= currentHeight; i++ {
		haeder, ok := blockchain.GetHeaderByNumber(i)
		if !ok {
			return fmt.Errorf("failed to read canonical hash, height: %d, header: %v", i, haeder)
		}

		block, ok := blockchain.GetBlock(haeder.Hash, i, true)
		if !ok {
			return fmt.Errorf("failed to read block, height: %d, header: %v", i, haeder)
		}

		if err := blockchain.VerifyFinalizedBlock(block); err != nil {
			return fmt.Errorf("failed to verify block, height: %d, header: %v", i, haeder)
		}

		logger.Info("verify block success", "height", i, "hash", haeder.Hash, "txs", len(block.Transactions))
	}

	logger.Info(fmt.Sprintf("verify height from %d to %d \n", startHeight, currentHeight))

	return nil
}
