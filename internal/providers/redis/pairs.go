package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/Velnbur/uniswapv2-indexer/internal/providers"
	"github.com/ethereum/go-ethereum/common"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
)

var _ providers.UniswapV2FactoryProvider = &UniswapV2FactoryProvider{}

type UniswapV2PairsProvider struct {
	redis *redis.Client
}

func NewUniswapV2PairsProvider(redis *redis.Client) *UniswapV2PairsProvider {
	return &UniswapV2PairsProvider{
		redis: redis,
	}
}

const uniswapV2TokensKey = "uniswav2-pair:%s:tokens"

type tokens struct {
	Token0 common.Address `json:"token0"`
	Token1 common.Address `json:"token1"`
}

func (p *UniswapV2PairsProvider) GetTokens(
	ctx context.Context, pair common.Address,
) (common.Address, common.Address, error) {
	key := fmt.Sprintf(uniswapV2TokensKey, pair.Hex())

	tokenStr, err := p.redis.Get(ctx, key).Result()

	switch err {
	case nil:
		var tokens tokens
		err = json.Unmarshal([]byte(tokenStr), &tokens)
		if err != nil {
			return common.Address{}, common.Address{}, errors.Wrap(err,
				"failed to unmarshal tokens",
			)
		}
		return tokens.Token0, tokens.Token1, nil
	case redis.Nil:
		return common.Address{}, common.Address{}, nil
	default:
		return common.Address{}, common.Address{}, errors.Wrap(err,
			"failed to get tokens from redis",
		)
	}
}

func (p *UniswapV2PairsProvider) SetTokens(
	ctx context.Context, pair, token0, token1 common.Address,
) error {
	key := fmt.Sprintf(uniswapV2TokensKey, pair.Hex())

	tokens := tokens{
		Token0: token0,
		Token1: token1,
	}

	tokensStr, err := json.Marshal(tokens)
	if err != nil {
		return errors.Wrap(err, "failed to marshal tokens")
	}

	return p.redis.Set(ctx, key, tokensStr, 0).Err()
}

type reserves struct {
	Reserve0 *big.Int `json:"reserve0"`
	Reserve1 *big.Int `json:"reserve1"`
}

const uniswapV2ReservesKey = "uniswapv2-pair:%s:reserves"

func (p *UniswapV2PairsProvider) GetReserves(
	ctx context.Context, pair common.Address,
) (*big.Int, *big.Int, error) {
	key := fmt.Sprintf(uniswapV2ReservesKey, pair.Hex())

	reservesStr, err := p.redis.Get(ctx, key).Result()

	switch err {
	case nil:
		var reserves reserves
		err = json.Unmarshal([]byte(reservesStr), &reserves)
		if err != nil {
			return nil, nil, errors.Wrap(err, "failed to unmarshal reserves")
		}
		return reserves.Reserve0, reserves.Reserve1, nil
	case redis.Nil:
		return nil, nil, nil
	default:
		return nil, nil, errors.Wrap(err, "failed to get reserves from redis")
	}
}

func (p *UniswapV2PairsProvider) SetReserves(
	ctx context.Context, pair common.Address,
	reserve0, reserve1 *big.Int,
) error {
	key := fmt.Sprintf(uniswapV2ReservesKey, pair.Hex())

	reserves := reserves{
		Reserve0: reserve0,
		Reserve1: reserve1,
	}

	reservesStr, err := json.Marshal(reserves)
	if err != nil {
		return errors.Wrap(err, "failed to marshal reserves")
	}

	// TODO: set expiration
	return p.redis.Set(ctx, key, reservesStr, 0).Err()
}
