package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/dogechain-lab/dogechain/chain"
	"github.com/dogechain-lab/dogechain/e2e/framework"
	"github.com/dogechain-lab/dogechain/helper/kvdb/leveldb"
	"github.com/dogechain-lab/dogechain/reverify"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
)

func TestReverify(t *testing.T) {
	const (
		toBlock      uint64 = 10
		finalToBlock uint64 = 15
	)

	svrs := framework.NewTestServers(t, 4, func(config *framework.TestServerConfig) {
		config.SetConsensus(framework.ConsensusDev)
		config.SetSeal(true)
		config.SetDevInterval(1)
	})

	errs := framework.WaitForServersToSeal(svrs, toBlock)
	for _, err := range errs {
		assert.NoError(t, err)
	}

	svr := svrs[0]
	// data dir
	svrRootDir := svr.Config.RootDir
	// block number
	currentBlockHeight, err := svr.JSONRPC().Eth().BlockNumber()
	assert.NoError(t, err)
	// stop server to make some db corruption
	svr.Stop()

	// wait for new blocks
	time.Sleep(time.Second * 2)

	// open trie database
	trie, err := leveldb.New(svr.StateDataDir())
	assert.NoError(t, err)

	// corrupt data
	{
		iter := trie.NewIterator(nil, nil)
		assert.NoError(t, iter.Error())

		var count uint64 = 0
		for iter.Next() {
			count++
			// do nothing to reach the end
			if count < currentBlockHeight {
				continue
			}

			err := trie.Set(iter.Key(), []byte("corrupted data"))
			assert.NoError(t, err)
		}

		assert.NoError(t, iter.Error())

		iter.Release()

		err := trie.Close()
		assert.NoError(t, err)
	}

	genesis, parseErr := chain.Import(svr.GenesisFile())
	assert.NoError(t, parseErr)

	err = reverify.ReverifyChain(
		hclog.NewNullLogger(),
		genesis,
		svrRootDir,
		1,
	)

	assert.NoError(t, err)

	resvr := framework.NewTestServer(t, svrRootDir, func(config *framework.TestServerConfig) {
		*config = *svr.Config
		config.Seal = false
	})

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	t.Cleanup(cancel)

	err = resvr.Start(ctx)
	assert.NoError(t, err)
	t.Cleanup(func() {
		resvr.Stop()
	})

	_, err = framework.WaitUntilBlockMined(ctx, resvr, finalToBlock)
	assert.NoError(t, err)
}
