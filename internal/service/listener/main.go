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
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"

	uniswapv2pair "github.com/Velnbur/uniswapv2-indexer/generated/uniswapv2-pair"
	"github.com/Velnbur/uniswapv2-indexer/internal/channels"
	"github.com/Velnbur/uniswapv2-indexer/internal/config"
	"github.com/Velnbur/uniswapv2-indexer/internal/contracts"
	"github.com/Velnbur/uniswapv2-indexer/internal/providers"
)

type Listener struct {
	client  *ethclient.Client
	logger  *logan.Entry
	pairABI abi.ABI

	uniswapV2 *contracts.UniswapV2

	currentBlock providers.CurrentBlockProvider

	reservesUpdateEvents channels.ReservesUpdateQueue
	pairCreationEvents   channels.PairCreationQueue

	eventHandlers map[common.Hash]EventHandler
	eventUnpacker *EventUnpacker
}

func NewListener(cfg config.Config) (*Listener, error) {
	pairABI, err := abi.JSON(strings.NewReader(
		string(uniswapv2pair.UniswapV2PairABI),
	))
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse pair ABI")
	}

	listener := &Listener{
		client:               cfg.EthereumClient(),
		logger:               cfg.Log().WithField("service", "listener"),
		pairABI:              pairABI,
		currentBlock:         providers.NewBlockProvider(cfg.Redis()),
		reservesUpdateEvents: channels.NewReservesUpdateChan(),
		eventUnpacker:        NewEventUnpacker(&pairABI),
	}
	listener.initHandlers(pairABI)

	return listener, nil
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
			if err := l.handleEvent(ctx, &vLog); err != nil {
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

	topics := make([]common.Hash, 0)
	for _, event := range AllEvents() {
		topic, ok := l.pairABI.Events[string(event)]
		if !ok {
			return ethereum.FilterQuery{}, errors.Wrap(err,
				"no such event in abi",
				logan.F{
					"event": event,
				})
		}
		topics = append(topics, topic.ID)
	}

	query := ethereum.FilterQuery{
		FromBlock: blockInt,
		Addresses: addresses,
		Topics: [][]common.Hash{
			topics,
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
