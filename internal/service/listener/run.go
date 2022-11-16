package listener

import (
	"context"

	"github.com/Velnbur/uniswapv2-indexer/internal/config"
)

func Run(ctx context.Context, cfg config.Config) {
	listener, err := NewListener(cfg)
	if err != nil {
		cfg.Log().WithError(err).Panic("failed to create listener")
	}
	defer listener.Close()

	if err := listener.Run(ctx); err != nil {
		cfg.Log().WithError(err).Panic("failed to run listener")
	}
}
