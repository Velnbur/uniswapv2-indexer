package channels

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// ReservesUpdate - event of that reserves in UniswapV2 pair
// were updated. All fields represent the change (difference)
// between old and new state. All could be nil, but not all at
// the same time.
type ReservesUpdate struct {
	Address        common.Address
	Token0, Token1 common.Address
	Reserve0Delta  *big.Int
	Reserve1Delta  *big.Int
}
