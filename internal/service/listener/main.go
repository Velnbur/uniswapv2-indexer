package listener

import (
	"context"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"

	uniswapv2factory "github.com/Velnbur/uniswapv2-indexer/generated/uniswapv2-factory"
	uniswapv2pair "github.com/Velnbur/uniswapv2-indexer/generated/uniswapv2-pair"
	"github.com/Velnbur/uniswapv2-indexer/internal/channels"
	"github.com/Velnbur/uniswapv2-indexer/internal/config"
	"github.com/Velnbur/uniswapv2-indexer/internal/contracts"
	"github.com/Velnbur/uniswapv2-indexer/internal/providers"
)

type Listener struct {
	client *ethclient.Client
	logger *logan.Entry

	pairABI    abi.ABI
	factoryABI abi.ABI

	uniswapV2 *contracts.UniswapV2
	tokens    []*contracts.ERC20

	currentBlock providers.CurrentBlockProvider

	eventQueue    channels.EventQueue
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

	factoryABI, err := abi.JSON(strings.NewReader(
		string(uniswapv2factory.UniswapV2FactoryABI),
	))
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse factory ABI")
	}

	listener := &Listener{
		client:        cfg.EthereumClient(),
		logger:        cfg.Log().WithField("service", "listener"),
		pairABI:       pairABI,
		factoryABI:    factoryABI,
		tokens:        cfg.Tokens(),
		currentBlock:  providers.NewBlockProvider(cfg.Redis()),
		eventQueue:    channels.NewEventChan(),
		eventUnpacker: NewEventUnpacker(&pairABI, &factoryABI),
	}
	listener.initHandlers(pairABI, factoryABI)

	return listener, nil
}

func (l *Listener) Run(ctx context.Context) error {
	if err := l.initContracts(ctx); err != nil {
		return errors.Wrap(err, "failed to init contracts")
	}

	return l.Listen(ctx)
}
