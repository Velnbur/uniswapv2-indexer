package providers

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
)

type UniswapV2FactoryProvider interface {
	GetPairByIndex(ctx context.Context, factory common.Address, index uint64) (common.Address, error)
	SetPairByIndex(ctx context.Context, factory, pair common.Address, index uint64) error
}
