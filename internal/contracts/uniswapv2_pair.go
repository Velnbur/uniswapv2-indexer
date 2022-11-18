package contracts

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"gitlab.com/distributed_lab/logan/v3"

	uniswapv2pair "github.com/Velnbur/uniswapv2-indexer/contracts/uniswapv2-pair"
)

type UniswapV2Pair struct {
	reserve0 *big.Int
	reserve1 *big.Int

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

func (u *UniswapV2Pair) GetReserves(ctx context.Context) (*big.Int, *big.Int, error) {
	if u.reserve0 != nil && u.reserve1 != nil {
		return u.reserve0, u.reserve1, nil
	}

	// get from cache first
	reserve0, reserve1, err := u.getReservesFromCache(ctx)
	if err != nil {
		u.logger.WithError(err).Error("failed to get reserves from cache")
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

	// store in cache
	err = u.storeReservesInCache(ctx, reserves{reserve0, reserve1})
	if err != nil {
		u.logger.WithError(err).Error("failed to store reserves in cache")
	}

	return u.reserve0, u.reserve1, nil
}

type reserves struct {
	Reserve0 *big.Int `json:"reserve0"`
	Reserve1 *big.Int `json:"reserve1"`
}

const uniswapV2ReservesKey = "uniswapv2-pair:%s:reserves"

func (u *UniswapV2Pair) getReservesFromCache(ctx context.Context) (*big.Int, *big.Int, error) {
	key := fmt.Sprintf(uniswapV2ReservesKey, u.Address.Hex())

	reservesStr, err := u.redis.Get(ctx, key).Result()
	switch err {
	case nil:
		var reserves reserves
		err = json.Unmarshal([]byte(reservesStr), &reserves)
		if err != nil {
			return nil, nil, errors.Wrap(err, "failed to unmarshal reserves")
		}

		return reserves.Reserve0, reserves.Reserve1, nil
	case redis.Nil:
		// do nothing
	default:
		u.logger.WithError(err).Error("failed to get reserves from redis")
	}

	return nil, nil, nil
}

func (u *UniswapV2Pair) storeReservesInCache(ctx context.Context, reserves reserves) error {
	key := fmt.Sprintf(uniswapV2ReservesKey, u.Address.Hex())

	reservesStr, err := json.Marshal(reserves)
	if err != nil {
		return errors.Wrap(err, "failed to marshal reserves")
	}

	err = u.redis.Set(ctx, key, reservesStr, 0).Err()
	if err != nil {
		return errors.Wrap(err, "failed to set reserves in redis")
	}

	return nil
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
