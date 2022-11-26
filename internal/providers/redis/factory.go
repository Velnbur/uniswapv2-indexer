package redis

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-redis/redis/v8"
)

type UniswapV2FactoryProvider struct {
	address common.Address

	redis *redis.Client
}

func NewUniswapV2FactoryProvider(
	address common.Address, redis *redis.Client,
) *UniswapV2FactoryProvider {
	return &UniswapV2FactoryProvider{
		address: address,
		redis:   redis,
	}
}

const uniswapV2FactoryPairKey = "uniswapV2:factory:%s:pair:%d"

func (p *UniswapV2FactoryProvider) GetPair(
	index uint64,
) (common.Address, error) {
	key := fmt.Sprintf(uniswapV2FactoryPairKey, p.address.Hex(), index)

	var value string
	err := p.redis.Get(p.redis.Context(), key).Scan(&value)

	switch err {
	case nil:
		return common.HexToAddress(value), nil
	case redis.Nil:
		return common.Address{}, nil
	default:
		return common.Address{}, err
	}
}

func (p *UniswapV2FactoryProvider) SetPair(
	index uint64, value common.Address,
) error {
	key := fmt.Sprintf(uniswapV2FactoryPairKey, p.address.Hex(), index)

	return p.redis.Set(p.redis.Context(), key, value.Hex(), 0).Err()
}
