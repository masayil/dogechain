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
		if !assert.NoError(t, err) {
			t.FailNow()
		}
	}

	svr := svrs[0]
	// data dir
	svrRootDir := svr.Config.RootDir
	// block number
	currentBlockHeight, err := svr.JSONRPC().Eth().BlockNumber()
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	// stop server to make some db corruption
	svr.Stop()

	// wait for the process to return leveldb file lock
	time.Sleep(3 * time.Second)

	// open trie database
	trie, err := leveldb.New(svr.StateDataDir())
	if !assert.NoError(t, err) {
		t.FailNow()
	}

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
			if !assert.NoError(t, err) {
				t.FailNow()
			}
		}

		if !assert.NoError(t, iter.Error()) {
			t.FailNow()
		}

		iter.Release()

		err := trie.Close()
		if !assert.NoError(t, err) {
			t.FailNow()
		}
	}

	genesis, parseErr := chain.Import(svr.GenesisFile())
	if !assert.NoError(t, parseErr) {
		t.FailNow()
	}

	err = reverify.ReverifyChain(
		hclog.NewNullLogger(),
		genesis,
		svrRootDir,
		1,
	)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	// wait for the process to return leveldb file lock
	time.Sleep(3 * time.Second)

	resvr := framework.NewTestServer(t, svrRootDir, func(config *framework.TestServerConfig) {
		*config = *svr.Config
	})

	func() {
		ctx, cancel := context.WithTimeout(context.Background(), serverStartTimeout)
		defer cancel()

		err := resvr.Start(ctx)
		if err != nil {
			t.Fatal(err)
		}
	}()

	t.Cleanup(func() {
		resvr.Stop()
	})

	func() {
		ctx, cancel := context.WithTimeout(context.Background(), transactionTimeout)
		defer cancel()

		_, err := framework.WaitUntilBlockMined(ctx, resvr, finalToBlock)
		assert.NoError(t, err)
	}()
}
