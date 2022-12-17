package indexer

import (
	"context"
	"time"

	"github.com/Velnbur/uniswapv2-indexer/internal/channels"
	"github.com/Velnbur/uniswapv2-indexer/internal/config"
	"github.com/Velnbur/uniswapv2-indexer/internal/providers"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

type Indexer struct {
	graph *Graph

	logger      *logan.Entry
	eventsQueue channels.EventQueue
	pathes      providers.PathesProvider
}

func New(cfg config.Config) *Indexer {
	return &Indexer{
		graph:       NewGraph(),
		eventsQueue: cfg.EventsQueue(),
		logger:      cfg.Log(),
		pathes:      providers.NewPathesRedisProvider(cfg.Redis()),
	}
}

const defaultDumpTimeout = time.Second * 5

func (ind *Indexer) Run(ctx context.Context) error {
	eventsSubscription, err := ind.eventsQueue.Receive(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to receive events from queue")
	}

	for {
		select {
		case <-ctx.Done():
			ind.dumpGraphWithTimeout()
			return nil
		case event := <-eventsSubscription:
			ind.processEvent(ctx, &event)
		}
	}
}

func (ind *Indexer) processEvent(ctx context.Context, event *channels.Event) {
	switch event.Type {
	case channels.BlockCreationEvent:
		ind.graph.Index()
	case channels.ReservesUpdateEvent:
		ind.graph.UpdateReserves(
			event.ReservesUpdate.Token0,
			event.ReservesUpdate.Token1,
			event.ReservesUpdate.Reserve0Delta,
			event.ReservesUpdate.Reserve1Delta,
		)
	case channels.PairCreationEvent:
		ind.graph.AddEdge(
			event.PairCreation.Token0,
			event.PairCreation.Token1,
			event.PairCreation.Reserve0,
			event.PairCreation.Reserve1,
		)
	}
}

func (ind *Indexer) dumpGraphWithTimeout() {
	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), defaultDumpTimeout)

	go func() {
		if err := ind.dumpGraph(ctxWithTimeout); err != nil {
			ind.logger.WithError(err).Error("failed to dump graph")
		}
	}()

	select {
	case <-ctxWithTimeout.Done():
		cancel()
	}
}

func (ind *Indexer) dumpGraph(ctx context.Context) error {
	for edge, pathes := range ind.graph.pathesMap.m {
		err := ind.pathes.SetPathes(ctx, edge.Token0, edge.Token1, pathes)
		if err != nil {
			return errors.Wrap(err, "failed to dump pathes", logan.F{
				"token0": edge.Token0.Hex(),
				"token1": edge.Token1.Hex(),
			})
		}
	}

	return nil
}
