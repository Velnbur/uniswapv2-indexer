package contracts

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"gitlab.com/distributed_lab/logan/v3"

	uniswapv2pair "github.com/Velnbur/uniswapv2-indexer/contracts/uniswapv2-pair"
)

type UniswapV2Pair struct {
	reserve0 *big.Int
	reserve1 *big.Int

	token0 common.Address
	token1 common.Address

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

func (u *UniswapV2Pair) GetReserves(ctx context.Context) (*big.Int, *big.Int, error) {
	if u.reserve0 != nil && u.reserve1 != nil {
		return u.reserve0, u.reserve1, nil
	}

	// get from cache first
	reserve0, reserve1 := u.getReservesFromCache(ctx)
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

	u.storeReservesInCache(ctx, reserves{reserve0, reserve1})

	return u.reserve0, u.reserve1, nil
}

type reserves struct {
	Reserve0 *big.Int `json:"reserve0"`
	Reserve1 *big.Int `json:"reserve1"`
}

const uniswapV2ReservesKey = "uniswapv2-pair:%s:reserves"

func (u *UniswapV2Pair) getReservesFromCache(ctx context.Context) (*big.Int, *big.Int) {
	key := fmt.Sprintf(uniswapV2ReservesKey, u.Address.Hex())

	reservesStr, err := u.redis.Get(ctx, key).Result()
	switch err {
	case nil:
		var reserves reserves
		err = json.Unmarshal([]byte(reservesStr), &reserves)
		if err != nil {
			u.logger.WithError(err).Error("failed to unmarshal reserves")
			return nil, nil
		}
		return reserves.Reserve0, reserves.Reserve1
	case redis.Nil:
		// do nothing
	default:
		u.logger.WithError(err).Error("failed to get reserves from redis")
	}

	return nil, nil
}

func (u *UniswapV2Pair) storeReservesInCache(ctx context.Context, reserves reserves) {
	key := fmt.Sprintf(uniswapV2ReservesKey, u.Address.Hex())

	reservesStr, err := json.Marshal(reserves)
	if err != nil {
		u.logger.WithError(err).Error("failed to marshal reserves")
		return
	}

	err = u.redis.Set(ctx, key, reservesStr, 0).Err()
	if err != nil {
		u.logger.WithError(err).Error("failed to store reserves in redis")
	}
}

func (u *UniswapV2Pair) Token0(ctx context.Context) (common.Address, error) {
	if !isAddressZero(u.token0) {
		return u.token0, nil
	}

	err := u.initTokens(ctx)

	return u.token0, err
}

func (u *UniswapV2Pair) Token1(ctx context.Context) (common.Address, error) {
	if !isAddressZero(u.token1) {
		return u.token1, nil
	}

	err := u.initTokens(ctx)

	return u.token1, err
}

func (u *UniswapV2Pair) initTokens(ctx context.Context) error {

	// get from cache first
	token0, token1 := u.getTokensFromCache(ctx)
	if !isAddressZero(token0) {
		u.token0 = token0
		u.token1 = token1
		return nil
	}

	token0, token1, err := u.getTokensFromContract(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get tokens")
	}

	u.storeTokensInCache(ctx, token0, token1)

	return nil
}

const uniswapV2TokensKey = "uniswav2-pair:%s:tokens"

type tokens struct {
	Token0 common.Address `json:"token0"`
	Token1 common.Address `json:"token1"`
}

func (u *UniswapV2Pair) getTokensFromCache(ctx context.Context) (common.Address, common.Address) {
	key := fmt.Sprintf(uniswapV2TokensKey, u.Address.Hex())
	// firstly, try to get token0 from redis
	tokenStr, err := u.redis.Get(ctx, key).Result()
	switch err {
	case nil:
		var tokens tokens
		err = json.Unmarshal([]byte(tokenStr), &tokens)
		if err != nil {
			u.logger.WithError(err).Error("failed to unmarshal tokens")
			return common.Address{}, common.Address{}
		}
		return tokens.Token0, tokens.Token1
	case redis.Nil:
		// do nothing
	default:
		u.logger.WithError(err).Error("failed to get token from redis")
	}

	return common.Address{}, common.Address{}
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

func (u *UniswapV2Pair) storeTokensInCache(
	ctx context.Context, token0, token1 common.Address,
) {
	key := fmt.Sprintf(uniswapV2TokensKey, u.Address.Hex())

	tokens := tokens{token0, token1}
	tokensStr, err := json.Marshal(tokens)
	if err != nil {
		u.logger.WithError(err).Error("failed to marshal tokens")
		return
	}

	err = u.redis.Set(ctx, key, tokensStr, 0).Err()
	if err != nil {
		u.logger.WithError(err).Error("failed to store tokens in redis")
	}
}

func isAddressZero(address common.Address) bool {
	return address == common.Address{}
}

func SwapTokenTopic() common.Hash {
	tokenSwapSig := []byte("Swap(address,uint256,uint256,uint256,uint256,address)")
	tokenSwapSigHash := crypto.Keccak256Hash(tokenSwapSig)
	return tokenSwapSigHash
}
