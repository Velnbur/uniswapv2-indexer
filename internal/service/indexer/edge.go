package indexer

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type EdgeKey struct {
	Token0, Token1 common.Address
}

type Edge struct {
	EdgeKey

	Reserve0, Reserve1 *big.Int
}

func NewEdge(token0, token1 common.Address, reserve0, reserve1 *big.Int) *Edge {
	return &Edge{
		EdgeKey: EdgeKey{
			Token0: token0,
			Token1: token1,
		},
		Reserve0: reserve0,
		Reserve1: reserve1,
	}
}
