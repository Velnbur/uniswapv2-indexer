package listener

import (
	"context"
	"runtime"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
	"gitlab.com/distributed_lab/logan/v3"

	uniswapv2pair "github.com/Velnbur/uniswapv2-indexer/contracts/uniswapv2-pair"
	"github.com/Velnbur/uniswapv2-indexer/internal/config"
	"github.com/Velnbur/uniswapv2-indexer/internal/contracts"
)

type Listener struct {
	client *ethclient.Client
	logger *logan.Entry

	factory *contracts.UniswapV2Factory
	pairs   *PairsMap

	swapEvents chan *uniswapv2pair.UniswapV2PairSwap
}

func NewListener(cfg config.Config) (*Listener, error) {
	client, err := ethclient.Dial(cfg.EthereumCfg().Node)
	if err != nil {
		return nil, err
	}

	factory, err := contracts.NewUniswapV2Factory(
		cfg.ContracterCfg().Factory, client, cfg.Redis(),
	)
	if err != nil {
		return nil, err
	}

	return &Listener{
		client:     client,
		logger:     cfg.Log().WithField("service", "listener"),
		factory:    factory,
		pairs:      NewPairsMap(),
		swapEvents: make(chan *uniswapv2pair.UniswapV2PairSwap),
	}, nil
}

func (l *Listener) Run(ctx context.Context) error {
	err := l.initialize(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to initialize pairs")
	}

	tokenSwapSig := []byte("Swap(address,uint256,uint256,uint256,uint256,address)")
	tokenSwapSigHash := crypto.Keccak256Hash(tokenSwapSig)

	contractAbi, err := abi.JSON(strings.NewReader(
		string(uniswapv2pair.UniswapV2PairABI),
	))
	if err != nil {
		return errors.Wrap(err, "failed to parse pair ABI")
	}

	addresses := make([]common.Address, 0, l.pairs.Len())
	l.pairs.Range(func(addr common.Address, _ *contracts.UniswapV2Pair) bool {
		addresses = append(addresses, addr)
		return true
	})

	query := ethereum.FilterQuery{
		Addresses: addresses,
		Topics: [][]common.Hash{
			{tokenSwapSigHash},
		},
	}

	logs := make(chan types.Log)
	sub, err := l.client.SubscribeFilterLogs(ctx, query, logs)
	if err != nil {
		return errors.Wrap(err, "failed to subscribe to logs")
	}
	defer sub.Unsubscribe()

	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-sub.Err():
			return errors.Wrap(err, "failed to subscribe to logs")
		case vLog := <-logs:
			var event uniswapv2pair.UniswapV2PairSwap

			err := contractAbi.UnpackIntoInterface(&event, "Swap", vLog.Data)
			if err != nil {
				l.logger.WithError(err).Error("failed to unpack Swap event")
				continue
			}
			l.logger.WithField("log", vLog).Debug("received log")
		}
	}
}

func (l *Listener) Close() {
	l.client.Close()
}

const errRateLimitStr = "Your app has exceeded its compute units per second capacity"

func (l *Listener) initialize(ctx context.Context) error {
	amount, err := l.factory.AllPairLength(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get amount of pairs")
	}

	workingPool := NewWorkingPool(runtime.NumCPU(), int64(amount))

	for i := uint64(0); i < amount; i++ {
		index := i
		workingPool.AddTask(func(ctx context.Context) error {
			if isCanceled(ctx) {
				return nil
			}
			pair, err := l.factory.AllPairs(ctx, index)
			if err != nil {
				if strings.Contains(err.Error(), errRateLimitStr) {
					return RetryError
				}
				return errors.Wrap(err, "failed to get pair address")
			}
			l.logger.WithFields(logan.F{
				"pair_num":  index,
				"pair_addr": pair.Address,
			}).Debug("got pair address")

			l.pairs.Set(pair.Address, pair)
			return nil
		})
	}

	if err := workingPool.Run(ctx); err != nil {
		return errors.Wrap(err, "failed to init one of the pairs")
	}
	return nil
}

func isCanceled(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}
