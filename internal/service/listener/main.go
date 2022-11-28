package listener

import (
	"context"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
	"gitlab.com/distributed_lab/logan/v3"

	uniswapv2pair "github.com/Velnbur/uniswapv2-indexer/generated/uniswapv2-pair"
	"github.com/Velnbur/uniswapv2-indexer/internal/channels"
	"github.com/Velnbur/uniswapv2-indexer/internal/channels/inmemory"
	"github.com/Velnbur/uniswapv2-indexer/internal/config"
	"github.com/Velnbur/uniswapv2-indexer/internal/contracts"
	"github.com/Velnbur/uniswapv2-indexer/internal/providers"
	"github.com/Velnbur/uniswapv2-indexer/internal/providers/redis"
)

type Listener struct {
	client  *ethclient.Client
	logger  *logan.Entry
	pairABI abi.ABI

	uniswapV2 *contracts.UniswapV2

	currentBlock providers.CurrentBlockProvider

	events channels.ReservesUpdateQueue
}

func NewListener(cfg config.Config) (*Listener, error) {
	pairABI, err := abi.JSON(strings.NewReader(
		string(uniswapv2pair.UniswapV2PairABI),
	))
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse pair ABI")
	}

	return &Listener{
		client:       cfg.EthereumClient(),
		logger:       cfg.Log().WithField("service", "listener"),
		pairABI:      pairABI,
		currentBlock: redis.NewBlockProvider(cfg.Redis()),
		events:       inmemory.NewSwapEventChan(),
	}, nil
}

func (l *Listener) Run(ctx context.Context) error {
	if err := l.initContracts(ctx); err != nil {
		return errors.Wrap(err, "failed to init contracts")
	}

	return l.Listen(ctx)
}

func (l *Listener) Listen(ctx context.Context) error {
	query, err := l.filters(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to set subscription filters")
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
			if err := l.handleEvent(ctx, vLog); err != nil {
				l.logger.WithError(err).Error("failed to handle event")
			}
		}
	}
}

func (l *Listener) filters(ctx context.Context) (ethereum.FilterQuery, error) {
	addresses := l.getAllPairsAddresses()

	block, err := l.currentBlock.CurrentBlock(ctx)
	if err != nil {
		return ethereum.FilterQuery{}, errors.Wrap(err,
			"failed to get last block",
		)
	}

	var blockInt *big.Int = nil
	// that means, that we weren't initialized yet,
	// and we need to start indexing from last block
	if block != 0 {
		// other wise, start indexing from last saved block
		blockInt = new(big.Int).SetUint64(block)
	}

	query := ethereum.FilterQuery{
		FromBlock: blockInt,
		Addresses: addresses,
		Topics: [][]common.Hash{
			{
				l.pairABI.Events["Swap"].ID,
				l.pairABI.Events["Sync"].ID,
				l.pairABI.Events["Mint"].ID,
				l.pairABI.Events["Burn"].ID,
			},
		},
	}

	return query, nil
}

func (l *Listener) getAllPairsAddresses() []common.Address {
	var addresses []common.Address
	l.uniswapV2.Pairs.Range(func(addr common.Address, _ *contracts.UniswapV2Pair) bool {
		addresses = append(addresses, addr)
		return true
	})

	return addresses
}

func (l *Listener) handleEvent(ctx context.Context, log types.Log) error {
	if err := l.currentBlock.UpdateBlock(ctx, log.BlockNumber); err != nil {
		return errors.Wrap(err, "failed to update current block number")
	}

	switch log.Topics[0] {
	case l.pairABI.Events["Swap"].ID:
		return l.handleSwap(ctx, log)
	case l.pairABI.Events["Sync"].ID:
		return l.handleSync(ctx, log)
	case l.pairABI.Events["Mint"].ID:
		return l.handleMint(ctx, log)
	case l.pairABI.Events["Burn"].ID:
		return nil
	default:
		return nil
	}
}

func (l *Listener) handleSwap(ctx context.Context, log types.Log) error {
	var event uniswapv2pair.UniswapV2PairSwap
	err := l.pairABI.UnpackIntoInterface(&event, "Swap", log.Data)
	if err != nil {
		return errors.Wrap(err, "failed to unpack Swap event")
	}
	l.logger.WithFields(logan.F{
		"sender":     event.Sender.Hex(),
		"amount0In":  event.Amount0In.String(),
		"amount1In":  event.Amount1In.String(),
		"amount0Out": event.Amount0Out.String(),
		"amount1Out": event.Amount1Out.String(),
		"to":         event.To.Hex(),
	}).Debug("received log")

	err = l.events.Send(ctx, channels.ReservesUpdate{
		Address:       event.Raw.Address,
		Reserve0Delta: &big.Int{},
		Reserve1Delta: &big.Int{},
	})
	return errors.Wrap(err, "failed to add event to queue")
}

func (l *Listener) handleSync(ctx context.Context, log types.Log) error {
	var event uniswapv2pair.UniswapV2PairSync
	err := l.pairABI.UnpackIntoInterface(&event, "Sync", log.Data)
	if err != nil {
		return errors.Wrap(err, "failed to unpack Sync event")
	}
	l.logger.WithFields(logan.F{
		"reserve0": event.Reserve0.String(),
		"reserve1": event.Reserve1.String(),
	}).Debug("received log")

	err = l.events.Send(ctx, channels.ReservesUpdate{
		Address:       event.Raw.Address,
		Reserve0Delta: event.Reserve0,
		Reserve1Delta: event.Reserve1,
	})

	return errors.Wrap(err, "failed to add event to queue")
}

func (l *Listener) handleMint(ctx context.Context, log types.Log) error {
	var event uniswapv2pair.UniswapV2PairMint
	err := l.pairABI.UnpackIntoInterface(&event, "Mint", log.Data)
	if err != nil {
		return errors.Wrap(err, "failed to unpack Mint event")
	}
	l.logger.WithFields(logan.F{
		"sender":  event.Sender.Hex(),
		"amount0": event.Amount0.String(),
		"amount1": event.Amount1.String(),
	}).Debug("received log")

	err = l.events.Send(ctx, channels.ReservesUpdate{
		Address:       event.Raw.Address,
		Reserve0Delta: event.Amount0,
		Reserve1Delta: event.Amount1,
	})

	return errors.Wrap(err, "failed to add event to queue")
}

func (l *Listener) handleBurn(ctx context.Context, log types.Log) error {
	var event uniswapv2pair.UniswapV2PairBurn
	err := l.pairABI.UnpackIntoInterface(&event, "Burn", log.Data)
	if err != nil {
		return errors.Wrap(err, "failed to unpack Mint event")
	}
	l.logger.WithFields(logan.F{
		"sender":  event.Sender.Hex(),
		"amount0": event.Amount0.String(),
		"amount1": event.Amount1.String(),
		"to":      event.To.Hex(),
	}).Debug("received log")

	err = l.events.Send(ctx, channels.ReservesUpdate{
		Address:       event.Raw.Address,
		Reserve0Delta: event.Amount0,
		Reserve1Delta: event.Amount1,
	})

	return errors.Wrap(err, "failed to add event to queue")
}
