package tracers

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/internal/ethapi"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"
)

var noopTracer = "noopTracer"

// provingPreflightResult is the result of a proving preflight request.
type provingPreflightResult struct {
	Block               *types.Block                   `json:"block"`
	ParentHeader        *types.Header                  `json:"parentHeader"`
	AccountProofs       []*ethapi.AccountResult        `json:"accountProofs"`
	ParentAccountProofs []*ethapi.AccountResult        `json:"parentAccountProofs"`
	Contracts           map[common.Hash]*hexutil.Bytes `json:"contracts"`
	AncestorHeaders     []*types.Header                `json:"ancestorHeaders"`
}

// ProvingPreflights traces the blockchain from the start block to the end block
// and returns a subscription to the results. This function is intended to be
// used with long-running operations, as tracing a chain can be time-consuming.
//
// Parameters:
// - ctx: The context for the operation.
// - start: The starting block number for the trace.
// - end: The ending block number for the trace.
// - config: The configuration for the trace.
//
// Returns:
// - *rpc.Subscription: A subscription to the trace results.
// - error: An error if the operation fails, or if the end block is not after the start block.
//
// Note:
//   - The function requires a notifier to be present in the context, otherwise it
//     returns an error indicating that notifications are unsupported.
func (api *API) ProvingPreflights(ctx context.Context, start, end rpc.BlockNumber, config *TraceConfig) (*rpc.Subscription, error) {
	from, err := api.blockByNumber(ctx, start)
	if err != nil {
		return nil, err
	}
	to, err := api.blockByNumber(ctx, end)
	if err != nil {
		return nil, err
	}
	if from.Number().Cmp(to.Number()) >= 0 {
		return nil, fmt.Errorf("end block (#%d) needs to come after start block (#%d)", end, start)
	}
	// Tracing a chain is a **long** operation, only do with subscriptions
	notifier, supported := rpc.NotifierFromContext(ctx)
	if !supported {
		return &rpc.Subscription{}, rpc.ErrNotificationsUnsupported
	}
	sub := notifier.CreateSubscription()

	resCh := api.provingPreflights(from, to, config, sub.Err())
	go func() {
		for result := range resCh {
			notifier.Notify(sub.ID, result)
		}
	}()
	return sub, nil
}

func (api *API) provingPreflights(start, end *types.Block, config *TraceConfig, closed <-chan error) chan *provingPreflightResult {
	reexec := defaultTraceReexec
	if config != nil && config.Reexec != nil {
		reexec = *config.Reexec
	}
	blocks := int(end.NumberU64() - start.NumberU64())
	threads := runtime.NumCPU()
	if threads > blocks {
		threads = blocks
	}
	// no need any execution traces when preflighting
	if config == nil {
		config = &TraceConfig{}
	}
	config.Tracer = &noopTracer
	var (
		pend    = new(sync.WaitGroup)
		ctx     = context.Background()
		taskCh  = make(chan *blockTraceTask, threads)
		resCh   = make(chan *blockTraceTask, threads)
		tracker = newStateTracker(maximumPendingTraceStates, start.NumberU64())
	)
	for th := 0; th < threads; th++ {
		pend.Add(1)
		go func() {
			defer pend.Done()

			// Fetch and execute the block trace taskCh
			for task := range taskCh {
				var (
					signer   = types.MakeSigner(api.backend.ChainConfig(), task.block.Number(), task.block.Time())
					blockCtx = core.NewEVMBlockContext(task.block.Header(), api.chainContext(ctx), nil)
				)
				// Trace all the transactions contained within
				for i, tx := range task.block.Transactions() {
					if i == 0 && api.backend.ChainConfig().Taiko {
						if err := tx.MarkAsAnchor(); err != nil {
							log.Warn("Mark anchor transaction error", "error", err)
							task.results[i] = &txTraceResult{TxHash: tx.Hash(), Error: err.Error()}
							break
						}
					}
					msg, _ := core.TransactionToMessage(tx, signer, task.block.BaseFee())
					txctx := &Context{
						BlockHash:   task.block.Hash(),
						BlockNumber: task.block.Number(),
						TxIndex:     i,
						TxHash:      tx.Hash(),
					}
					res, err := api.traceTx(ctx, tx, msg, txctx, blockCtx, task.statedb, config)
					if err != nil {
						task.results[i] = &txTraceResult{TxHash: tx.Hash(), Error: err.Error()}
						log.Warn("Tracing failed", "hash", tx.Hash(), "block", task.block.NumberU64(), "err", err)
						break
					}
					task.results[i] = &txTraceResult{TxHash: tx.Hash(), Result: res}
				}
				// Tracing state is used up, queue it for de-referencing. Note the
				// state is the parent state of trace block, use block.number-1 as
				// the state number.
				tracker.releaseState(task.block.NumberU64()-1, task.release)

				// Stream the result back to the result catcher or abort on teardown
				select {
				case resCh <- task:
				case <-closed:
					return
				}
			}
		}()
	}
	// Start a goroutine to feed all the blocks into the tracers
	go func() {
		var (
			logged  time.Time
			begin   = time.Now()
			number  uint64
			traced  uint64
			failed  error
			statedb *state.StateDB
			release StateReleaseFunc
		)
		// Ensure everything is properly cleaned up on any exit path
		defer func() {
			close(taskCh)
			pend.Wait()

			// Clean out any pending release functions of trace states.
			tracker.callReleases()

			// Log the chain result
			switch {
			case failed != nil:
				log.Warn("Chain tracing failed", "start", start.NumberU64(), "end", end.NumberU64(), "transactions", traced, "elapsed", time.Since(begin), "err", failed)
			case number < end.NumberU64():
				log.Warn("Chain tracing aborted", "start", start.NumberU64(), "end", end.NumberU64(), "abort", number, "transactions", traced, "elapsed", time.Since(begin))
			default:
				log.Info("Chain tracing finished", "start", start.NumberU64(), "end", end.NumberU64(), "transactions", traced, "elapsed", time.Since(begin))
			}
			close(resCh)
		}()
		// Feed all the blocks both into the tracer, as well as fast process concurrently
		for number = start.NumberU64(); number < end.NumberU64(); number++ {
			// Stop tracing if interruption was requested
			select {
			case <-closed:
				return
			default:
			}
			// Print progress logs if long enough time elapsed
			if time.Since(logged) > 8*time.Second {
				logged = time.Now()
				log.Info("Tracing chain segment", "start", start.NumberU64(), "end", end.NumberU64(), "current", number, "transactions", traced, "elapsed", time.Since(begin))
			}
			// Retrieve the parent block and target block for tracing.
			block, err := api.blockByNumber(ctx, rpc.BlockNumber(number))
			if err != nil {
				failed = err
				break
			}
			next, err := api.blockByNumber(ctx, rpc.BlockNumber(number+1))
			if err != nil {
				failed = err
				break
			}
			// Make sure the state creator doesn't go too far. Too many unprocessed
			// trace state may cause the oldest state to become stale(e.g. in
			// path-based scheme).
			if err = tracker.wait(number); err != nil {
				failed = err
				break
			}
			// Prepare the statedb for tracing. Don't use the live database for
			// tracing to avoid persisting state junks into the database. Switch
			// over to `preferDisk` mode only if the memory usage exceeds the
			// limit, the trie database will be reconstructed from scratch only
			// if the relevant state is available in disk.
			var preferDisk bool
			if statedb != nil {
				s1, s2, s3 := statedb.Database().TrieDB().Size()
				preferDisk = s1+s2+s3 > defaultTracechainMemLimit
			}
			statedb, release, err = api.backend.StateAtBlock(ctx, block, reexec, statedb, false, preferDisk)
			if err != nil {
				failed = err
				break
			}
			// Insert block's parent beacon block root in the state
			// as per EIP-4788.
			if beaconRoot := next.BeaconRoot(); beaconRoot != nil {
				context := core.NewEVMBlockContext(next.Header(), api.chainContext(ctx), nil)
				vmenv := vm.NewEVM(context, vm.TxContext{}, statedb, api.backend.ChainConfig(), vm.Config{})
				core.ProcessBeaconBlockRoot(*beaconRoot, vmenv, statedb)
			}
			// Insert parent hash in history contract.
			if api.backend.ChainConfig().IsPrague(next.Number(), next.Time()) {
				context := core.NewEVMBlockContext(next.Header(), api.chainContext(ctx), nil)
				vmenv := vm.NewEVM(context, vm.TxContext{}, statedb, api.backend.ChainConfig(), vm.Config{})
				core.ProcessParentBlockHash(next.ParentHash(), vmenv, statedb)
			}
			// Clean out any pending release functions of trace state. Note this
			// step must be done after constructing tracing state, because the
			// tracing state of block next depends on the parent state and construction
			// may fail if we release too early.
			tracker.callReleases()

			// Send the block over to the concurrent tracers (if not in the fast-forward phase)
			txs := next.Transactions()
			select {
			case taskCh <- &blockTraceTask{statedb: statedb.Copy(), block: next, release: release, results: make([]*txTraceResult, len(txs))}:
			case <-closed:
				tracker.releaseState(number, release)
				return
			}
			traced += uint64(len(txs))
		}
	}()

	// Keep reading the trace results and stream them to result channel.
	retCh := make(chan *blockTraceResult)
	go func() {
		defer close(retCh)
		var (
			next = start.NumberU64() + 1
			done = make(map[uint64]*blockTraceResult)
		)
		for res := range resCh {
			// Queue up next received result
			result := &blockTraceResult{
				Block:  hexutil.Uint64(res.block.NumberU64()),
				Hash:   res.block.Hash(),
				Traces: res.results,
			}
			done[uint64(result.Block)] = result

			// Stream completed traces to the result channel
			for result, ok := done[next]; ok; result, ok = done[next] {
				if len(result.Traces) > 0 || next == end.NumberU64() {
					// It will be blocked in case the channel consumer doesn't take the
					// tracing result in time(e.g. the websocket connect is not stable)
					// which will eventually block the entire chain tracer. It's the
					// expected behavior to not waste node resources for a non-active user.
					retCh <- result
				}
				delete(done, next)
				next++
			}
		}
	}()
	return retCh
}
