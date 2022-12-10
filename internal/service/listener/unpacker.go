package listener

import (
	abiPkg "github.com/ethereum/go-ethereum/accounts/abi"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

// EventUnpacker - from ABI unpackes configured events
// into structs
type EventUnpacker struct {
	abi *abiPkg.ABI
}

func NewEventUnpacker(abi *abiPkg.ABI) *EventUnpacker {
	return &EventUnpacker{abi}
}

func (e *EventUnpacker) Unpack(dest interface{}, event EventKey, data []byte) error {
	err := e.abi.UnpackIntoInterface(dest, string(event), data)

	return errors.Wrap(err, "failed to unpack event", logan.F{
		"event": event,
	})
}
