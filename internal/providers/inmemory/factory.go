package inmemory

import (
	"context"
	"sync"

	"github.com/Velnbur/uniswapv2-indexer/internal/providers"
	"github.com/ethereum/go-ethereum/common"
)

var _ providers.UniswapV2FactoryProvider = &UniswapV2FactoryProvider{}

type uniswapV2Pairkey struct {
	factory common.Address
	index   uint64
}

type UniswapV2FactoryProvider struct {
	cache sync.Map
}

func NewUniswapV2FactoryProvider() *UniswapV2FactoryProvider {
	return &UniswapV2FactoryProvider{}
}

// GetPairByIndex implements providers.UniswapV2FactoryProvider
func (p *UniswapV2FactoryProvider) GetPairByIndex(
	ctx context.Context, factory common.Address, index uint64,
) (common.Address, error) {
	res, ok := p.cache.Load(uniswapV2Pairkey{factory: factory, index: index})
	if !ok {
		return common.Address{}, nil
	}

	return res.(common.Address), nil
}

// SetPairByIndex implements providers.UniswapV2FactoryProvider
func (p *UniswapV2FactoryProvider) SetPairByIndex(
	ctx context.Context, factory, pair common.Address, index uint64,
) error {
	p.cache.Store(uniswapV2Pairkey{factory: factory, index: index}, pair)
	return nil
}
