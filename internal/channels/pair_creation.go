package channels

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// PairCreation - event of pair creation in factory
// contract
type PairCreation struct {
	Address            common.Address
	Reserve0, Reserve1 *big.Int
}

type PairCreationQueue interface {
	Send(ctx context.Context, events ...PairCreation) error
	Receive(ctx context.Context) (<-chan PairCreation, error)
}
