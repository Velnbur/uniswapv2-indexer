package listener

import (
	"context"
	"math/big"
	"runtime"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
	"gitlab.com/distributed_lab/logan/v3"

	uniswapv2factory "github.com/Velnbur/uniswapv2-indexer/contracts/uniswapv2-factory"
	uniswapv2pair "github.com/Velnbur/uniswapv2-indexer/contracts/uniswapv2-pair"
	"github.com/Velnbur/uniswapv2-indexer/internal/config"
)

type Listener struct {
	client *ethclient.Client

	factoryAddress common.Address
	factory        *uniswapv2factory.UniswapV2Factory
	pairs          *PairsMap

	logger *logan.Entry

	wgPairs          *sync.WaitGroup
	swapEvents       chan *uniswapv2pair.UniswapV2PairSwap
	createPairEvents chan *uniswapv2factory.UniswapV2FactoryPairCreated
}

func NewListener(cfg config.Config) (*Listener, error) {
	client, err := ethclient.Dial(cfg.EthereumCfg().Node)
	if err != nil {
		return nil, err
	}

	return &Listener{
		client:         client,
		pairs:          NewPairsMap(),
		swapEvents:     make(chan *uniswapv2pair.UniswapV2PairSwap),
		logger:         cfg.Log().WithField("service", "listener"),
		factoryAddress: cfg.ContracterCfg().Factory,
		wgPairs:        new(sync.WaitGroup),
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
	l.pairs.Range(func(addr common.Address, _ *Pair) bool {
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
			l.wgPairs.Wait()
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
	if err := l.initFactory(ctx, l.factoryAddress); err != nil {
		return errors.Wrap(err, "failed to initialize factory")
	}

	amount, err := l.factory.AllPairsLength(nil)
	if err != nil {
		return errors.Wrap(err, "failed to get amount of pairs")
	}

	amountInt := amount.Int64()
	workingPool := NewWorkingPool(runtime.NumCPU(), amountInt)

	for i := int64(0); i < amountInt; i++ {
		index := i
		workingPool.AddTask(func(ctx context.Context) error {
			if isCanceled(ctx) {
				return nil
			}
			pairAddr, err := l.factory.AllPairs(nil, big.NewInt(index))
			if err != nil {
				if strings.Contains(err.Error(), errRateLimitStr) {
					return RetryError
				}
				return errors.Wrap(err, "failed to get pair address")
			}
			l.logger.WithFields(logan.F{
				"pair_num":  index,
				"pair_addr": pairAddr,
			}).Debug("got pair address")

			if err := l.initPair(ctx, pairAddr); err != nil {
				if strings.Contains(err.Error(), errRateLimitStr) {
					return RetryError
				}
				return errors.Wrap(err, "failed to initialize pair")
			}
			l.logger.WithFields(logan.F{
				"pair": pairAddr.Hex(),
			}).Debug("initialized pair")
			return nil
		})
	}

	if err := workingPool.Run(ctx); err != nil {
		return errors.Wrap(err, "failed to init one of the pairs")
	}
	return nil
}

func (l *Listener) initFactory(ctx context.Context, factoryAddr common.Address) error {
	factory, err := uniswapv2factory.NewUniswapV2Factory(factoryAddr, l.client)
	if err != nil {
		return errors.Wrap(err, "failed to create factory contract")
	}
	l.logger.WithField(
		"address", factoryAddr.String(),
	).Debug("initialized factory")

	l.factory = factory
	return nil
}

func (l *Listener) initPair(ctx context.Context, pair common.Address) error {
	pairContract, err := uniswapv2pair.NewUniswapV2Pair(pair, l.client)
	if err != nil {
		return errors.Wrap(err, "failed to create pair contract")
	}

	token0, err := pairContract.Token0(nil)
	if err != nil {
		return errors.Wrap(err, "failed to get token0 address")
	}
	token1, err := pairContract.Token1(nil)
	if err != nil {
		return errors.Wrap(err, "failed to get token1 address")
	}

	l.pairs.Set(pair, &Pair{
		Address:  pair,
		Token0:   token0,
		Token1:   token1,
		Contract: pairContract,
	})
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
