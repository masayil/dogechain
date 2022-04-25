package chain

import (
	"encoding/json"
	"fmt"

	"github.com/dogechain-lab/jury/helper/hex"
	"github.com/dogechain-lab/jury/types"
)

// Some Genesis common constant
const (
	defaultBlockGasLimit = 30000000 // 0x1c9c380
	defaultEpochSize     = 7200
	defaultConsensusType = "ibft"
	defaultIBFTConsensus = "PoS"
)

// Genesis hashes to enforce below configs on.
var (
	MainnetGenesisHash = types.StringToHash("0x8318ce1509497cb025cf4c9455a13274bdfb4ce391350bdb8d4a493c07d8bee4")
	TeddyGenesisHash   = types.StringToHash("0x5cec6fdbba607e2c9ccaa54aa1065c45731d92170dd8a631e0507e1b9dde75b4")
)

// Genesis share the same engine config
var (
	defaultEngineConf = map[string]interface{}{
		defaultConsensusType: map[string]interface{}{
			"epochSize": defaultEpochSize,
			"type":      defaultConsensusType,
		},
	}
)

var (
	// nolint: lll
	defaultMainnetExtraData, _ = hex.DecodeString("0x0000000000000000000000000000000000000000000000000000000000000000f858f854945ede2f630adc5e5ab5fd9e5c7493b9296c420d99941087a8d89405baceb185b5ef7645480c725fdf8d94a986753ee853bb869f82e7f4e611946b9a79d0589420a58cfa71e17168d437b551b9373ffc2a8ccbda80c0")
	// nolint: lll
	defaultTeddyExtraData, _ = hex.DecodeString("0x0000000000000000000000000000000000000000000000000000000000000000f858f8549421da36cb4fe7b11ee84c560084c45c0e8886e42b94b113213506bfc320ad0ab9f2992d3129224b22b8947945eb690589dcfd424457e030234ad8f7204fc494a2ad008adedbe39b2b1b06fe5040e11bef0c26af80c0")
)

var (
	// MainnetChainConfig is the chain parameters to run a node on the main network.
	// nolint: lll
	MainnetChainConfig = &Chain{
		Name: "DogeChain",
		Genesis: &Genesis{
			ExtraData:  defaultMainnetExtraData,
			GasLimit:   defaultBlockGasLimit,
			Difficulty: 1,
			Alloc:      defaultMainnetAlloc,
			StateRoot:  types.StringToHash("0x47a93367b26beff86279ce341b693a9ce654413c426b43fb0af76667bd1648f3"),
			GasUsed:    458752,
		},
		Params: &Params{
			ChainID: 2000,
			Engine:  defaultEngineConf,
		},
	}

	// TeddyChainConfig is the chain parameters to run a node on the Teddy test network.
	// nolint: lll
	TeddyChainConfig = &Chain{
		Name: "Teddy",
		Genesis: &Genesis{
			ExtraData:  defaultTeddyExtraData,
			GasLimit:   defaultBlockGasLimit,
			Difficulty: 1,
			Alloc:      defaultTeddyAlloc,
			StateRoot:  types.StringToHash("0x47a93367b26beff86279ce341b693a9ce654413c426b43fb0af76667bd1648f3"),
			GasUsed:    458752,
			ParentHash: types.ZeroHash,
		},
		Params: &Params{
			ChainID: 568,
			Engine:  defaultEngineConf,
		},
	}
)

// GenesisAlloc is the initial account state of the genesis block.
type GenesisAlloc map[types.Address]*GenesisAccount

// DecodePrealloc decodes geneis alloc account
func DecodePrealloc(hexData string) GenesisAlloc {
	data, err := hex.DecodeString(hexData)
	if err != nil {
		panic(fmt.Errorf("genesis alloc decode failed: %v", err))
	}

	alloc := make(GenesisAlloc)
	if err := json.Unmarshal(data, &alloc); err != nil {
		panic(fmt.Errorf("genesis alloc decode failed: %v", err))
	}

	return alloc
}

// EncodePrealloc encodes geneis alloc account
func EncodePrealloc(alloc GenesisAlloc) (string, error) {
	data, err := json.Marshal(alloc)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(data), nil
}
