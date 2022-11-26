package providers

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type UniswapV2PairProvider interface {
	GetTokens(ctx context.Context, pair common.Address) (common.Address, common.Address, error)
	SetTokens(ctx context.Context, pair, token0, token1 common.Address) error
	GetReserves(ctx context.Context, pair common.Address) (*big.Int, *big.Int, error)
	SetReserves(ctx context.Context, pair common.Address, reserve0, reserve1 *big.Int) error
}
