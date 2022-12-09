package providers

import (
	"context"
	"sync"

	"github.com/ethereum/go-ethereum/common"
)

type UniswapV2FactoryInMemoryProvider struct {
	cache sync.Map
}

type uniswapV2FactoryInMemoryProviderkey struct {
	factory common.Address
	index   uint64
}

var _ UniswapV2FactoryProvider = &UniswapV2FactoryInMemoryProvider{}

func NewUniswapV2FactoryInMemoryProvider() *UniswapV2FactoryInMemoryProvider {
	return &UniswapV2FactoryInMemoryProvider{}
}

func (p *UniswapV2FactoryInMemoryProvider) GetPairByIndex(
	ctx context.Context, factory common.Address, index uint64,
) (common.Address, error) {
	key := uniswapV2FactoryInMemoryProviderkey{
		factory: factory,
		index:   index,
	}

	if pair, ok := p.cache.Load(key); ok {
		return pair.(common.Address), nil
	}

	return common.Address{}, nil
}

func (p *UniswapV2FactoryInMemoryProvider) SetPairByIndex(
	ctx context.Context, factory, pair common.Address, index uint64,
) error {
	key := factory.String() + ":" + string(index)
	p.cache.Store(key, pair)
	return nil
}
