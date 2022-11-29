package redis

import (
	"context"
	"fmt"

	"github.com/Velnbur/uniswapv2-indexer/internal/providers"
	"github.com/ethereum/go-ethereum/common"
	"github.com/go-redis/redis/v8"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

var _ providers.Erc20Provider = &Erc20Provider{}

type Erc20Provider struct {
	cache *redis.Client
}

func NewErc20Provider(cache *redis.Client) *Erc20Provider {
	return &Erc20Provider{cache: cache}
}

const (
	erc20SymbolKey = "erc20:%s:symbol"
)

func (p *Erc20Provider) GetSymbol(ctx context.Context, address common.Address) (string, error) {
	key := fmt.Sprintf(erc20SymbolKey, address.Hex())

	name, err := p.cache.Get(ctx, key).Result()

	switch err {
	case nil:
		return name, nil
	case redis.Nil:
		return "", nil
	default:
		return "", errors.Wrap(err, "failed to get erc20 symbol")
	}
}

func (p *Erc20Provider) SetSymbol(ctx context.Context, address common.Address, symbol string) error {
	key := fmt.Sprintf(erc20SymbolKey, address.Hex())

	err := p.cache.Set(ctx, key, symbol, 0).Err()
	if err != nil {
		return errors.Wrap(err, "failed to set erc20 symbol")
	}

	return nil
}
