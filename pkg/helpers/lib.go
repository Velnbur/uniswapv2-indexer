package helpers

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
)

var ZeroAddress = common.Address{}

func IsCanceled(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

func IsAddressZero(address common.Address) bool {
	return address == ZeroAddress
}
