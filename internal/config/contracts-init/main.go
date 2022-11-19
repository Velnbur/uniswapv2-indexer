package contractsinit

import (
	"context"
	"runtime"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"gitlab.com/distributed_lab/logan/v3"

	"github.com/Velnbur/uniswapv2-indexer/internal/contracts"
	workerspool "github.com/Velnbur/uniswapv2-indexer/pkg/workers-pool"
)

type ContractInitializer struct {
	logger *logan.Entry
	client *ethclient.Client
	redis  *redis.Client

	uniswapV2 *contracts.UniswapV2
}

func NewContractInitializer(
	logger *logan.Entry, client *ethclient.Client, redis *redis.Client,
	factoryAddr common.Address,
) (*ContractInitializer, error) {
	ci := &ContractInitializer{
		logger: logger,
		client: client,
		redis:  redis,
	}

	uniswapV2, err := contracts.NewUniswapV2(
		factoryAddr, ci.client, ci.redis, ci.logger,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create UniswapV2 contracts")
	}
	ci.uniswapV2 = uniswapV2

	return ci, nil
}

// FIXME: this is a temporary solution, need to find a better way to handle rate
// limit errors that alchemy makes, when we make too many requests. May be,
// infura has better solutions for that
const errRateLimitStr = "Your app has exceeded its compute units"

func (ci *ContractInitializer) Init(ctx context.Context) (*contracts.UniswapV2, error) {
	amount, err := ci.uniswapV2.Factory.AllPairLength(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get amount of pairs")
	}

	// TODO: may be, make workers amount configurable,
	workingPool := workerspool.NewWorkingPool(runtime.NumCPU(), int64(amount))

	for i := uint64(0); i < amount; i++ {
		index := i // this is necessary to copy value inside closure

		workingPool.AddTask(func(ctx context.Context) error {
			if isCanceled(ctx) {
				return nil
			}
			pair, err := ci.uniswapV2.Factory.AllPairs(ctx, index)
			if err != nil {
				// FIXME: see comment above
				if strings.Contains(err.Error(), errRateLimitStr) {
					return workerspool.RetryError // this makes task to be retried
				}
				return errors.Wrap(err, "failed to get pair address")
			}
			ci.logger.WithFields(logan.F{
				"pair_num": index,
				"address":  pair.Address,
			}).Debug("got pair")

			ci.uniswapV2.Pairs.Set(pair.Address, pair)
			return nil
		})
	}

	if err := workingPool.Run(ctx); err != nil {
		return nil, errors.Wrap(err, "failed to init one of the pairs")
	}
	return ci.uniswapV2, nil
}

// TODO: move this to a separate package
func isCanceled(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}
