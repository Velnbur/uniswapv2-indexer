package listener

import (
	"context"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	etypes "github.com/ethereum/go-ethereum/core/types"
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

	uniswapV2 *contracts.UniswapV2

	swapEvents chan *uniswapv2pair.UniswapV2PairSwap
}

func NewListener(cfg config.Config) (*Listener, error) {
	return &Listener{
		client:     cfg.EthereumClient(),
		logger:     cfg.Log().WithField("service", "listener"),
		uniswapV2:  cfg.UniswapV2(),
		swapEvents: make(chan *uniswapv2pair.UniswapV2PairSwap),
	}, nil
}

func (l *Listener) Run(ctx context.Context) error {
	tokenSwapSig := []byte("Swap(address,uint256,uint256,uint256,uint256,address)")
	tokenSwapSigHash := crypto.Keccak256Hash(tokenSwapSig)

	contractAbi, err := abi.JSON(strings.NewReader(
		string(uniswapv2pair.UniswapV2PairABI),
	))
	if err != nil {
		return errors.Wrap(err, "failed to parse pair ABI")
	}

	var addresses []common.Address
	l.uniswapV2.Pairs.Range(func(addr common.Address, _ *contracts.UniswapV2Pair) bool {
		addresses = append(addresses, addr)
		return true
	})

	query := ethereum.FilterQuery{
		Addresses: addresses,
		Topics: [][]common.Hash{
			{tokenSwapSigHash},
		},
	}

	logs := make(chan etypes.Log)
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

func (l *Listener) getAllPairsAddresses() []common.Address {
	var addresses []common.Address
	l.uniswapV2.Pairs.Range(func(addr common.Address, _ *contracts.UniswapV2Pair) bool {
		addresses = append(addresses, addr)
		return true
	})

	return addresses
}
func (l *Listener) Listen(ctx context.Context) error {
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
