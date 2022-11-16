package listener

import (
	"context"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
	"gitlab.com/distributed_lab/logan/v3"

	uniswapv2factory "github.com/Velnbur/uniswapv2-indexer/contracts/uniswapv2-factory"
	uniswapv2pair "github.com/Velnbur/uniswapv2-indexer/contracts/uniswapv2-pair"
	"github.com/Velnbur/uniswapv2-indexer/internal/config"
)

type pairKey struct {
	token0 string
	token1 string
}

type Listener struct {
	client *ethclient.Client

	factoryAddress common.Address
	factory        *uniswapv2factory.UniswapV2Factory
	pairs          map[pairKey]*uniswapv2pair.UniswapV2Pair

	events chan Event
	logger *logan.Entry

	wgPairs *sync.WaitGroup
}

func NewListener(cfg config.Config) (*Listener, error) {
	client, err := ethclient.Dial(cfg.EthereumCfg().Node)
	if err != nil {
		return nil, err
	}

	return &Listener{
		client:         client,
		pairs:          make(map[pairKey]*uniswapv2pair.UniswapV2Pair),
		events:         make(chan Event),
		logger:         cfg.Log().WithField("service", "listener"),
		factoryAddress: cfg.ContracterCfg().Factory,
		wgPairs:        new(sync.WaitGroup),
	}, nil
}

func (l *Listener) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)

	if err := l.initialize(ctx, cancel); err != nil {
		return errors.Wrap(err, "failed to initialize pairs")
	}

	for {
		select {
		case <-ctx.Done():
			l.wgPairs.Wait()
			return nil
		case event := <-l.events:
			l.logger.WithFields(logan.F{
				"type": event.Type,
			}).Info("got event")
		}
	}
}

func (l *Listener) Close() {
	l.client.Close()
}

func (l *Listener) initialize(ctx context.Context, cancel context.CancelFunc) error {
	if err := l.initFactory(ctx, cancel, l.factoryAddress); err != nil {
		return errors.Wrap(err, "failed to initialize factory")
	}

	amount, err := l.factory.AllPairsLength(nil)
	if err != nil {
		return errors.Wrap(err, "failed to get amount of pairs")
	}

	for i := big.NewInt(0); i.Cmp(amount) < 0; i.Add(i, big.NewInt(1)) {
		pairAddr, err := l.factory.AllPairs(nil, i)
		if err != nil {
			return errors.Wrap(err, "failed to get pair address")
		}
		if err := l.initPair(ctx, cancel, pairAddr); err != nil {
			return errors.Wrap(err, "failed to initialize pair")
		}
	}
	return nil
}

func (l *Listener) initFactory(
	ctx context.Context, cancel context.CancelFunc, factoryAddr common.Address,
) error {
	factory, err := uniswapv2factory.NewUniswapV2Factory(factoryAddr, l.client)
	if err != nil {
		return errors.Wrap(err, "failed to create factory contract")
	}
	l.logger.WithField(
		"address", factoryAddr.String(),
	).Debug("initialized factory")

	l.factory = factory
	go func() {
		if err := l.listenFactory(ctx); err != nil {
			l.logger.WithError(err).Error("failed to listen factory")
			cancel()
		}
	}()
	return nil
}

func (l *Listener) initPair(
	ctx context.Context, cancel context.CancelFunc, pair common.Address,
) error {
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
	l.pairs[pairKey{token0.String(), token1.String()}] = pairContract

	l.wgPairs.Add(1)
	go func() {
		defer l.wgPairs.Done()
		if err := l.listenPair(ctx, pairContract); err != nil {
			l.logger.WithError(err).Error("failed to listen pair")
			cancel()
		}
	}()
	return nil
}
