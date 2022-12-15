package channels

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// PairCreation - event of pair creation in factory
// contract
type PairCreation struct {
	Address            common.Address
	Token0, Token1     common.Address
	Reserve0, Reserve1 *big.Int
}
