package network

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/dogechain-lab/dogechain/chain"
	"github.com/dogechain-lab/dogechain/helper/common"
	"github.com/dogechain-lab/dogechain/helper/telemetry"
	"github.com/dogechain-lab/dogechain/helper/tests"
	"github.com/dogechain-lab/dogechain/secrets"
	"github.com/dogechain-lab/dogechain/secrets/local"
	"github.com/hashicorp/go-hclog"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/assert"

	netcommon "github.com/dogechain-lab/dogechain/network/common"
)

const (
	DefaultLeaveTimeout = 30 * time.Second
)

// JoinAndWait is a helper method for joining a destination server
// and waiting for the connection to be successful (destination node is a peer of source)
func JoinAndWait(
	t *testing.T,
	source,
	destination Server,
	connectTimeout time.Duration,
	joinTimeout time.Duration,
	static bool,
) error {
	t.Helper()

	if joinTimeout == 0 {
		joinTimeout = DefaultJoinTimeout
	}

	if connectTimeout < joinTimeout {
		// In case the connect timeout is smaller than the join timeout, align them
		connectTimeout = joinTimeout
	}

	// Mark the destination address as ready for dialing
	source.JoinPeer(netcommon.AddrInfoToString(destination.AddrInfo()), static)

	connectCtx, cancelFn := context.WithTimeout(context.Background(), connectTimeout)
	defer cancelFn()

	// Wait for the peer to be connected
	_, connectErr := WaitUntilPeerConnectsTo(connectCtx, source, destination.AddrInfo().ID)

	t.Logf("JoinAndWait: source server peers: %v\n", source.Peers())
	t.Logf("JoinAndWait: destination server peers: %v\n", destination.Peers())

	return connectErr
}

// JoinAndWait is a helper method to make multiple servers connect to corresponding peer
func JoinAndWaitMultiple(
	t *testing.T,
	timeout time.Duration,
	servers ...Server,
) error {
	t.Helper()

	if len(servers)%2 != 0 {
		return errors.New("number of servers must be even")
	}

	numPair := len(servers) / 2

	var (
		wg    sync.WaitGroup
		errCh = make(chan error, numPair)
	)

	for i := 0; i < len(servers)-1; i += 2 {
		s1, s2 := servers[i], servers[i+1]

		wg.Add(1)

		go func() {
			defer wg.Done()

			errCh <- JoinAndWait(t, s1, s2, timeout, timeout, false)
		}()
	}

	wg.Wait()

	close(errCh)

	for err := range errCh {
		if err != nil {
			return err
		}
	}

	return nil
}

func DisconnectAndWait(
	source Server,
	target peer.ID,
	leaveTimeout time.Duration,
) error {
	if leaveTimeout == 0 {
		leaveTimeout = DefaultLeaveTimeout
	}

	// Mark the destination address as ready for dialing
	source.DisconnectFromPeer(target, "test")

	disconnectCtx, cancelFn := context.WithTimeout(context.Background(), leaveTimeout)
	defer cancelFn()

	// Wait for the peer to be disconnected
	_, err := WaitUntilPeerDisconnectsFrom(disconnectCtx, source, target)

	return err
}

func WaitUntilPeerConnectsTo(ctx context.Context, srv Server, ids ...peer.ID) (bool, error) {
	peersConnected := 0
	targetPeers := len(ids)

	res, err := tests.RetryUntilTimeout(ctx, func() (interface{}, bool) {
		for _, v := range ids {
			if srv.HasPeer(v) {
				peersConnected++
			}

			if peersConnected == targetPeers {
				return true, false
			}
		}

		return nil, true
	})
	if err != nil {
		return false, err
	}

	resVal, ok := res.(bool)
	if !ok {
		return false, errors.New("invalid type assert")
	}

	return resVal, nil
}

func WaitUntilPeerDisconnectsFrom(ctx context.Context, srv Server, ids ...peer.ID) (bool, error) {
	peersDisconnected := 0
	targetPeers := len(ids)

	res, err := tests.RetryUntilTimeout(ctx, func() (interface{}, bool) {
		for _, v := range ids {
			if !srv.HasPeer(v) {
				peersDisconnected++
			}

			if peersDisconnected == targetPeers {
				return true, false
			}
		}

		return nil, true
	})
	if err != nil {
		return false, err
	}

	resVal, ok := res.(bool)
	if !ok {
		return false, errors.New("invalid type assert")
	}

	return resVal, nil
}

// WaitUntilRoutingTableToBeAdded check routing table has given ids and retry by timeout
func WaitUntilRoutingTableToBeFilled(ctx context.Context, srv *DefaultServer, size int) (bool, error) {
	res, err := tests.RetryUntilTimeout(ctx, func() (interface{}, bool) {
		if size == srv.discovery.RoutingTableSize() {
			return true, false
		}

		return false, true
	})
	if err != nil {
		return false, err
	}

	resVal, ok := res.(bool)
	if !ok {
		return false, errors.New("invalid type assert")
	}

	return resVal, nil
}

// constructMultiAddrs is a helper function for converting raw IPs to mutliaddrs
func constructMultiAddrs(addresses []string) ([]multiaddr.Multiaddr, error) {
	returnAddrs := make([]multiaddr.Multiaddr, 0)

	for _, addr := range addresses {
		multiAddr, err := multiaddr.NewMultiaddr(addr)
		if err != nil {
			return nil, err
		}

		returnAddrs = append(returnAddrs, multiAddr)
	}

	return returnAddrs, nil
}

func createServers(
	count int,
	paramsMap map[int]*CreateServerParams,
) ([]*DefaultServer, error) {
	servers := make([]*DefaultServer, count)

	if paramsMap == nil {
		paramsMap = map[int]*CreateServerParams{}
	}

	for i := 0; i < count; i++ {
		server, createErr := CreateServer(paramsMap[i])
		if createErr != nil {
			return nil, createErr
		}

		servers[i] = server
	}

	return servers, nil
}

type CreateServerParams struct {
	ConfigCallback func(c *Config)             // Additional logic that needs to be executed on the configuration
	ServerCallback func(server *DefaultServer) // Additional logic that needs to be executed on the server before starting
	Logger         hclog.Logger
}

var (
	emptyParams = &CreateServerParams{}
)

// initBootnodes is a helper method for specifying the server's bootnode configuration
func initBootnodes(server *DefaultServer, bootnodes ...string) {
	savedBootnodes := bootnodes
	if len(savedBootnodes) == 0 {
		// Set the default bootnode to be the server itself
		savedBootnodes = []string{
			fmt.Sprintf(
				"%s/p2p/%s",
				server.addrs[0].String(),
				server.host.ID().String(),
			),
		}
	}

	server.config.Chain.Bootnodes = savedBootnodes
}

// Avoid interference between parallel tests on a single machine
var (
	randomChainID int
)

func init() {
	seed := time.Now().UnixNano()
	//#nosec G404
	r := rand.New(rand.NewSource(seed))

	randomChainID = r.Int()
}

func CreateServer(params *CreateServerParams) (*DefaultServer, error) {
	cfg := DefaultConfig()
	port, portErr := tests.GetFreePort()

	// fix join timeout, but allow for tests to override
	cfg.MaxInboundPeers = 1024
	cfg.MaxOutboundPeers = 1024

	if portErr != nil {
		return nil, fmt.Errorf("unable to fetch free port, %w", portErr)
	}

	cfg.Addr.Port = port
	cfg.Chain = &chain.Chain{
		Params: &chain.Params{
			ChainID: randomChainID,
		},
	}

	if params == nil {
		params = emptyParams
	}

	if params.ConfigCallback != nil {
		params.ConfigCallback(cfg)
	}

	if params.Logger == nil {
		params.Logger = hclog.NewNullLogger()
	}

	secretsManager, factoryErr := local.SecretsManagerFactory(
		nil,
		&secrets.SecretsManagerParams{
			Logger: params.Logger,
			Extra: map[string]interface{}{
				secrets.Path: cfg.DataDir,
			},
		},
	)
	if factoryErr != nil {
		return nil, factoryErr
	}

	cfg.SecretsManager = secretsManager
	cfg.Metrics = NilMetrics()

	tracerProvider := telemetry.NewNilTracerProvider(context.Background())

	server, err := newServer(context.Background(), params.Logger, tracerProvider.NewTracer("test"), cfg)
	if err != nil {
		return nil, err
	}

	initBootnodes(server)

	if params.ServerCallback != nil {
		params.ServerCallback(server)
	}

	startErr := server.Start()

	return server, startErr
}

// MeshJoin is a helper method for joining all the passed in servers into a mesh
func MeshJoin(t *testing.T, servers ...*DefaultServer) []error {
	t.Helper()

	if len(servers) < 2 {
		return nil
	}

	// Join errors are used to gather all errors that happen
	// inside the go routines, so they can be handled when they finish
	joinErrors := make([]error, 0)

	var joinErrorsLock sync.Mutex

	appendJoinError := func(joinErr error) {
		joinErrorsLock.Lock()
		joinErrors = append(joinErrors, joinErr)
		joinErrorsLock.Unlock()
	}

	numServers := len(servers)

	var wg sync.WaitGroup

	for indx := 0; indx < numServers; indx++ {
		for innerIndx := 0; innerIndx < numServers; innerIndx++ {
			if innerIndx > indx {
				wg.Add(1)

				go func(src, dest int) {
					defer wg.Done()

					if joinErr := JoinAndWait(
						t,
						servers[src],
						servers[dest],
						DefaultBufferTimeout,
						DefaultJoinTimeout,
						false,
					); joinErr != nil {
						appendJoinError(fmt.Errorf("unable to join peers, %w", joinErr))
					}
				}(indx, innerIndx)
			}
		}
	}

	wg.Wait()

	return joinErrors
}

func GenerateTestLibp2pKey(t *testing.T) (crypto.PrivKey, string) {
	t.Helper()

	dir, err := ioutil.TempDir(os.TempDir(), "")
	assert.NoError(t, err)

	// Instantiate the correct folder structure
	setupErr := common.SetupDataDir(dir, []string{"libp2p"})
	if setupErr != nil {
		t.Fatalf("unable to generate libp2p folder structure, %v", setupErr)
	}

	localSecretsManager, factoryErr := local.SecretsManagerFactory(
		nil,
		&secrets.SecretsManagerParams{
			Logger: hclog.NewNullLogger(),
			Extra: map[string]interface{}{
				secrets.Path: dir,
			},
		})
	assert.NoError(t, factoryErr)

	libp2pKey, libp2pKeyEncoded, keyErr := GenerateAndEncodeLibp2pKey()
	if keyErr != nil {
		t.Fatalf("unable to generate libp2p key, %v", keyErr)
	}

	if setErr := localSecretsManager.SetSecret(secrets.NetworkKey, libp2pKeyEncoded); setErr != nil {
		t.Fatalf("unable to save libp2p key, %v", setErr)
	}

	t.Cleanup(func() {
		// remove directory after test is done
		_ = os.RemoveAll(dir)
	})

	return libp2pKey, dir
}

func closeTestServers(t *testing.T, servers []*DefaultServer) {
	t.Helper()

	for _, server := range servers {
		assert.NoError(t, server.Close())
	}
}
