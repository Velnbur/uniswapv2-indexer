package listener

import (
	abiPkg "github.com/ethereum/go-ethereum/accounts/abi"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

// EventUnpacker - from ABI unpackes configured events
// into structs
type EventUnpacker struct {
	pair    *abiPkg.ABI
	factory *abiPkg.ABI
}

func NewEventUnpacker(pair, factory *abiPkg.ABI) *EventUnpacker {
	return &EventUnpacker{pair, factory}
}

func (e *EventUnpacker) Unpack(dest interface{}, event EventKey, data []byte) error {
	var err error

	switch event {
	case SwapEvent, SyncEvent, MintEvent, BurnEvent:
		err = e.pair.UnpackIntoInterface(dest, string(event), data)
	case PairCreatedEvent:
		err = e.factory.UnpackIntoInterface(dest, string(event), data)
	}

	return errors.Wrap(err, "failed to unpack event", logan.F{
		"event": event,
	})
}
