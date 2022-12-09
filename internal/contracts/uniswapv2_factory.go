package contracts

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
	"gitlab.com/distributed_lab/logan/v3"

	uniswapv2factory "github.com/Velnbur/uniswapv2-indexer/generated/uniswapv2-factory"
	"github.com/Velnbur/uniswapv2-indexer/internal/providers"
	"github.com/Velnbur/uniswapv2-indexer/pkg/helpers"
)

type UniswapV2FactoryConfig struct {
	Address common.Address
	Client  *ethclient.Client
	Logger  *logan.Entry

	Provider      providers.UniswapV2FactoryProvider
	PairProvider  providers.UniswapV2PairProvider
	Erc20Provider providers.Erc20Provider
}

type UniswapV2Factory struct {
	address  common.Address
	contract *uniswapv2factory.UniswapV2Factory

	client *ethclient.Client
	logger *logan.Entry

	provider      providers.UniswapV2FactoryProvider
	pairProvider  providers.UniswapV2PairProvider
	erc20Provider providers.Erc20Provider
}

// NewUniswapV2Factory creates a new UniswapV2Factory instance
func NewUniswapV2Factory(cfg UniswapV2FactoryConfig) (*UniswapV2Factory, error) {
	contract, err := uniswapv2factory.NewUniswapV2Factory(
		cfg.Address, cfg.Client,
	)
	if err != nil {
		return nil, err
	}
	return &UniswapV2Factory{
		address:       cfg.Address,
		client:        cfg.Client,
		contract:      contract,
		provider:      cfg.Provider,
		pairProvider:  cfg.PairProvider,
		logger:        cfg.Logger,
		erc20Provider: cfg.Erc20Provider,
	}, nil
}

// AllPairLength returns the number of all pairs
func (u *UniswapV2Factory) AllPairLength(ctx context.Context) (uint64, error) {
	// TODO: may be cache this too
	length, err := u.contract.AllPairsLength(&bind.CallOpts{
		Context: ctx,
	})
	if err != nil {
		return 0, errors.New("failed to get all pairs length")
	}
	return length.Uint64(), nil
}

// AllPairs return pair by index
func (u *UniswapV2Factory) AllPairs(
	ctx context.Context, index uint64,
) (*UniswapV2Pair, error) {
	// first check cache
	if u.provider != nil {
		pair, err := u.provider.GetPairByIndex(ctx, u.address, index)
		if err != nil {
			u.logger.WithError(err).Error("failed to get pair from cache")
		}
		if !helpers.IsAddressZero(pair) {
			return NewUniswapV2Pair(
				UniswapV2PairConfig{
					Address:       pair,
					Client:        u.client,
					Logger:        u.logger,
					Provider:      u.pairProvider,
					Erc20Provider: u.erc20Provider,
				},
			)
		}
	}

	// then check ethereum
	pairAddress, err := u.contract.AllPairs(&bind.CallOpts{
		Context: ctx,
	}, new(big.Int).SetUint64(index))
	if err != nil {
		return nil, errors.Wrap(err, "failed to get pair address")
	}

	// save to cache
	if u.provider != nil {
		err = u.provider.SetPairByIndex(ctx, u.address, pairAddress, index)
		if err != nil {
			u.logger.WithError(err).Error("failed to set pair to cache")
		}
	}

	return NewUniswapV2Pair(
		UniswapV2PairConfig{
			Address:       pairAddress,
			Client:        u.client,
			Logger:        u.logger,
			Provider:      u.pairProvider,
			Erc20Provider: u.erc20Provider,
		},
	)
}
