package server

import (
	"net"
	"time"

	"github.com/hashicorp/go-hclog"

	"github.com/dogechain-lab/dogechain/chain"
	"github.com/dogechain-lab/dogechain/network"
	"github.com/dogechain-lab/dogechain/secrets"
)

const (
	DefaultGRPCPort    int = 9632
	DefaultJSONRPCPort int = 8545
	DefaultGraphQLPort int = 9898
	DefaultPprofPort   int = 6060

	TriesInMemory = 128
)

// Config is used to parametrize the minimal client
type Config struct {
	Chain *chain.Chain

	JSONRPC       *JSONRPC
	EnableGraphQL bool
	GraphQL       *GraphQL
	GRPCAddr      *net.TCPAddr
	LibP2PAddr    *net.TCPAddr

	PriceLimit            uint64
	MaxSlots              uint64
	BlockTime             uint64
	PruneTickSeconds      uint64
	PromoteOutdateSeconds uint64

	Telemetry *Telemetry
	Network   *network.Config

	DataDir     string
	RestoreFile *string

	LeveldbOptions *LeveldbOptions

	Seal           bool
	SecretsManager *secrets.SecretsManagerConfig

	LogLevel    hclog.Level
	LogFilePath string

	Daemon       bool
	ValidatorKey string

	BlockBroadcast bool

	CacheConfig *CacheConfig
}

// LeveldbOptions holds the leveldb options
type LeveldbOptions struct {
	CacheSize           int
	Handles             int
	BloomKeyBits        int
	CompactionTableSize int
	CompactionTotalSize int
	NoSync              bool
}

// Telemetry holds the config details for metric services
type Telemetry struct {
	PrometheusAddr  *net.TCPAddr
	EnableIOMetrics bool
}

// JSONRPC holds the config details for the JSON-RPC server
type JSONRPC struct {
	JSONRPCAddr              *net.TCPAddr
	AccessControlAllowOrigin []string
	BatchLengthLimit         uint64
	BlockRangeLimit          uint64
	JSONNamespace            []string
	EnableWS                 bool
	EnablePprof              bool
}

type GraphQL struct {
	GraphQLAddr              *net.TCPAddr
	AccessControlAllowOrigin []string
	BlockRangeLimit          uint64
	EnablePprof              bool
}

// CacheConfig contains the configuration values for the trie database
// that's resident in a blockchain.
type CacheConfig struct {
	TrieCleanLimit     int           // Memory allowance (MB) to use for caching trie nodes in memory
	TrieCleanJournal   string        // Disk journal for saving clean cache entries.
	TrieCleanRejournal time.Duration // Time interval to dump clean cache to disk periodically
	TrieDirtyLimit     int           // Memory limit (MB) at which to start flushing dirty trie nodes to disk
	TrieTimeLimit      time.Duration // Time limit after which to flush the current in-memory trie to disk
	SnapshotLimit      int           // Memory allowance (MB) to use for caching snapshot entries in memory

	SnapshotNoBuild bool // Whether the background generation is allowed
	// Wait for snapshot construction on startup. TODO(karalabe): This is a dirty hack for testing, nuke it
	SnapshotWait bool
}

// defaultCacheConfig are the default caching values if none are specified by the
// user (also used during testing).
var defaultCacheConfig = &CacheConfig{
	TrieCleanLimit: 256,
	TrieDirtyLimit: 256,
	TrieTimeLimit:  5 * time.Minute,
	SnapshotLimit:  256,
	SnapshotWait:   true,
}
