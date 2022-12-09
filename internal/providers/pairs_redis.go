package providers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
)

var _ UniswapV2PairProvider = &UniswapV2PairsRedisProvider{}

type UniswapV2PairsRedisProvider struct {
	redis *redis.Client
}

func NewUniswapV2PairsRedisProvider(redis *redis.Client) *UniswapV2PairsRedisProvider {
	return &UniswapV2PairsRedisProvider{
		redis: redis,
	}
}

const uniswapV2TokensKey = "uniswav2-pair:%s:tokens"

type tokens struct {
	Token0 common.Address `json:"token0"`
	Token1 common.Address `json:"token1"`
}

func (p *UniswapV2PairsRedisProvider) GetTokens(
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

func (p *UniswapV2PairsRedisProvider) SetTokens(
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
