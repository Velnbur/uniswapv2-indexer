package providers

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-redis/redis/v8"
)

var _ UniswapV2FactoryProvider = &UniswapV2FactoryRedisProvider{}

type UniswapV2FactoryRedisProvider struct {
	redis *redis.Client
}

const uniswapV2FactoryPairKey = "uniswapV2:factory:%s:pair:%d"

func (p *UniswapV2FactoryRedisProvider) GetPairByIndex(
	ctx context.Context, factory common.Address, index uint64,
) (common.Address, error) {
	key := fmt.Sprintf(uniswapV2FactoryPairKey, factory.Hex(), index)

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

func (p *UniswapV2FactoryRedisProvider) SetPairByIndex(
	ctx context.Context, factory, pair common.Address, index uint64,
) error {
	key := fmt.Sprintf(uniswapV2FactoryPairKey, factory.Hex(), index)

	return p.redis.Set(p.redis.Context(), key, pair.Hex(), 0).Err()
}

func NewUniswapV2FactoryRedisProvider(
	redis *redis.Client,
) *UniswapV2FactoryRedisProvider {
	return &UniswapV2FactoryRedisProvider{
		redis: redis,
	}
}
