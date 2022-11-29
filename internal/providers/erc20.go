package providers

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
)

type Erc20Provider interface {
	GetSymbol(ctx context.Context, address common.Address) (string, error)
	SetSymbol(ctx context.Context, address common.Address, symbol string) error
}
