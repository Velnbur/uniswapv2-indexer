package math

import (
	"math/big"
)

// SwapAmountOut - calculates amounts of tokens that you will get
// after swap
func SwapAmountOut(
	reserves0, reserves1 []*big.Int, amountIn *big.Int,
) *big.Int {
	n := len(reserves0)

	reserves0Prod := Product(reserves0...)
	reserves1Prod := Product(reserves1...)

	sum := big.NewInt(0)

	for i := 0; i < n-1; i++ {
		sum.Add(sum, new(big.Int).Mul(
			Product(reserves0[:i]...),
			Product(reserves0[i+1:n]...),
		))
	}

	sum.Mul(sum, amountIn)

	result := new(big.Int).Mul(reserves1Prod, amountIn)

	// result = result / (reserves0Prod + sum)
	result.Quo(
		result,
		new(big.Int).Add(reserves0Prod, sum),
	)

	return result
}

// Product - return Product of all big integers in array
func Product(nums ...*big.Int) *big.Int {
	result := big.NewInt(1)

	for _, num := range nums {
		result.Mul(result, num)
	}

	return result
}

func Sum(nums ...*big.Int) *big.Int {
	result := big.NewInt(0)

	for _, num := range nums {
		result.Add(result, num)
	}

	return result
}
