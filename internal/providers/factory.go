package providers

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
)

type UniswapV2FactoryProvider interface {
	GetPairByIndex(ctx context.Context, factory common.Address, index uint64) (common.Address, error)
	SetPairByIndex(ctx context.Context, factory, pair common.Address, index uint64) error
	GetPairByTokens(ctx context.Context, factory, token0, token1 common.Address) (common.Address, error)
	SetPairByTokens(ctx context.Context, factory, token0, token1, pair common.Address) error
}
