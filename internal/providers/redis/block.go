package redis

import (
	"context"

	"github.com/go-redis/redis/v8"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

type BlockProvider struct {
	redis *redis.Client

	block uint64
}

// NewBlockProvider returns a new BlockProvider.
func NewBlockProvider(redis *redis.Client) *BlockProvider {
	return &BlockProvider{
		redis: redis,
	}
}

const currentBlockKey = "current_block"

// CurrentBlock returns the current block.
func (p *BlockProvider) CurrentBlock(ctx context.Context) (uint64, error) {
	if p.block == 0 {
		block, err := p.redis.Get(ctx, currentBlockKey).Uint64()
		switch err {
		case nil:
			p.block = block
			return block, nil
		case redis.Nil:
			return 0, nil
		default:
			return 0, errors.Wrap(err, "failed to get current block")
		}
	}

	return p.block, nil
}

// UpdateBlock updates the current block.
func (p *BlockProvider) UpdateBlock(ctx context.Context, block uint64) error {
	if p.block != block {
		p.block = block
		return p.redis.Set(ctx, currentBlockKey, block, 0).Err()
	}
	return nil
}
