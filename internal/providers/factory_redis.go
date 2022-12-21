package providers

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
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

	if err != nil {
		if errors.Is(err, redis.Nil) {
			return common.Address{}, nil
		}
		return common.Address{}, errors.Wrap(err, "failed  to get pair by index")
	}

	return common.HexToAddress(value), nil
}

func (p *UniswapV2FactoryRedisProvider) SetPairByIndex(
	ctx context.Context, factory, pair common.Address, index uint64,
) error {
	key := fmt.Sprintf(uniswapV2FactoryPairKey, factory.Hex(), index)

	return p.redis.Set(ctx, key, pair.Hex(), 0).Err()
}

func NewUniswapV2FactoryRedisProvider(
	redis *redis.Client,
) *UniswapV2FactoryRedisProvider {
	return &UniswapV2FactoryRedisProvider{
		redis: redis,
	}
}

const uniswapV2FactoryPairByTokensKey = "uniswapV2:factory:%s:pair:%s-%s"

func (p *UniswapV2FactoryRedisProvider) GetPairByTokens(
	ctx context.Context, factory, token0, token1 common.Address,
) (common.Address, error) {
	key := fmt.Sprintf(uniswapV2FactoryPairByTokensKey, factory.Hex(), token0.Hex(), token1.Hex())

	var value string
	err := p.redis.Get(p.redis.Context(), key).Scan(&value)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return common.Address{}, nil
		}
		return common.Address{}, errors.Wrap(err, "failed to get pair by tokens")
	}

	return common.HexToAddress(value), nil
}

func (p *UniswapV2FactoryRedisProvider) SetPairByTokens(
	ctx context.Context, factory, token0, token1, pair common.Address,
) error {
	key := fmt.Sprintf(uniswapV2FactoryPairByTokensKey, factory.Hex(), token0.Hex(), token1.Hex())

	return p.redis.Set(ctx, key, pair.Hex(), 0).Err()
}
