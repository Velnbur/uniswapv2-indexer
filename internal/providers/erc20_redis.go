package providers

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
)

var _ Erc20Provider = &Erc20RedisProvider{}

type Erc20RedisProvider struct {
	cache *redis.Client
}

func NewErc20RedisProvider(cache *redis.Client) *Erc20RedisProvider {
	return &Erc20RedisProvider{cache: cache}
}

const (
	erc20SymbolKey = "erc20:%s:symbol"
)

func (p *Erc20RedisProvider) GetSymbol(
	ctx context.Context, address common.Address,
) (string, error) {
	key := fmt.Sprintf(erc20SymbolKey, address.Hex())

	name, err := p.cache.Get(ctx, key).Result()

	if err == nil {
		return name, nil
	}
	if errors.Is(err, redis.Nil) {
		return "", nil
	}

	return "", errors.Wrap(err, "failed to get erc20 symbol")
}

func (p *Erc20RedisProvider) SetSymbol(
	ctx context.Context, address common.Address, symbol string,
) error {
	key := fmt.Sprintf(erc20SymbolKey, address.Hex())

	err := p.cache.Set(ctx, key, symbol, 0).Err()
	if err != nil {
		return errors.Wrap(err, "failed to set erc20 symbol")
	}

	return nil
}
