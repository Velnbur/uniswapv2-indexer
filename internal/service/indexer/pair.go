package indexer

import "math/big"

type Pair struct {
	Token1 string
	Token2 string

	Amount1 *big.Int
	Amount2 *big.Int
}
