package contracts

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"gitlab.com/distributed_lab/logan/v3"

	uniswapv2pair "github.com/Velnbur/uniswapv2-indexer/contracts/uniswapv2-pair"
)

type UniswapV2Pair struct {
	token0 *common.Address
	token1 *common.Address

	Address  common.Address
	contract *uniswapv2pair.UniswapV2Pair

	redis  *redis.Client
	logger *logan.Entry
}

func NewUniswapV2Pair(
	address common.Address, client *ethclient.Client, redis *redis.Client,
	logger *logan.Entry,
) (*UniswapV2Pair, error) {
	contract, err := uniswapv2pair.NewUniswapV2Pair(address, client)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create uniswapv2pair contract")
	}

	return &UniswapV2Pair{
		Address:  address,
		contract: contract,
		redis:    redis,
		logger:   logger,
	}, nil
}

func (u *UniswapV2Pair) Token0(ctx context.Context) (common.Address, error) {
	if u.token0 != nil {
		return *u.token0, nil
	}

	tokenAddr, err := u.getAndStoreToken(ctx, ":token0", token0)
	if err != nil {
		return common.Address{}, errors.Wrap(err, "failed to get token0")
	}
	u.token0 = &tokenAddr

	return *u.token0, nil
}

func (u *UniswapV2Pair) Token1(ctx context.Context) (common.Address, error) {
	if u.token1 != nil {
		return *u.token1, nil
	}

	tokenAddr, err := u.getAndStoreToken(ctx, ":token1", token1)
	if err != nil {
		return common.Address{}, errors.Wrap(err, "failed to get token1")
	}
	u.token1 = &tokenAddr

	return *u.token1, nil
}

type tokenNumber int

const (
	token0 tokenNumber = iota
	token1
)

func (u *UniswapV2Pair) getAndStoreToken(
	ctx context.Context, key string, tokenNum tokenNumber,
) (common.Address, error) {
	key = fmt.Sprintf("uniswapv2-pair:%s:%s", u.Address.Hex(), key)
	// firstly, try to get token0 from redis
	tokenStr, err := u.redis.Get(ctx, key).Result()
	switch err {
	case nil:
		return common.HexToAddress(tokenStr), nil
	case redis.Nil:
		// do nothing
	default:
		u.logger.WithError(err).Error("failed to get token from redis")
	}

	var tokenAddr common.Address
	switch tokenNum {
	case token0:
		tokenAddr, err = u.contract.Token0(&bind.CallOpts{Context: ctx})
	case token1:
		tokenAddr, err = u.contract.Token1(&bind.CallOpts{Context: ctx})
	default:
		panic("unknown token number")
	}
	if err != nil {
		return common.Address{}, errors.Wrap(err, "failed to get token from contract")
	}

	// save token0 to redis
	err = u.redis.Set(ctx, key, tokenAddr.Hex(), 0).Err()
	if err != nil {
		u.logger.WithError(err).Error("failed to save token to redis")
	}

	return tokenAddr, nil
}
