package channels

import "context"

var _ EventQueue = &EventChan{}

type EventChan struct {
	subscribers []chan<- Event
}

func NewEventChan() *EventChan {
	return &EventChan{
		subscribers: make([]chan<- Event, 0, 16),
	}
}

const DefaultEventsChanLen = 256

func (e *EventChan) Receive(ctx context.Context) (<-chan Event, error) {
	subscriber := make(chan Event, DefaultEventsChanLen)

	e.subscribers = append(e.subscribers, subscriber)

	return subscriber, nil
}

// Send implements EventQueue
func (e *EventChan) Send(ctx context.Context, events ...Event) error {
	for _, event := range events {
		e.sendToSubscribers(event)
	}

	return nil
}

func (e *EventChan) sendToSubscribers(event Event) {
	for _, sub := range e.subscribers {
		sub <- event
	}
}
