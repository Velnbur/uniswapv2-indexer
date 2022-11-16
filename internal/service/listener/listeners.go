package listener

import (
	"context"

	uniswapv2factory "github.com/Velnbur/uniswapv2-indexer/contracts/uniswapv2-factory"
	uniswapv2pair "github.com/Velnbur/uniswapv2-indexer/contracts/uniswapv2-pair"
	"github.com/pkg/errors"
)

func (l *Listener) listenPair(ctx context.Context, pair *uniswapv2pair.UniswapV2Pair) error {
	burn := make(chan *uniswapv2pair.UniswapV2PairBurn)
	swap := make(chan *uniswapv2pair.UniswapV2PairSwap)
	mint := make(chan *uniswapv2pair.UniswapV2PairMint)

	subBurn, err := pair.WatchBurn(nil, burn, nil, nil)
	if err != nil {
		return errors.Wrap(err, "failed to subscribe to burn event")
	}
	subSwap, err := pair.WatchSwap(nil, swap, nil, nil)
	if err != nil {
		return errors.Wrap(err, "failed to subscribe to swap event")
	}
	subMint, err := pair.WatchMint(nil, mint, nil)
	if err != nil {
		return errors.Wrap(err, "failed to subscribe to mint event")
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case event := <-burn:
			l.events <- Event{
				Type: EventTypeBurn,
				Burn: event,
			}
		case event := <-swap:
			l.events <- Event{
				Type: EventTypeSwap,
				Swap: event,
			}
		case event := <-mint:
			l.events <- Event{
				Type: EventTypeMint,
				Mint: event,
			}
		case err := <-subBurn.Err():
			return errors.Wrap(err, "burn subscription error")
		case err := <-subSwap.Err():
			return errors.Wrap(err, "swap subscription error")
		case err := <-subMint.Err():
			return errors.Wrap(err, "mint subscription error")
		}
	}
}

func (l *Listener) listenFactory(ctx context.Context) error {
	createPair := make(chan *uniswapv2factory.UniswapV2FactoryPairCreated)
	subCreatePair, err := l.factory.WatchPairCreated(nil, createPair, nil, nil)
	if err != nil {
		return errors.Wrap(err, "failed to subscribe to create pair event")
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case event := <-createPair:
			l.events <- Event{
				Type:       EventTypePairCreated,
				CreatePair: event,
			}
		case err := <-subCreatePair.Err():
			return errors.Wrap(err, "create pair subscription error")
		}
	}
}
