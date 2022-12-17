package providers

import (
	"context"

	"github.com/Velnbur/uniswapv2-indexer/internal/data"
	"github.com/ethereum/go-ethereum/common"
)

type PathesProvider interface {
	GetPathes(ctx context.Context, token0, token1 common.Address) ([]data.Path, error)
	SetPathes(ctx context.Context, token0, token1 common.Address, pathes []data.Path) error
}
