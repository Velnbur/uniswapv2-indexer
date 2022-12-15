package indexer

import (
	"context"

	"github.com/Velnbur/uniswapv2-indexer/internal/channels"
	"github.com/Velnbur/uniswapv2-indexer/internal/config"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

type Indexer struct {
	graph *Graph

	logger      *logan.Entry
	eventsQueue channels.EventQueue
}

func New(cfg config.Config) *Indexer {
	return &Indexer{
		graph:       NewGraph(),
		eventsQueue: cfg.EventsQueue(),
		logger:      cfg.Log(),
	}
}

func (ind *Indexer) Run(ctx context.Context) error {
	eventsSubscription, err := ind.eventsQueue.Receive(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to receive events from queue")
	}

	for {
		select {
		case <-ctx.Done():
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
