package api

import (
	"context"

	"github.com/go-chi/chi"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/logan/v3"

	"github.com/Velnbur/uniswapv2-indexer/internal/config"
)

type API struct {
	log    *logan.Entry
	router chi.Router

	// FIXME: see "fix me" comment below
	run func(ctx context.Context) error
}

func New(cfg config.Config) *API {
	router := newRouter(cfg)

	api := &API{
		log:    cfg.Log(),
		router: router,
		// FIXME: I don't like the idea to save `config.Config` inside
		// service structure, but `ape.Serve` literally needs it to as
		// parameter
		run: func(ctx context.Context) error {
			ape.Serve(ctx, router, cfg, ape.ServeOpts{})
			return nil
		},
	}

	return api
}

func (api *API) Run(ctx context.Context) error {
	if api.run == nil {
		api.log.Panic("run function was not provided")
	}
	return api.run(ctx)
}
