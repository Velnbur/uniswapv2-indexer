package channels

import "context"

type EventType int

const (
	PairCreationEvent EventType = iota + 1
	BlockCreationEvent
	ReservesUpdateEvent
)

type Event struct {
	Type EventType

	BlockCreation  *BlockCreation
	PairCreation   *PairCreation
	ReservesUpdate *ReservesUpdate
}

type EventQueue interface {
	Send(ctx context.Context, events ...Event) error
	Receive(ctx context.Context) (<-chan Event, error)
}
