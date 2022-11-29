package inmemory

import (
	"context"
	"math/big"
	"sync"

	"github.com/Velnbur/uniswapv2-indexer/internal/providers"
	"github.com/ethereum/go-ethereum/common"
)

var _ providers.UniswapV2PairProvider = &UniswapV2PairProvider{}

type uniswapV2PairValue struct {
	Token0, Token1       common.Address
	Reserves0, Reserves1 *big.Int
}

type UniswapV2PairProvider struct {
	cache sync.Map
}

func NewUniswapV2PairProvider() *UniswapV2PairProvider {
	return &UniswapV2PairProvider{}
}

func (p *UniswapV2PairProvider) GetReserves(
	_ context.Context, pair common.Address,
) (reserves0, reserves1 *big.Int, err error) {
	res, ok := p.cache.Load(pair)
	if !ok {
		return nil, nil, nil
	}

	pairValue := res.(uniswapV2PairValue)

	return pairValue.Reserves0, pairValue.Reserves1, nil
}

func (p *UniswapV2PairProvider) GetTokens(
	_ context.Context, pair common.Address,
) (token0, token1 common.Address, err error) {
	res, ok := p.cache.Load(pair)
	if !ok {
		return common.Address{}, common.Address{}, nil
	}

	pairValue := res.(uniswapV2PairValue)

	return pairValue.Token0, pairValue.Token1, nil
}

// SetReserves implements providers.UniswapV2PairProvider
func (p *UniswapV2PairProvider) SetReserves(
	_ context.Context, pair common.Address, reserve0, reserve1 *big.Int,
) error {
	res, ok := p.cache.Load(pair)
	if !ok {
		res = uniswapV2PairValue{}
	}
	pairValue := res.(uniswapV2PairValue)

	pairValue.Reserves0 = reserve0
	pairValue.Reserves1 = reserve1

	p.cache.Store(pair, pairValue)
	return nil
}

// SetTokens implements providers.UniswapV2PairProvider
func (p *UniswapV2PairProvider) SetTokens(
	_ context.Context, pair common.Address, token0, token1 common.Address,
) error {
	res, ok := p.cache.Load(pair)
	if !ok {
		res = uniswapV2PairValue{}
	}
	pairValue := res.(uniswapV2PairValue)

	pairValue.Token0 = token0
	pairValue.Token1 = token1

	p.cache.Store(pair, pairValue)
	return nil
}
