package utils

import (
	"os"

	"github.com/tenderly/net-taiko-geth/eth"
	"github.com/tenderly/net-taiko-geth/eth/ethconfig"
	"github.com/tenderly/net-taiko-geth/node"
	"github.com/tenderly/net-taiko-geth/params"
	"github.com/tenderly/net-taiko-geth/rpc"
	"github.com/urfave/cli/v2"
)

var (
	TaikoFlag = cli.BoolFlag{
		Name:  "taiko",
		Usage: "Taiko network",
	}
)

// RegisterTaikoAPIs initializes and registers the Taiko RPC APIs.
func RegisterTaikoAPIs(stack *node.Node, cfg *ethconfig.Config, backend *eth.Ethereum) {
	if os.Getenv("TAIKO_TEST") != "" {
		return
	}
	// Add methods under "taiko_" RPC namespace to the available APIs list
	stack.RegisterAPIs([]rpc.API{
		{
			Namespace: "taiko",
			Version:   params.VersionWithMeta,
			Service:   eth.NewTaikoAPIBackend(backend),
			Public:    true,
		},
		{
			Namespace:     "taikoAuth",
			Version:       params.VersionWithMeta,
			Service:       eth.NewTaikoAuthAPIBackend(backend),
			Authenticated: true,
		},
	})
}
