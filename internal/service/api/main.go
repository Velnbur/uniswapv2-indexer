package api

import (
	"context"
	"net"

	"github.com/go-chi/chi"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/logan/v3"

	"github.com/Velnbur/uniswapv2-indexer/internal/config"
)

type service struct {
	log      *logan.Entry
	listener net.Listener
	router   chi.Router
	cfg      config.Config
}

func (s *service) run(ctx context.Context) error {
	ape.Serve(ctx, s.router, s.cfg, ape.ServeOpts{})
	return nil
}

func newService(cfg config.Config) *service {
	return &service{
		cfg:      cfg,
		log:      cfg.Log(),
		listener: cfg.Listener(),
		router:   newRouter(cfg),
	}
}

func Run(ctx context.Context, cfg config.Config) {
	if err := newService(cfg).run(ctx, cfg); err != nil {
		panic(err)
	}
}
