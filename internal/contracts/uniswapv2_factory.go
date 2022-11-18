package contracts

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-redis/redis/v8"
	"gitlab.com/distributed_lab/logan/v3"

	uniswapv2factory "github.com/Velnbur/uniswapv2-indexer/contracts/uniswapv2-factory"
)

type UniswapV2Factory struct {
	address  common.Address
	contract *uniswapv2factory.UniswapV2Factory

	client *ethclient.Client
	redis  *redis.Client
	logger *logan.Entry
}

// NewUniswapV2Factory creates a new UniswapV2Factory instance
func NewUniswapV2Factory(
	address common.Address, client *ethclient.Client, redis *redis.Client,
) (*UniswapV2Factory, error) {
	contract, err := uniswapv2factory.NewUniswapV2Factory(address, client)
	if err != nil {
		return nil, err
	}
	return &UniswapV2Factory{
		address:  address,
		client:   client,
		contract: contract,
		redis:    redis,
	}, nil
}

// AllPairLength returns the number of all pairs
func (u *UniswapV2Factory) AllPairLength(ctx context.Context) (uint64, error) {
	length, err := u.contract.AllPairsLength(&bind.CallOpts{
		Context: ctx,
	})
	if err != nil {
		return 0, errors.New("failed to get all pairs length")
	}
	return length.Uint64(), nil
}

// AllPairs return pair by index
func (u *UniswapV2Factory) AllPairs(ctx context.Context, index uint64) (*UniswapV2Pair, error) {
	// first check redis
	key := fmt.Sprintf("uniswapv2-factory:%s:all-pairs:%d", u.address.Hex(), index)
	pairAddressStr, err := u.redis.Get(ctx, key).Result()
	switch err {
	case nil:
		return NewUniswapV2Pair(
			common.HexToAddress(pairAddressStr),
			u.client, u.redis, u.logger,
		)
	case redis.Nil:
		// do nothing
	default:
		u.logger.WithError(err).Error("failed to get pair address from redis")
	}

	// then check ethereum
	pairAddress, err := u.contract.AllPairs(&bind.CallOpts{
		Context: ctx,
	}, new(big.Int).SetUint64(index))
	if err != nil {
		return nil, errors.New("failed to get pair address from ethereum")
	}

	// save to redis
	err = u.redis.Set(ctx, key, pairAddress.Hex(), 0).Err()
	if err != nil {
		u.logger.WithError(err).Error("failed to save pair address to redis")
	}

	return NewUniswapV2Pair(pairAddress, u.client, u.redis, u.logger)
}
