package indexer

import (
	"context"

	"github.com/Velnbur/uniswapv2-indexer/internal/config"
)

func Run(ctx context.Context, cfg config.Config) {
	if err := New(cfg).Run(ctx); err != nil {
		cfg.Log().WithError(err).Panic("indexer running failed")
	}
}
