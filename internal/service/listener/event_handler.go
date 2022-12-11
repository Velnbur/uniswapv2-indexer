package listener

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
	"gitlab.com/distributed_lab/logan/v3"

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

	err = l.eventQueue.Send(ctx, channels.Event{
		Type: channels.ReservesUpdateEvent,
		ReservesUpdate: &channels.ReservesUpdate{
			Address: event.Raw.Address,
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

	err = l.eventQueue.Send(ctx, channels.Event{
		Type: channels.ReservesUpdateEvent,
		ReservesUpdate: &channels.ReservesUpdate{
			Address:       event.Raw.Address,
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

	err = l.eventQueue.Send(ctx, channels.Event{
		Type: channels.ReservesUpdateEvent,
		ReservesUpdate: &channels.ReservesUpdate{
			Address:       event.Raw.Address,
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

	err = l.eventQueue.Send(ctx, channels.Event{
		Type: channels.ReservesUpdateEvent,
		ReservesUpdate: &channels.ReservesUpdate{
			Address:       event.Raw.Address,
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

	err = l.eventQueue.Send(ctx, channels.Event{
		Type: channels.PairCreationEvent,
		PairCreation: &channels.PairCreation{
			Address: event.Pair,
			// FIXME:
			Reserve0: &big.Int{},
			Reserve1: &big.Int{},
		},
	})

	return errors.Wrap(err, "failed to sent pair creation event")
}
