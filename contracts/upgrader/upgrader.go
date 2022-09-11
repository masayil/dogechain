package upgrader

import (
	"fmt"
	"math/big"

	"github.com/dogechain-lab/dogechain/chain"
	"github.com/dogechain-lab/dogechain/helper/hex"
	"github.com/dogechain-lab/dogechain/state"
	"github.com/dogechain-lab/dogechain/types"
	"github.com/hashicorp/go-hclog"
)

type UpgradeType uint8

const (
	UpgradeTypeRebalance UpgradeType = iota // only re-balance upgrade due to bridge exception
	UpgradeTypeContract                     // contract code upgrade
)

type UpgradeConfig struct {
	Type         UpgradeType
	Rebalancer   map[types.Address]*big.Int
	ContractAddr types.Address
	CommitURL    string
	Code         string
}

type Upgrade struct {
	UpgradeName string
	Configs     []*UpgradeConfig
}

// type upgradeHook func(blockNumber uint64, contractAddr types.Address, statedb *state.State) error

const (
	MainNetChainID = 2000
)

const (
	mainNet = "Mainnet"
)

var (
	GenesisHash types.Hash
	//upgrade config
	_portlandUpgrade = make(map[string]*Upgrade)
)

func init() {
	_testInt, _ := new(big.Int).SetString("55000000000000000000", 0)

	_portlandUpgrade[mainNet] = &Upgrade{
		UpgradeName: "portland",
		Configs: []*UpgradeConfig{
			{
				Type: UpgradeTypeRebalance,
				Rebalancer: map[types.Address]*big.Int{
					types.StringToAddress("0x1b051e5D1548326284493BfA380E02C3C149Da4E"): _testInt,
					types.StringToAddress("0xa516CF76d083b4cBe93Ebdfb85FbE72aFfFb7a0c"): big.NewInt(0),
					types.StringToAddress("0xC7aD3276180f8dfb628d975473a81Af2836CDf2b"): big.NewInt(0),
					types.StringToAddress("0x521299a363f1847863e4a6c68c91df722d149c3b"): big.NewInt(0),
					types.StringToAddress("0x3a9185A6b49617cC2d5BE65aF199B73f24834F4f"): big.NewInt(0),
				},
			},
		},
	}
}

func UpgradeSystem(
	chainID int,
	config *chain.Forks,
	blockNumber uint64,
	txn *state.Txn,
	logger hclog.Logger,
) {
	if config == nil || blockNumber == 0 || txn == nil {
		return
	}

	var network string

	switch chainID {
	case MainNetChainID:
		fallthrough
	default:
		network = mainNet
	}

	if config.IsOnPortland(blockNumber) { // only upgrade portland once
		up := _portlandUpgrade[network]
		applySystemContractUpgrade(up, blockNumber, txn,
			logger.With("upgrade", up.UpgradeName, "network", network))
	}
}

func applySystemContractUpgrade(upgrade *Upgrade, blockNumber uint64, txn *state.Txn, logger hclog.Logger) {
	if upgrade == nil {
		logger.Info("Empty upgrade config", "height", blockNumber)

		return
	}

	logger.Info(fmt.Sprintf("Apply upgrade %s at height %d", upgrade.UpgradeName, blockNumber))

	for _, cfg := range upgrade.Configs {
		logger.Info(fmt.Sprintf("Upgrade contract %s to commit %s", cfg.ContractAddr.String(), cfg.CommitURL))

		switch cfg.Type {
		case UpgradeTypeRebalance:
			for addr, balance := range cfg.Rebalancer {
				txn.SetBalance(addr, balance)
				logger.Info(fmt.Sprintf("%s upgrade set balance", upgrade.UpgradeName),
					"addr", addr, "balance", balance)
			}
		case UpgradeTypeContract:
			newContractCode, err := hex.DecodeHex(cfg.Code)
			if err != nil {
				panic(fmt.Errorf("failed to decode new contract code: %w", err))
			}

			txn.SetCode(cfg.ContractAddr, newContractCode)
		}
	}
}
