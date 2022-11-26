package listener

import (
	"context"
	"runtime"
	"strings"

	"github.com/pkg/errors"
	"gitlab.com/distributed_lab/logan/v3"

	"github.com/Velnbur/uniswapv2-indexer/pkg/helpers"
	workerspool "github.com/Velnbur/uniswapv2-indexer/pkg/workers-pool"
)

// FIXME: this is a temporary solution, need to find a better way to handle rate
// limit errors that alchemy makes, when we make too many requests. May be,
// infura has better solutions for that
const errRateLimitStr = "Your app has exceeded its compute units"

func (l *Listener) initContracts(ctx context.Context) error {
	amount, err := l.uniswapV2.Factory.AllPairLength(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get amount of pairs")
	}

	// TODO: may be, make workers amount configurable,
	workingPool := workerspool.NewWorkingPool(runtime.NumCPU(), int64(amount))

	for i := uint64(0); i < amount; i++ {
		index := i // this is necessary to copy value inside closure

		workingPool.AddTask(func(ctx context.Context) error {
			if helpers.IsCanceled(ctx) {
				return nil
			}
			pair, err := l.uniswapV2.Factory.AllPairs(ctx, index)
			if err != nil {
				// FIXME: see comment above
				if strings.Contains(err.Error(), errRateLimitStr) {
					return workerspool.RetryError // this makes task to be retried
				}
				return errors.Wrap(err, "failed to get pair address")
			}
			l.logger.WithFields(logan.F{
				"pair_num": index,
				"address":  pair.Address,
			}).Debug("got pair")

			l.uniswapV2.Pairs.Set(pair.Address, pair)
			return nil
		})
	}

	if err := workingPool.Run(ctx); err != nil {
		return errors.Wrap(err, "failed to init one of the pairs")
	}
	return nil
}
