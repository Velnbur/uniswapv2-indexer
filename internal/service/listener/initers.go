package listener

import (
	"context"

	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"

	"github.com/Velnbur/uniswapv2-indexer/internal/channels"
)

// FIXME: this is a temporary solution, need to find a better way to handle rate
// limit errors that alchemy makes, when we make too many requests. May be,
// infura has better solutions for that
// const errRateLimitStr = "Your app has exceeded its compute units"

func (l *Listener) initContracts(ctx context.Context) error {
	for i, token0 := range l.tokens {
		for _, token1 := range l.tokens[i+1:] {
			pair, err := l.uniswapV2.Factory.GetPool(ctx, token0.Address(), token1.Address())
			if err != nil {
				return errors.Wrap(err, "failed to get pair address")
			}

			l.uniswapV2.Pairs.Set(pair.Address, pair)

			reserve0, reserve1, err := pair.GetReserves(ctx)
			if err != nil {
				return errors.Wrap(err, "failed to get reserves", logan.F{
					"address": pair.Address,
				})
			}

			err = l.eventQueue.Send(ctx, channels.Event{
				Type: channels.PairCreationEvent,
				PairCreation: &channels.PairCreation{
					Address:  pair.Address,
					Reserve0: reserve0,
					Reserve1: reserve1,
				},
			})
			if err != nil {
				return errors.Wrap(err, "failed to send pair creation event",
					logan.F{
						"reserve0": reserve0.String(),
						"reserve1": reserve1.String(),
						"address":  pair.Address,
					})
			}
		}
	}
	return nil
}
