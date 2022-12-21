package providers

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
)

type BlockRedisProvider struct {
	redis *redis.Client

	block uint64
}

// NewBlockProvider returns a new BlockProvider.
func NewBlockProvider(redis *redis.Client) *BlockRedisProvider {
	return &BlockRedisProvider{
		redis: redis,
	}
}

const currentBlockKey = "current_block"

// CurrentBlock returns the current block.
func (p *BlockRedisProvider) CurrentBlock(ctx context.Context) (uint64, error) {
	block, err := p.redis.Get(ctx, currentBlockKey).Uint64()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, nil
		}
		return 0, errors.Wrap(err, "failed to get current block")
	}

	return block, nil
}

// UpdateBlock updates the current block.
func (p *BlockRedisProvider) UpdateBlock(ctx context.Context, block uint64) error {
	if p.block != block {
		p.block = block
		return p.redis.Set(ctx, currentBlockKey, block, 0).Err()
	}
	return nil
}
