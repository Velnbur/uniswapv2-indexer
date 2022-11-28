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

type UniswapV2Pair struct {
	reserve0 *big.Int
	reserve1 *big.Int

	token0 common.Address
	token1 common.Address

	Address  common.Address
	contract *uniswapv2pair.UniswapV2Pair

	provider providers.UniswapV2PairProvider
	logger   *logan.Entry
}

func NewUniswapV2Pair(
	address common.Address, client *ethclient.Client, logger *logan.Entry,
	provider providers.UniswapV2PairProvider,
) (*UniswapV2Pair, error) {
	contract, err := uniswapv2pair.NewUniswapV2Pair(address, client)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create uniswapv2pair contract")
	}

	return &UniswapV2Pair{
		Address:  address,
		contract: contract,
		provider: provider,
		logger:   logger,
	}, nil
}

func (u *UniswapV2Pair) GetReserves(ctx context.Context) (*big.Int, *big.Int, error) {
	if u.reserve0 != nil && u.reserve1 != nil {
		return u.reserve0, u.reserve1, nil
	}

	// get from cache first
	reserve0, reserve1, err := u.provider.GetReserves(ctx, u.Address)
	if err != nil {
		u.logger.WithError(err).Error("failed to get reserves from provider")
	}
	if reserve0 != nil && reserve1 != nil {
		u.reserve0 = reserve0
		u.reserve1 = reserve1
		return reserve0, reserve1, nil
	}

	reservesRes, err := u.contract.GetReserves(&bind.CallOpts{Context: ctx})
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to get reserves")
	}

	u.reserve0 = reservesRes.Reserve0
	u.reserve1 = reservesRes.Reserve1

	if err = u.provider.SetReserves(ctx, u.Address, reserve0, reserve1); err != nil {
		u.logger.WithError(err).Error("failed to set reserves to provider")
	}

	return u.reserve0, u.reserve1, nil
}

func (u *UniswapV2Pair) Token0(ctx context.Context) (common.Address, error) {
	if !helpers.IsAddressZero(u.token0) {
		return u.token0, nil
	}

	err := u.initTokens(ctx)

	return u.token0, err
}

func (u *UniswapV2Pair) Token1(ctx context.Context) (common.Address, error) {
	if !helpers.IsAddressZero(u.token1) {
		return u.token1, nil
	}

	err := u.initTokens(ctx)

	return u.token1, err
}

func (u *UniswapV2Pair) initTokens(ctx context.Context) error {
	// get from cache first

	token0, token1, err := u.provider.GetTokens(ctx, u.Address)
	if err != nil {
		u.logger.WithError(err).Error("failed to get tokens from provider")
	}
	if !helpers.IsAddressZero(token0) && !helpers.IsAddressZero(token1) {
		u.token0 = token0
		u.token1 = token1
		return nil
	}

	token0, token1, err = u.getTokensFromContract(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get tokens")
	}

	u.token0 = token0
	u.token1 = token1

	if err = u.provider.SetTokens(ctx, u.Address, token0, token1); err != nil {
		u.logger.WithError(err).Error("failed to set tokens to provider")
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
