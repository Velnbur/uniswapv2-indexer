package providers

import (
	"context"
	"sync"

	"github.com/ethereum/go-ethereum/common"
)

var _ Erc20Provider = &Erc20InMemoryProvider{}

type Erc20InMemoryProvider struct {
	cache sync.Map
}

func NewErc20InMemoryProvider() *Erc20InMemoryProvider {
	return &Erc20InMemoryProvider{}
}

func (p *Erc20InMemoryProvider) GetSymbol(
	ctx context.Context, address common.Address,
) (string, error) {
	if v, ok := p.cache.Load(address); ok {
		return v.(string), nil
	}

	return "", nil
}

func (p *Erc20InMemoryProvider) SetSymbol(
	ctx context.Context, address common.Address, symbol string,
) error {
	p.cache.Store(address, symbol)
	return nil
}
