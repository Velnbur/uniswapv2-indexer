package api

import (
	"context"

	"github.com/Velnbur/uniswapv2-indexer/internal/config"
)

func Run(ctx context.Context, cfg config.Config) {
	if err := New(cfg).run(ctx); err != nil {
		cfg.Log().WithError(err).Panic("failed to start api")
	}
}
