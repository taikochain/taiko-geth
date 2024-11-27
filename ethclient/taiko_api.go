package ethclient

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
)

// HeadL1Origin returns the latest L2 block's corresponding L1 origin.
func (ec *Client) HeadL1Origin(ctx context.Context) (*rawdb.L1Origin, error) {
	var res *rawdb.L1Origin

	if err := ec.c.CallContext(ctx, &res, "taiko_headL1Origin"); err != nil {
		return nil, err
	}

	return res, nil
}

// L1OriginByID returns the L2 block's corresponding L1 origin.
func (ec *Client) L1OriginByID(ctx context.Context, blockID *big.Int) (*rawdb.L1Origin, error) {
	var res *rawdb.L1Origin

	if err := ec.c.CallContext(ctx, &res, "taiko_l1OriginByID", hexutil.EncodeBig(blockID)); err != nil {
		return nil, err
	}

	return res, nil
}

// GetSyncMode returns the current sync mode of the L2 node.
func (ec *Client) GetSyncMode(ctx context.Context) (string, error) {
	var res string

	if err := ec.c.CallContext(ctx, &res, "taiko_getSyncMode"); err != nil {
		return "", err
	}

	return res, nil
}

// CHANGE(taiko):
// SubscribeNewSoftBlock subscribes to notifications about the current soft block
// on the given channel.
func (ec *Client) SubscribeNewSoftBlock(ctx context.Context, ch chan<- *types.Block) (ethereum.Subscription, error) {
	sub, err := ec.c.EthSubscribe(ctx, ch, "newSoftBlocks")
	if err != nil {
		return nil, err
	}
	return sub, nil
}
