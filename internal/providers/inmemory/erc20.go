package inmemory

import (
	"context"
	"sync"

	"github.com/Velnbur/uniswapv2-indexer/internal/providers"
	"github.com/ethereum/go-ethereum/common"
)

type erc20Value struct {
	Decimals uint8
	Symbol   string
}

var _ providers.Erc20Provider = &Erc20Provider{}

type Erc20Provider struct {
	cache sync.Map
}

func NewErc20Provider() *Erc20Provider {
	return &Erc20Provider{}
}

func (e *Erc20Provider) GetSymbol(ctx context.Context, address common.Address) (string, error) {
	res, ok := e.cache.Load(address)
	if !ok {
		return "", nil
	}

	return res.(erc20Value).Symbol, nil
}

func (e *Erc20Provider) SetSymbol(ctx context.Context, address common.Address, symbol string) error {
	res, ok := e.cache.Load(address)
	if !ok {
		res = erc20Value{}
	}

	value := res.(erc20Value)
	value.Symbol = symbol

	e.cache.Store(address, value)
	return nil
}
