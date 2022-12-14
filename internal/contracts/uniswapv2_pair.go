package contracts

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
	"gitlab.com/distributed_lab/logan/v3"

	uniswapv2pair "github.com/Velnbur/uniswapv2-indexer/generated/uniswapv2-pair"
	"github.com/Velnbur/uniswapv2-indexer/internal/providers"
	"github.com/Velnbur/uniswapv2-indexer/pkg/helpers"
)

type UniswapV2PairConfig struct {
	Address common.Address
	Client  *ethclient.Client
	Logger  *logan.Entry

	Provider      providers.UniswapV2PairProvider
	Erc20Provider providers.Erc20Provider
}

type UniswapV2Pair struct {
	token0 common.Address
	token1 common.Address

	Address  common.Address
	contract *uniswapv2pair.UniswapV2Pair

	provider      providers.UniswapV2PairProvider
	erc20Provider providers.Erc20Provider
	logger        *logan.Entry

	client *ethclient.Client
}

func NewUniswapV2Pair(cfg UniswapV2PairConfig) (*UniswapV2Pair, error) {
	contract, err := uniswapv2pair.NewUniswapV2Pair(cfg.Address, cfg.Client)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create uniswapv2pair contract")
	}

	return &UniswapV2Pair{
		Address:       cfg.Address,
		contract:      contract,
		provider:      cfg.Provider,
		logger:        cfg.Logger,
		client:        cfg.Client,
		erc20Provider: cfg.Erc20Provider,
	}, nil
}

func (u *UniswapV2Pair) GetReserves(ctx context.Context) (*big.Int, *big.Int, error) {
	reservesRes, err := u.contract.GetReserves(&bind.CallOpts{Context: ctx})
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to get reserves")
	}

	return reservesRes.Reserve0, reservesRes.Reserve1, nil
}

func (u *UniswapV2Pair) Token0(ctx context.Context) (*ERC20, error) {
	if !helpers.IsAddressZero(u.token0) {
		return NewERC20(
			Erc20Config{
				Address:  u.token0,
				Client:   u.client,
				Provider: u.erc20Provider,
			},
		)
	}

	err := u.initTokens(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init tokens")
	}

	return NewERC20(Erc20Config{
		Address:  u.token0,
		Client:   u.client,
		Provider: u.erc20Provider,
	})
}

func (u *UniswapV2Pair) Token1(ctx context.Context) (*ERC20, error) {
	if !helpers.IsAddressZero(u.token1) {
		return NewERC20(Erc20Config{
			Address:  u.token1,
			Client:   u.client,
			Provider: u.erc20Provider,
		})
	}

	err := u.initTokens(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init tokens")
	}

	return NewERC20(Erc20Config{
		Address:  u.token1,
		Client:   u.client,
		Provider: u.erc20Provider,
	})
}

func (u *UniswapV2Pair) initTokens(ctx context.Context) error {
	// get from cache first

	if u.provider != nil {
		token0, token1, err := u.provider.GetTokens(ctx, u.Address)
		if err != nil {
			u.logger.WithError(err).Error("failed to get tokens from provider")
		}
		if !helpers.IsAddressZero(token0) && !helpers.IsAddressZero(token1) {
			u.token0 = token0
			u.token1 = token1
			return nil
		}
	}

	token0, token1, err := u.getTokensFromContract(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get tokens")
	}

	u.token0 = token0
	u.token1 = token1

	if u.provider != nil {
		if err = u.provider.SetTokens(ctx, u.Address, token0, token1); err != nil {
			u.logger.WithError(err).Error("failed to set tokens to provider")
		}
	}

	return nil
}

func (u *UniswapV2Pair) getTokensFromContract(
	ctx context.Context,
) (token0, token1 common.Address, err error) {

	token0, err = u.contract.Token0(&bind.CallOpts{Context: ctx})
	if err != nil {
		return common.Address{}, common.Address{}, errors.Wrap(err, "failed to get token0")
	}

	token1, err = u.contract.Token1(&bind.CallOpts{Context: ctx})
	if err != nil {
		return common.Address{}, common.Address{}, errors.Wrap(err, "failed to get token1")
	}

	return token0, token1, nil
}
