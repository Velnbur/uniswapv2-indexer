package api

import (
	"github.com/go-chi/chi"
	"gitlab.com/distributed_lab/ape"

	"github.com/Velnbur/uniswapv2-indexer/internal/config"
	"github.com/Velnbur/uniswapv2-indexer/internal/service/api/handlers"
)

func newRouter(cfg config.Config) chi.Router {
	r := chi.NewRouter()

	r.Use(
		ape.RecoverMiddleware(cfg.Log()),
		ape.LoganMiddleware(cfg.Log()),
		ape.CtxMiddleware(
			handlers.CtxLog(cfg.Log()),
		),
	)
	r.Route("/", func(r chi.Router) {
		// configure endpoints here
	})

	return r
}
