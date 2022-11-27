package inmemory

import (
	"context"

	"github.com/Velnbur/uniswapv2-indexer/internal/providers"
	"github.com/Velnbur/uniswapv2-indexer/pkg/helpers"
	"github.com/ethereum/go-ethereum/common"
)

var _ providers.UniswapV2FactoryProvider = &UniswapV2FactoryProvider{}

type uniswapV2Pairkey struct {
	factory common.Address
	index   uint64
}

type UniswapV2FactoryProvider struct {
	pairs map[uniswapV2Pairkey]common.Address
}

// GetPairByIndex implements providers.UniswapV2FactoryProvider
func (p *UniswapV2FactoryProvider) GetPairByIndex(
	ctx context.Context, factory common.Address, index uint64,
) (common.Address, error) {
	if helpers.IsCanceled(ctx) {
		return common.Address{}, ctx.Err()
	}

	pair, ok := p.pairs[uniswapV2Pairkey{factory: factory, index: index}]
	if !ok {
		return common.Address{}, nil
	}

	return pair, nil
}

// SetPairByIndex implements providers.UniswapV2FactoryProvider
func (p *UniswapV2FactoryProvider) SetPairByIndex(
	ctx context.Context, factory, pair common.Address, index uint64,
) error {
	if helpers.IsCanceled(ctx) {
		return ctx.Err()
	}

	p.pairs[uniswapV2Pairkey{factory: factory, index: index}] = pair

	return nil
}
