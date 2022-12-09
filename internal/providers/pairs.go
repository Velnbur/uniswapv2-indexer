package providers

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
)

type UniswapV2PairProvider interface {
	GetTokens(ctx context.Context, pair common.Address) (common.Address, common.Address, error)
	SetTokens(ctx context.Context, pair, token0, token1 common.Address) error
}
