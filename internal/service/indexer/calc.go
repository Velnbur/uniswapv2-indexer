package indexer

import "math/big"

func CalculateSwapAmount(amountIn, reserve0, reserve1 *big.Int) *big.Int {
	// amountOut = amountIn * reserve1 / (reserve0 + amountIn)
	amountOut := new(big.Int).Mul(amountIn, reserve1)
	amountOut = amountOut.Div(amountOut, new(big.Int).Add(reserve0, amountIn))

	return amountOut
}

func CalculateSwapRatio(reserve0, reserve1 *big.Int) *big.Float {
	// ratio = reserve1 / reserve0
	ratio := new(big.Float).Quo(new(big.Float).SetInt(reserve1), new(big.Float).SetInt(reserve0))

	return ratio
}
