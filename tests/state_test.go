package tests

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"math/big"
	"strings"
	"testing"

	"github.com/dogechain-lab/dogechain/chain"
	"github.com/dogechain-lab/dogechain/crypto"
	"github.com/dogechain-lab/dogechain/helper/hex"
	"github.com/dogechain-lab/dogechain/state"
	"github.com/dogechain-lab/dogechain/state/runtime/evm"
	"github.com/dogechain-lab/dogechain/state/runtime/precompiled"
	"github.com/dogechain-lab/dogechain/types"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
)

var (
	stateTests       = "GeneralStateTests"
	legacyStateTests = "LegacyTests/Constantinople/GeneralStateTests"
)

type stateCase struct {
	Info        *info                                   `json:"_info"`
	Env         *env                                    `json:"env"`
	Pre         map[types.Address]*chain.GenesisAccount `json:"pre"`
	Post        map[string]*postState                   `json:"post"`
	Transaction *stTransaction                          `json:"transaction"`
}

var ripemd = types.StringToAddress("0000000000000000000000000000000000000003")

func RunSpecificTest(t *testing.T, file string, c stateCase, name, fork string, index int, p postEntry) {
	t.Helper()

	config, ok := Forks[fork]
	if !ok {
		t.Fatalf("config %s not found", fork)
	}

	env := c.Env.ToEnv(t)

	msg, err := c.Transaction.At(p.Indexes)
	if err != nil {
		t.Fatal(err)
	}

	_, s, snapshot, pastRoot, err := buildState(c.Pre)
	assert.NoError(t, err)

	forks := config.At(uint64(env.Number))

	xxx := state.NewExecutor(&chain.Params{Forks: config, ChainID: 1}, hclog.NewNullLogger(), s)
	xxx.SetRuntime(precompiled.NewPrecompiled())
	xxx.SetRuntime(evm.NewEVM())

	xxx.PostHook = func(t *state.Transition) {
		if name == "failed_tx_xcf416c53" {
			// create the account
			t.Txn().TouchAccount(ripemd)
			// now remove it
			t.Txn().Suicide(ripemd)
		}
	}
	xxx.GetHash = func(*types.Header) func(i uint64) types.Hash {
		return vmTestBlockHash
	}

	executor, _ := xxx.BeginTxn(pastRoot, c.Env.ToHeader(t), env.Coinbase)
	executor.Apply(msg) //nolint:errcheck

	txn := executor.Txn()

	// mining rewards
	txn.AddSealingReward(env.Coinbase, big.NewInt(0))

	objs := txn.Commit(forks.EIP155)
	_, root, err := snapshot.Commit(objs)
	assert.NoError(t, err)

	if !bytes.Equal(root, p.Root.Bytes()) {
		t.Fatalf(
			"root mismatch (%s %s %s %d): expected %s but found %s",
			file,
			name,
			fork,
			index,
			p.Root.String(),
			hex.EncodeToHex(root),
		)
	}

	if logs := rlpHashLogs(txn.Logs()); logs != p.Logs {
		t.Fatalf(
			"logs mismatch (%s, %s %d): expected %s but found %s",
			name,
			fork,
			index,
			p.Logs.String(),
			logs.String(),
		)
	}
}

func RunSpecificTestWithSnapshot(t *testing.T, file string, c *stateCase, name, fork string, index int, p *postEntry) {
	t.Helper()

	config, ok := Forks[fork]
	if !ok {
		t.Fatalf("config %s not found", fork)
	}

	env := c.Env.ToEnv(t)

	msg, err := c.Transaction.At(p.Indexes)
	if err != nil {
		t.Fatal(err)
	}

	triedb, s, snapshot, pastRoot, err := buildState(c.Pre)
	assert.NoError(t, err)

	// _, err = buildSnapshotTree(triedb, pastRoot)
	snaps, err := buildSnapshotTree(triedb, pastRoot)
	if err != nil {
		t.Fatal(err)
	}

	forks := config.At(uint64(env.Number))

	xxx := state.NewExecutor(&chain.Params{Forks: config, ChainID: 1}, hclog.NewNullLogger(), s)
	xxx.SetRuntime(precompiled.NewPrecompiled())
	xxx.SetRuntime(evm.NewEVM())

	// set executor snapshot to test state features
	xxx.SetSnaps(snaps)

	xxx.PostHook = func(t *state.Transition) {
		if name == "failed_tx_xcf416c53" {
			// create the account
			t.Txn().TouchAccount(ripemd)
			// now remove it
			t.Txn().Suicide(ripemd)
		}
	}
	xxx.GetHash = func(*types.Header) func(i uint64) types.Hash {
		return vmTestBlockHash
	}

	executor, _ := xxx.BeginTxn(pastRoot, c.Env.ToHeader(t), env.Coinbase)
	executor.Apply(msg) //nolint:errcheck

	txn := executor.Txn()

	// mining rewards
	txn.AddSealingReward(env.Coinbase, big.NewInt(0))

	objs := txn.Commit(forks.EIP155)
	_, root, err := snapshot.Commit(objs)
	assert.NoError(t, err)

	if !bytes.Equal(root, p.Root.Bytes()) {
		t.Fatalf(
			"root mismatch (%s %s %s %d): expected %s but found %s",
			file,
			name,
			fork,
			index,
			p.Root,
			hex.EncodeToHex(root),
		)
	}

	// check post logs
	if logs := rlpHashLogs(txn.Logs()); logs != p.Logs {
		t.Fatalf(
			"logs mismatch (%s, %s %d): expected %s but found %s",
			name,
			fork,
			index,
			p.Logs.String(),
			logs.String(),
		)
	}

	// update snapshot before checking
	executor.UpdateSnapshot(types.BytesToHash(root), objs)

	// check pre snapshot
	parentSnap := snaps.Snapshot(pastRoot)
	if parentSnap == nil {
		t.Fatalf("parent snapshot(%s) not generated", pastRoot)
	}

	for addr, account := range c.Pre {
		addrhash := crypto.Keccak256Hash(addr.Bytes())

		snapAccount, err := parentSnap.Account(addrhash)
		if err != nil {
			t.Fatalf("parent snapshot account(%s) unmarshal failed: %v", addr, err)
		}

		if account.Balance.Cmp(snapAccount.Balance) != 0 {
			t.Fatalf(
				"parent snapshot account(%s) balance not right, want(%s), got(%s)",
				addr,
				account.Balance,
				snapAccount.Balance,
			)
		}
		if len(account.Code) > 0 {
			codeHash := crypto.Keccak256Hash(account.Code)
			if codeHash != types.BytesToHash(snapAccount.CodeHash) {
				t.Fatalf(
					"parent snapshot account(%s) codehash not right, want(%s), got(%s)",
					addr,
					codeHash,
					hex.EncodeToString(snapAccount.CodeHash),
				)
			}
		}
		if account.Nonce != snapAccount.Nonce {
			t.Fatalf(
				"parent snapshot account(%s) nonce not right, want(%d), got(%d)",
				addr,
				account.Nonce,
				snapAccount.Nonce,
			)
		}

		// storage
		for k, v := range account.Storage {
			// query storage no matter exists or not
			sv, err := parentSnap.Storage(addrhash, crypto.Keccak256Hash(k.Bytes()))
			if err != nil {
				t.Fatalf(
					"parent snapshot account(%s) storage(%s) getting failed: %v",
					addr,
					k,
					err,
				)
			}
			// empty hash is "deleted"
			if v == (types.Hash{}) && len(sv) == 0 {
				continue
			}
			// rlp unmarshal
			fv, err := types.RlpUnmarshal(sv)
			if err != nil {
				t.Fatalf(
					"parent snapshot account(%s) storage(%s) unmarshal failed: %v",
					addr,
					k,
					err,
				)
			}
			// fastrlp value
			vv, err := fv.Bytes()
			if err != nil {
				t.Fatalf(
					"parent snapshot account(%s) storage(%s) fastrlp failed: %v",
					addr,
					k,
					err,
				)
			}
			// hash
			hv := types.BytesToHash(vv)
			if v != hv {
				t.Fatalf(
					"parent snapshot account(%s) storage(%s) not right, want(%s), got(%s)",
					addr,
					k,
					v,
					hv,
				)
			}
		}
	}

	// check post snapshot
	snap := snaps.Snapshot(p.Root)
	if snap == nil {
		t.Fatalf("snapshot(%s) not generated", p.Root)
	}

	// check snapshot from account
	from, err := snap.Account(crypto.Keccak256Hash(msg.From.Bytes()))
	if err != nil {
		t.Fatalf("snapshot account(%s) unmarshal failed: %v", msg.From, err)
	}
	if from.Nonce != msg.Nonce+1 {
		t.Fatalf(
			"snapshot account(%s) nonce not right, want(%d), got(%d)",
			msg.From,
			msg.Nonce,
			from.Nonce,
		)
	}
}

func TestState(t *testing.T) {
	t.Parallel()

	long := []string{
		"static_Call50000",
		"static_Return50000",
		"static_Call1MB",
		"stQuadraticComplexityTest",
		"stTimeConsuming",
	}

	skip := []string{
		"RevertPrecompiledTouch",
	}

	// There are two folders in spec tests, one for the current tests for the Istanbul fork
	// and one for the legacy tests for the other forks
	folders, err := listFolders(stateTests, legacyStateTests)
	if err != nil {
		t.Fatal(err)
	}

	for _, folder := range folders {
		files, err := listFiles(folder)
		if err != nil {
			t.Fatal(err)
		}

		for _, file := range files {
			file := file
			t.Run(file, func(t *testing.T) {
				t.Parallel()

				if !strings.HasSuffix(file, ".json") {
					return
				}

				if contains(long, file) && testing.Short() {
					t.Skipf("Long tests are skipped in short mode")

					return
				}

				if contains(skip, file) {
					t.Skip()

					return
				}

				data, err := ioutil.ReadFile(file)
				if err != nil {
					t.Fatal(err)
				}

				var c map[string]*stateCase
				if err := json.Unmarshal(data, &c); err != nil {
					t.Fatal(err)
				}

				for name, i := range c {
					for fork, f := range i.Post {
						for indx, e := range *f {
							RunSpecificTest(t, file, *i, name, fork, indx, *e)
						}
					}
				}
			})
		}
	}
}

func TestStateWithSnapshot(t *testing.T) {
	t.Parallel()

	long := []string{
		"static_Call50000",
		"static_Return50000",
		"static_Call1MB",
		"stQuadraticComplexityTest",
		"stTimeConsuming",
	}

	skip := []string{
		"RevertPrecompiledTouch",
	}

	// There are two folders in spec tests, one for the current tests for the Istanbul fork
	// and one for the legacy tests for the other forks
	folders, err := listFolders(stateTests, legacyStateTests)
	if err != nil {
		t.Fatal(err)
	}

	for _, folder := range folders {
		files, err := listFiles(folder)
		if err != nil {
			t.Fatal(err)
		}

		for _, file := range files {
			file := file
			t.Run(file, func(t *testing.T) {
				t.Parallel()

				if !strings.HasSuffix(file, ".json") {
					return
				}

				if contains(long, file) && testing.Short() {
					t.Skipf("Long tests are skipped in short mode")

					return
				}

				if contains(skip, file) {
					t.Skip()

					return
				}

				data, err := ioutil.ReadFile(file)
				if err != nil {
					t.Fatal(err)
				}

				var c map[string]*stateCase
				if err := json.Unmarshal(data, &c); err != nil {
					t.Fatal(err)
				}

				for name, i := range c {
					for fork, f := range i.Post {
						for indx, e := range *f {
							RunSpecificTestWithSnapshot(t, file, i, name, fork, indx, e)
						}
					}
				}
			})
		}
	}
}
