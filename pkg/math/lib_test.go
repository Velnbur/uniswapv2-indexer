package math

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func mustFromString(t *testing.T, s string) *big.Int {
	i, ok := new(big.Int).SetString(s, 10)
	require.True(t, ok)
	return i
}

func Test_SwapAmountsOut(t *testing.T) {

	t.Run("ETH -> USDT", func(t *testing.T) {
		ethAmount, ok := new(big.Int).SetString("1000000000000000000", 10)
		require.True(t, ok)

		reserves0 := []*big.Int{
			mustFromString(t, "11904476979297547639664"),
		}

		reserves1 := []*big.Int{
			mustFromString(t, "15161485837452"),
		}

		amountOut := SwapAmountOut(reserves0, reserves1, ethAmount)
		t.Log(amountOut)
	})

	t.Run("USDT -> ETH -> DAI", func(t *testing.T) {
		ethAmount, ok := new(big.Int).SetString("1000000", 10)
		require.True(t, ok)

		reserves0 := []*big.Int{
			mustFromString(t, "15161485837452"),
			mustFromString(t, "5165403989650444294732"),
		}

		reserves1 := []*big.Int{
			mustFromString(t, "11904476979297547639664"),
			mustFromString(t, "6587199298527047793486029"),
		}

		amountOut := SwapAmountOut(reserves0, reserves1, ethAmount)
		t.Log(amountOut)
	})

	t.Run("USDT -> ETH -> DAI -> USDT", func(t *testing.T) {
		reserves0 := []*big.Int{
			mustFromString(t, "15161485837452"),
			mustFromString(t, "5165403989650444294732"),
			mustFromString(t, "3405046456596189505100521"),
		}

		reserves1 := []*big.Int{
			mustFromString(t, "11904476979297547639664"),
			mustFromString(t, "6587199298527047793486029"),
			mustFromString(t, "3411362463968"),
		}

		amountIn := mustFromString(t, "1000000")

		amountOut := SwapAmountOut(reserves0, reserves1, amountIn)
		t.Log(amountOut)
	})
}
