package channels

import "context"

var _ PairCreationQueue = &PairCreationChan{}

type PairCreationChan struct {
	subscribers []chan<- PairCreation
}

// Receive implements PairCreationQueue
func (p *PairCreationChan) Receive(ctx context.Context) (<-chan PairCreation, error) {
	subscriber := make(chan PairCreation, DefaultPairCreationChannelLen)

	p.subscribers = append(p.subscribers, subscriber)

	return subscriber, nil
}

// Send implements PairCreationQueue
func (p *PairCreationChan) Send(ctx context.Context, events ...PairCreation) error {
	for _, event := range events {
		p.sendEventToSubscribers(event)
	}

	return nil
}

func (p *PairCreationChan) sendEventToSubscribers(event PairCreation) {
	for _, sub := range p.subscribers {
		sub <- event
	}
}

// TODO: may be, update it to more proper value
const DefaultPairCreationChannelLen = 256

func NewPairCreationChannel() *PairCreationChan {
	return &PairCreationChan{
		subscribers: make([]chan<- PairCreation, 0),
	}
}
