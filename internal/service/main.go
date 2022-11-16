package service

import (
	"context"
	"sync"

	"github.com/Velnbur/uniswapv2-indexer/internal/config"
	"github.com/Velnbur/uniswapv2-indexer/internal/service/api"
	"github.com/Velnbur/uniswapv2-indexer/internal/service/listener"
)

type Runner func(ctx context.Context, cfg config.Config)

var services = map[string]Runner{
	"api":      api.Run,
	"listener": listener.Run,
}

func Run(ctx context.Context, cfg config.Config) {
	logger := cfg.Log()
	wg := new(sync.WaitGroup)

	for name, service := range services {
		logger.WithField("service", name).Info("starting service")
		wg.Add(1)

		go func(ctx context.Context, cfg config.Config, runner Runner) {
			defer wg.Done()
			runner(ctx, cfg)
		}(ctx, cfg, service)
	}

	logger.Info("all services started")
	wg.Wait()
}
