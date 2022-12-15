package listener

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"

	uniswapv2factory "github.com/Velnbur/uniswapv2-indexer/generated/uniswapv2-factory"
	uniswapv2pair "github.com/Velnbur/uniswapv2-indexer/generated/uniswapv2-pair"
	"github.com/Velnbur/uniswapv2-indexer/internal/channels"
)

type EventHandler func(ctx context.Context, log *types.Log) error

func (l *Listener) initHandlers(pair, factory abi.ABI) {
	l.eventHandlers = map[common.Hash]EventHandler{
		// TODO:
		pair.Events[SwapEvent.String()].ID: l.handleSwap,
		pair.Events[SyncEvent.String()].ID: l.handleSync,
		pair.Events[BurnEvent.String()].ID: l.handleBurn,
		pair.Events[MintEvent.String()].ID: l.handleMint,

		factory.Events[PairCreatedEvent.String()].ID: l.handlePairCreation,
	}
}

func (l *Listener) handleEvent(ctx context.Context, log *types.Log) error {
	if err := l.currentBlock.UpdateBlock(ctx, log.BlockNumber); err != nil {
		return errors.Wrap(err, "failed to update current block number")
	}

	for _, topic := range log.Topics {
		handler, ok := l.eventHandlers[topic]
		if !ok {
			l.logger.
				WithField("topic", topic).
				Warn("unknown topic")
			continue
		}

		if err := handler(ctx, log); err != nil {
			return errors.Wrap(err, "failed to handle log")
		}
	}
	return nil
}

func (l *Listener) handleSwap(ctx context.Context, log *types.Log) error {
	var event uniswapv2pair.UniswapV2PairSwap

	err := l.eventUnpacker.Unpack(&event, SwapEvent, log.Data)
	if err != nil {
		return errors.Wrap(err, "failed to unpack log")
	}

	l.logger.WithFields(logan.F{
		"pair":       event.Raw.Address,
		"sender":     event.Sender.Hex(),
		"amount0In":  event.Amount0In.String(),
		"amount1In":  event.Amount1In.String(),
		"amount0Out": event.Amount0Out.String(),
		"amount1Out": event.Amount1Out.String(),
		"to":         event.To.Hex(),
	}).Debug("pair swap")

	token0, token1, err := l.getTokens(ctx, event.Raw.Address)
	if err != nil {
		return errors.Wrap(err, "failed get tokens")
	}

	err = l.eventQueue.Send(ctx, channels.Event{
		Type: channels.ReservesUpdateEvent,
		ReservesUpdate: &channels.ReservesUpdate{
			Address: event.Raw.Address,
			Token0:  token0,
			Token1:  token1,
			// FIXME:
			Reserve0Delta: &big.Int{},
			Reserve1Delta: &big.Int{},
		},
	})
	return errors.Wrap(err, "failed to add event to queue")
}

func (l *Listener) handleSync(ctx context.Context, log *types.Log) error {
	var event uniswapv2pair.UniswapV2PairSync

	err := l.eventUnpacker.Unpack(&event, SyncEvent, log.Data)
	if err != nil {
		return errors.Wrap(err, "failed to unpack log")
	}

	l.logger.WithFields(logan.F{
		"pair":     event.Raw.Address,
		"reserve0": event.Reserve0.String(),
		"reserve1": event.Reserve1.String(),
	}).Debug("pair sync")

	token0, token1, err := l.getTokens(ctx, event.Raw.Address)
	if err != nil {
		return errors.Wrap(err, "failed get tokens")
	}

	err = l.eventQueue.Send(ctx, channels.Event{
		Type: channels.ReservesUpdateEvent,
		ReservesUpdate: &channels.ReservesUpdate{
			Address:       event.Raw.Address,
			Token0:        token0,
			Token1:        token1,
			Reserve0Delta: event.Reserve0,
			Reserve1Delta: event.Reserve1,
		},
	})

	return errors.Wrap(err, "failed to add event to queue")
}

func (l *Listener) handleMint(ctx context.Context, log *types.Log) error {
	var event uniswapv2pair.UniswapV2PairMint

	err := l.eventUnpacker.Unpack(&event, MintEvent, log.Data)
	if err != nil {
		return errors.Wrap(err, "failed to unpack log")
	}

	l.logger.WithFields(logan.F{
		"pair":    event.Raw.Address,
		"sender":  event.Sender.Hex(),
		"amount0": event.Amount0.String(),
		"amount1": event.Amount1.String(),
	}).Debug("pair mint")

	token0, token1, err := l.getTokens(ctx, event.Raw.Address)
	if err != nil {
		return errors.Wrap(err, "failed get tokens")
	}

	err = l.eventQueue.Send(ctx, channels.Event{
		Type: channels.ReservesUpdateEvent,
		ReservesUpdate: &channels.ReservesUpdate{
			Address:       event.Raw.Address,
			Token0:        token0,
			Token1:        token1,
			Reserve0Delta: event.Amount0,
			Reserve1Delta: event.Amount1,
		},
	})

	return errors.Wrap(err, "failed to add event to queue")
}

func (l *Listener) handleBurn(ctx context.Context, log *types.Log) error {
	var event uniswapv2pair.UniswapV2PairBurn

	err := l.eventUnpacker.Unpack(&event, BurnEvent, log.Data)
	if err != nil {
		return errors.Wrap(err, "failed to unpack log")
	}

	l.logger.WithFields(logan.F{
		"sender":  event.Sender.Hex(),
		"amount0": event.Amount0.String(),
		"amount1": event.Amount1.String(),
		"to":      event.To.Hex(),
	}).Debug("pair burn")

	token0, token1, err := l.getTokens(ctx, event.Raw.Address)
	if err != nil {
		return errors.Wrap(err, "failed get tokens")
	}

	err = l.eventQueue.Send(ctx, channels.Event{
		Type: channels.ReservesUpdateEvent,
		ReservesUpdate: &channels.ReservesUpdate{
			Address:       event.Raw.Address,
			Token0:        token0,
			Token1:        token1,
			Reserve0Delta: event.Amount0,
			Reserve1Delta: event.Amount1,
		},
	})

	return errors.Wrap(err, "failed to add event to queue")
}

func (l *Listener) handlePairCreation(ctx context.Context, log *types.Log) error {
	var event uniswapv2factory.UniswapV2FactoryPairCreated

	err := l.eventUnpacker.Unpack(&event, PairCreatedEvent, log.Data)
	if err != nil {
		return errors.Wrap(err, "failed to unpack log")
	}

	l.logger.WithFields(logan.F{
		"pair":   event.Pair,
		"token0": event.Token0,
		"token1": event.Token1,
	}).Debug("pair created")

	token0, token1, err := l.getTokens(ctx, event.Raw.Address)
	if err != nil {
		return errors.Wrap(err, "failed get tokens")
	}

	pair := l.uniswapV2.Pairs.Get(event.Raw.Address)
	reserve0, reserve1, err := pair.GetReserves(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get reserves", logan.F{
			"pair": pair.Address,
		})
	}

	err = l.eventQueue.Send(ctx, channels.Event{
		Type: channels.PairCreationEvent,
		PairCreation: &channels.PairCreation{
			Address:  event.Pair,
			Token0:   token0,
			Token1:   token1,
			Reserve0: reserve0,
			Reserve1: reserve1,
		},
	})

	return errors.Wrap(err, "failed to sent pair creation event")
}

func (l *Listener) getTokens(
	ctx context.Context, pairAddress common.Address,
) (token0, token1 common.Address, err error) {

	pair := l.uniswapV2.Pairs.Get(pairAddress)

	token0Contract, err := pair.Token0(ctx)
	if err != nil {
		return common.Address{}, common.Address{}, errors.Wrap(err,
			"failed to get token0",
			logan.F{
				"pair": pairAddress.Hex(),
			})
	}
	token1Contract, err := pair.Token1(ctx)
	if err != nil {
		return common.Address{}, common.Address{}, errors.Wrap(err,
			"failed to get token1",
			logan.F{
				"pair": pairAddress.Hex(),
			})
	}

	return token0Contract.Address(), token1Contract.Address(), nil
}
