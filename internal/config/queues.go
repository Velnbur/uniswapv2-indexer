package config

import (
	"github.com/Velnbur/uniswapv2-indexer/internal/channels"
	"gitlab.com/distributed_lab/kit/comfig"
)

type Queuer interface {
	EventsQueue() channels.EventQueue
}

type queuer struct {
	onceEvents comfig.Once
}

func (q *queuer) EventsQueue() channels.EventQueue {
	return q.onceEvents.Do(func() interface{} {
		return channels.NewEventChan()
	}).(channels.EventQueue)
}
