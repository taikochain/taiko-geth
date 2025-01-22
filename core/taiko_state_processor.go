package core

import (
	"context"
	"errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"
)

// ApplyTransactionWithContext applies a transaction to the current state with the given context.
// It converts the transaction to a message, creates a new EVM environment, and executes the transaction.
//
// Parameters:
//   - ctx: The context to control the execution timeout and cancellation.
//   - hashFuncWrapper: A function to wrap the block's GetHash function.
//   - config: The chain configuration parameters.
//   - bc: The blockchain context.
//   - author: The address of the block author.
//   - gp: The gas pool for the current block.
//   - statedb: The state database.
//   - header: The block header.
//   - tx: The transaction to be applied.
//   - usedGas: A pointer to the amount of gas used.
//   - cfg: The EVM configuration.
//
// Returns:
//   - *types.Receipt: The receipt of the transaction.
//   - error: An error if the transaction could not be applied.
func ApplyTransactionWithContext(ctx context.Context, hashFuncWrapper func(vm.GetHashFunc) vm.GetHashFunc, config *params.ChainConfig, bc ChainContext, author *common.Address, gp *GasPool, statedb *state.StateDB, header *types.Header, tx *types.Transaction, usedGas *uint64, cfg vm.Config) (*types.Receipt, error) {
	msg, err := TransactionToMessage(tx, types.MakeSigner(config, header.Number, header.Time), header.BaseFee)
	if err != nil {
		return nil, err
	}
	// CHANGE(taiko): decode the basefeeSharingPctg config from the extradata, and
	// add it to the Message, if its an ontake block.
	if config.IsOntake(header.Number) {
		msg.BasefeeSharingPctg = DecodeOntakeExtraData(header.Extra)
	}
	// Create a new context to be used in the EVM environment
	blockContext := NewEVMBlockContext(header, bc, author)
	blockContext.GetHash = hashFuncWrapper(blockContext.GetHash)
	txContext := NewEVMTxContext(msg)
	vmenv := vm.NewEVM(blockContext, txContext, statedb, config, cfg)
	go func() {
		<-ctx.Done()
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			// Stop evm execution. Note cancellation is not necessarily immediate.
			vmenv.Cancel()
		}
	}()
	return ApplyTransactionWithEVM(msg, config, gp, statedb, header.Number, header.Hash(), tx, usedGas, vmenv)
}
