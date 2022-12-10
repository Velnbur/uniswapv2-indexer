package channels

import (
	"context"
	"errors"

	"github.com/Velnbur/uniswapv2-indexer/pkg/helpers"
)

// HACK: check if SwapEventChan implements channels.SwapEventsQueue
var _ ReservesUpdateQueue = &ReservesUpdateChan{}

// ReservesUpdateChan - is a realization of in memory channels.SwapEventQueue
// that makes everything through Go's channels
type ReservesUpdateChan struct {
	subscribers []chan ReservesUpdate
}

func NewReservesUpdateInmemoryChan() *ReservesUpdateChan {
	return &ReservesUpdateChan{
		subscribers: make([]chan ReservesUpdate, 0),
	}
}

// TODO: may be, update it to more proper value
const DefaultReservesUpdateChanLen = 256

func (ch *ReservesUpdateChan) Receive(ctx context.Context) (<-chan ReservesUpdate, error) {
	if helpers.IsCanceled(ctx) {
		return nil, errors.New("context is canceled")
	}

	subscription := make(chan ReservesUpdate, DefaultReservesUpdateChanLen)

	ch.subscribers = append(ch.subscribers, subscription)

	return subscription, nil
}

func (ch *ReservesUpdateChan) Send(ctx context.Context, events ...ReservesUpdate) error {
	if helpers.IsCanceled(ctx) {
		return errors.New("context is canceled")
	}

	for _, subscriber := range ch.subscribers {
		ch.sendEventsToSubscriber(subscriber, events...)
	}

	return nil
}

func (ch *ReservesUpdateChan) sendEventsToSubscriber(
	subscriber chan<- ReservesUpdate, events ...ReservesUpdate,
) {
	for _, event := range events {
		subscriber <- event
	}
}
