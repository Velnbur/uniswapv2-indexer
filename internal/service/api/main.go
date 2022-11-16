package api

import (
	"context"
	"net"

	"github.com/Velnbur/uniswapv2-indexer/internal/config"
	"github.com/go-chi/chi"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/kit/copus/types"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

type service struct {
	log      *logan.Entry
	copus    types.Copus
	listener net.Listener
	router   chi.Router
}

func (s *service) run(ctx context.Context, cfg config.Config) error {
	if err := s.copus.RegisterChi(s.router); err != nil {
		return errors.Wrap(err, "cop failed")
	}

	ape.Serve(ctx, s.router, cfg, ape.ServeOpts{})
	return nil
}

func newService(cfg config.Config) *service {
	return &service{
		log:      cfg.Log(),
		copus:    cfg.Copus(),
		listener: cfg.Listener(),
		router:   newRouter(cfg),
	}
}

func Run(ctx context.Context, cfg config.Config) {
	if err := newService(cfg).run(ctx, cfg); err != nil {
		panic(err)
	}
}
